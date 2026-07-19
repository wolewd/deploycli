package commands

import (
	"os"
	"path/filepath"
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
	p := parseArgs(args)
	log := output.NewLogger(p.json)

	if len(p.positional) < 1 {
		log.Error("usage: deploycli send [user]@[ip]:[deploy_path] --image [image_name]:[image_tag] [--image ...] [--port 22]")
		os.Exit(exitcode.TransferErr)
	}
	if len(p.images) == 0 {
		log.Error("at least one --image is required, e.g. --image myapp:latest")
		os.Exit(exitcode.TransferErr)
	}

	target, err := shell.ParseTarget(p.positional[0], p.port)
	if err != nil {
		log.Error("%v", err)
		os.Exit(exitcode.TransferErr)
	}

	if err := shell.CheckBinary(static.BinDocker); err != nil {
		log.Error("docker not found. Install it first: " + static.URLDockerInstall)
		os.Exit(exitcode.TransferErr)
	}
	if err := shell.CheckBinary(static.BinSCP); err != nil {
		log.Error("scp not found. Install the openssh-client package first.")
		os.Exit(exitcode.TransferErr)
	}
	if err := shell.CheckBinary(static.BinSSH); err != nil {
		log.Error("ssh not found. Install the openssh-client package first.")
		os.Exit(exitcode.TransferErr)
	}

	log.Step("Probing connection to %s", target.UserHost())
	if err := shell.ProbeConnection(target); err != nil {
		reportProbeError(log, target, err, exitcode.TransferErr)
		return
	}

	// Use a unique temp dir per invocation so parallel requests never collide.
	tempDir, err := os.MkdirTemp("", "deploycli-send-*")
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
		log.StepResult("docker-save", false, time.Since(start), nil)
		log.Error("docker save failed: %v", err)
		os.Exit(exitcode.TransferErr)
	}
	log.StepResult("docker-save", true, time.Since(start), map[string]any{"images": p.images})

	log.Step("Uploading bundle to %s:%s", target.UserHost(), target.Path)
	start = time.Now()
	if err := shell.ScpUpload(target, bundlePath, target.Path); err != nil {
		log.StepResult("scp-upload", false, time.Since(start), nil)
		log.Error("scp upload failed: %v", err)
		os.Exit(exitcode.TransferErr)
	}
	log.StepResult("scp-upload", true, time.Since(start), nil)
}
