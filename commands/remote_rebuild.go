package commands

import (
	"fmt"
	"os"
	"time"

	"deploycli/internal/exitcode"
	"deploycli/internal/output"
	"deploycli/internal/shell"
	"deploycli/internal/static"
)

// RemoteRebuild re-deploys images on the VPS. All steps run in a single
// SSH session so they share one working directory and one connection.
func RemoteRebuild(args []string) {
	runCommand(args, commandSpec{
		name:          "remote-rebuild",
		minPositional: 1,
		usage:         "deploycli remote-rebuild [user]@[ip]:[deploy_path] --image [image_name]:[image_tag] [--image ...] [--port 22]",
		exitCode:      exitcode.RemoteErr,
		binaries:      []string{static.BinSSH},
		needsTarget:   true,
		needsProbe:    true,
		needsImages:   true,
		run: func(p parsedArgs, log *output.Logger, target *shell.Target) {
			steps := []string{
				fmt.Sprintf("cd %s", target.Path),
				"docker compose down",
				"docker load -i bundle.tar",
				"rm -f bundle.tar",
				"docker compose up -d",
				"docker image prune -f",
			}

			log.Step("Rebuilding containers at %s with %d image(s)", target.UserHost(), len(p.images))
			start := time.Now()
			if err := shell.RunRemoteSession(target, steps); err != nil {
				log.StepResult("remote-rebuild", false, time.Since(start), map[string]any{"images": p.images})
				log.Error("remote rebuild failed: %v", err)
				os.Exit(exitcode.RemoteErr)
			}
			log.StepResult("remote-rebuild", true, time.Since(start), map[string]any{"images": p.images})
		},
	})
}
