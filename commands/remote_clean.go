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
	p := parseArgs(args)
	log := output.NewLogger(p.json)

	if len(p.positional) < 1 {
		log.Error("usage: deploycli remote-clean [user]@[ip]:[deploy_path] [--port 22] [--json]")
		os.Exit(exitcode.RemoteErr)
	}

	target, err := shell.ParseTarget(p.positional[0], p.port)
	if err != nil {
		log.Error("%v", err)
		os.Exit(exitcode.RemoteErr)
	}

	if err := shell.CheckBinary(static.BinSSH); err != nil {
		log.Error("ssh not found. Install the openssh-client package first.")
		os.Exit(exitcode.RemoteErr)
	}

	log.Step("Probing connection to %s", target.UserHost())
	if err := shell.ProbeConnection(target); err != nil {
		reportProbeError(log, target, err, exitcode.RemoteErr)
		return
	}

	cmd := fmt.Sprintf("cd %s && rm -f bundle.tar", target.Path)

	log.Step("Cleaning deploy path at %s", target.UserHost())
	start := time.Now()
	if err := shell.RunRemote(target, cmd); err != nil {
		log.StepResult("remote-clean", false, time.Since(start), nil)
		log.Error("remote clean failed: %v", err)
		os.Exit(exitcode.RemoteErr)
	}
	log.StepResult("remote-clean", true, time.Since(start), nil)
}
