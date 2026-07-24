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

// Send streams Docker images directly to the VPS via a compressed SSH pipe
// and loads them into the remote Docker daemon. No temp files, no SCP.
func Send(args []string) {
	runCommand(args, commandSpec{
		name:          "send",
		minPositional: 1,
		usage:         "deploycli send [user]@[ip]:[deploy_path] --image [image_name]:[image_tag] [--image ...] [--port 22]",
		exitCode:      exitcode.TransferErr,
		binaries:      []string{static.BinDocker, static.BinSSH},
		needsTarget:   true,
		needsProbe:    true,
		needsImages:   true,
		run: func(p parsedArgs, log *output.Logger, target *shell.Target) {
			log.Step("Sending %d image(s) to %s", len(p.images), target.UserHost())
			start := time.Now()
			if err := shell.PipeImages(target, p.images); err != nil {
				log.StepResult("send", false, time.Since(start), map[string]any{"images": p.images})
				log.Error("send failed: %v", err)
				os.Exit(exitcode.TransferErr)
			}
			log.StepResult("send", true, time.Since(start), map[string]any{"images": p.images})

			for _, img := range p.images {
				out, err := shell.RunRemoteCapture(target, fmt.Sprintf("docker image inspect %s", img))
				if err != nil {
					log.Error("image %s did not load on remote: %v\n%s", img, err, out)
					os.Exit(exitcode.TransferErr)
				}
			}
		},
	})
}
