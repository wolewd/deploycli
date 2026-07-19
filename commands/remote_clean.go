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

// RemoteClean deletes leftover bundle.tar from the deploy path on the VPS.
func RemoteClean(args []string) {
	runCommand(args, commandSpec{
		name:          "remote-clean",
		minPositional: 1,
		usage:         "deploycli remote-clean [user]@[ip]:[deploy_path] [--port 22] [--json]",
		exitCode:      exitcode.RemoteErr,
		binaries:      []string{static.BinSSH},
		needsTarget:   true,
		needsProbe:    true,
		run: func(p parsedArgs, log *output.Logger, target *shell.Target) {
			cmd := fmt.Sprintf("cd %s && rm -f bundle.tar", target.Path)
			log.Step("Cleaning deploy path at %s", target.UserHost())
			start := time.Now()
			if err := shell.RunRemote(target, cmd); err != nil {
				log.StepResult("remote-clean", false, time.Since(start), nil)
				log.Error("remote clean failed: %v", err)
				os.Exit(exitcode.RemoteErr)
			}
			log.StepResult("remote-clean", true, time.Since(start), nil)
		},
	})
}
