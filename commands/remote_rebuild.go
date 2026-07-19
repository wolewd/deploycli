package commands

import (
	"fmt"
	"os"
	"strings"
	"time"

	"deploycli/internal/exitcode"
	"deploycli/internal/output"
	"deploycli/internal/shell"
	"deploycli/internal/static"
)

// RemoteRebuild re-deploys images on the VPS. All steps run in a single
// SSH session so they share one working directory and one connection.
func RemoteRebuild(args []string) {
	p := parseArgs(args)
	log := output.NewLogger(p.json)

	if len(p.positional) < 1 {
		log.Error("usage: deploycli remote-rebuild [user]@[ip]:[deploy_path] --image [image_name]:[image_tag] [--image ...] [--port 22]")
		os.Exit(exitcode.RemoteErr)
	}
	if len(p.images) == 0 {
		log.Error("at least one --image is required, matching what was used with 'send', e.g. --image myapp:latest")
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

	// Remove old images one by one. "|| true" makes each removal tolerant of
	// an image not existing yet (e.g. the very first deploy of a service),
	// without masking real failures in the other steps.
	rmSteps := make([]string, 0, len(p.images))
	for _, img := range p.images {
		rmSteps = append(rmSteps, fmt.Sprintf("(docker rmi %s || true)", img))
	}

	steps := []string{
		fmt.Sprintf("cd %s", target.Path),
		"docker compose down",
		strings.Join(rmSteps, " && "),
		"docker load -i bundle.tar",
		"rm -f bundle.tar",
		"docker compose up -d",
	}

	log.Step("Rebuilding containers at %s with %d image(s)", target.UserHost(), len(p.images))
	start := time.Now()
	if err := shell.RunRemoteSession(target, steps); err != nil {
		log.StepResult("remote-rebuild", false, time.Since(start), map[string]any{"images": p.images})
		log.Error("remote rebuild failed: %v", err)
		os.Exit(exitcode.RemoteErr)
	}
	log.StepResult("remote-rebuild", true, time.Since(start), map[string]any{"images": p.images})
}
