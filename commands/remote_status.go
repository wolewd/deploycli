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

func RemoteStatus(args []string) {
	runCommand(args, commandSpec{
		name:          "remote-status",
		minPositional: 1,
		usage:         "deploycli remote-status [user]@[ip]:[deploy_path] [--port 22]",
		exitCode:      exitcode.RemoteErr,
		binaries:      []string{static.BinSSH},
		needsTarget:   true,
		needsProbe:    true,
		run: func(p parsedArgs, log *output.Logger, target *shell.Target) {
			start := time.Now()
			out, err := shell.RunRemoteCapture(target, fmt.Sprintf("cd %s && docker compose ps", target.Path))
			if err != nil {
				log.StepResult("remote-status", false, time.Since(start), nil)
				log.Error("failed to get container status: %v\n%s", err, out)
				os.Exit(exitcode.RemoteErr)
			}
			if p.json {
				log.StepResult("remote-status", true, time.Since(start), map[string]any{"output": out})
			} else {
				fmt.Print(out)
			}
		},
	})
}
