package commands

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"deploycli/internal/exitcode"
	"deploycli/internal/output"
	"deploycli/internal/shell"
	"deploycli/internal/static"
)

// Send bundles local Docker images into a single tar archive and uploads it
// to the VPS via SCP. The atomic bundle means a dropped connection never
// leaves the VPS with a partial set of images.
func Send(args []string) {
	runCommand(args, commandSpec{
		name:          "send",
		minPositional: 1,
		usage:         "deploycli send [user]@[ip]:[deploy_path] --image [image_name]:[image_tag] [--image ...] [--port 22]",
		exitCode:      exitcode.TransferErr,
		binaries:      []string{static.BinDocker, static.BinSCP, static.BinSSH},
		checkDaemon:   true,
		needsTarget:   true,
		needsProbe:    true,
		needsImages:   true,
		run: func(p parsedArgs, log *output.Logger, target *shell.Target) {
			tempDir, err := os.MkdirTemp(".", "deploycli-send-*")
			if err != nil {
				log.Error("failed to create local temp directory: %v", err)
				os.Exit(exitcode.TransferErr)
			}
			defer os.RemoveAll(tempDir)

			bundlePath := filepath.Join(tempDir, "bundle.tar")
			log.Step("Saving %d image(s) into %s", len(p.images), bundlePath)
			start := time.Now()
			saveArgs := append([]string{"save", "-o", bundlePath}, p.images...)
			if err := shell.RunLocal("", static.BinDocker, saveArgs...); err != nil {
				os.RemoveAll(tempDir)
				log.StepResult("docker-save", false, time.Since(start), nil)
				log.Error("docker save failed: %v", err)
				os.Exit(exitcode.TransferErr)
			}
			log.StepResult("docker-save", true, time.Since(start), map[string]any{"images": p.images})

			// Compute local hash before upload so we can verify after.
			localHash, err := fileSHA256(bundlePath)
			if err != nil {
				os.RemoveAll(tempDir)
				log.Error("failed to hash bundle: %v", err)
				os.Exit(exitcode.TransferErr)
			}

			log.Step("Uploading bundle to %s:%s", target.UserHost(), target.Path)
			start = time.Now()
			if err := shell.ScpUpload(target, bundlePath, target.Path); err != nil {
				os.RemoveAll(tempDir)
				log.StepResult("scp-upload", false, time.Since(start), nil)
				log.Error("scp upload failed: %v", err)
				os.Exit(exitcode.TransferErr)
			}
			log.StepResult("scp-upload", true, time.Since(start), nil)

			remoteBundle := target.Path + "/bundle.tar"
			out, err := shell.RunRemoteCapture(target, "sha256sum "+remoteBundle)
			if err != nil {
				os.RemoveAll(tempDir)
				log.Error("failed to verify bundle on remote: %v", err)
				os.Exit(exitcode.TransferErr)
			}
			remoteHash := strings.Fields(out)[0]
			if remoteHash != localHash {
				os.RemoveAll(tempDir)
				shell.RunRemote(target, "rm -f "+remoteBundle) // best-effort cleanup
				log.Error("bundle integrity check failed. local=%s remote=%s", localHash, remoteHash)
				os.Exit(exitcode.TransferErr)
			}
		},
	})
	}

	func fileSHA256(path string) (string, error) {
		f, err := os.Open(path)
		if err != nil {
			return "", err
		}
		defer f.Close()
		h := sha256.New()
		if _, err := io.Copy(h, f); err != nil {
			return "", err
		}
		return fmt.Sprintf("%x", h.Sum(nil)), nil
	}
