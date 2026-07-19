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
	p := parseArgs(args)
	log := output.NewLogger(p.json)

	if len(p.positional) < 1 {
		log.Error("usage: deploycli remote-status [user]@[ip]:[deploy_path] [--port 22]")
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
}
