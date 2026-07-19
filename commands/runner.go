package commands

import (
	"fmt"
	"os"

	"deploycli/internal/exitcode"
	"deploycli/internal/output"
	"deploycli/internal/shell"
	"deploycli/internal/static"
)

type commandSpec struct {
	name          string
	minPositional int
	usage         string
	exitCode      int
	binaries      []string
	checkDaemon   bool
	needsTarget   bool
	needsProbe    bool
	needsImages   bool
	run           func(p parsedArgs, log *output.Logger, target *shell.Target)
}

func runCommand(args []string, spec commandSpec) {
	p, err := parseArgs(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(exitcode.GeneralErr)
	}
	log := output.NewLogger(p.json)

	if len(p.positional) < spec.minPositional {
		log.Error("usage: %s", spec.usage)
		os.Exit(spec.exitCode)
	}
	if spec.needsImages && len(p.images) == 0 {
		log.Error("at least one --image is required, e.g. --image myapp:latest")
		os.Exit(spec.exitCode)
	}

	for _, bin := range spec.binaries {
		if err := shell.CheckBinary(bin); err != nil {
			switch bin {
			case "docker":
				log.Error("docker not found. Install it first: " + static.URLDockerInstall)
			case "git":
				log.Error("git not found. Install and configure it first: " + static.URLGitInstall)
			case "scp":
				log.Error("scp not found. Install the openssh-client package first.")
			case "ssh":
				log.Error("ssh not found. Install the openssh-client package first.")
			default:
				log.Error("%s not found on PATH", bin)
			}
			os.Exit(spec.exitCode)
		}
	}
	if spec.checkDaemon {
		if err := shell.CheckDockerDaemon(); err != nil {
			log.Error("Docker daemon is not running locally. Start Docker and try again.")
			os.Exit(spec.exitCode)
		}
	}

	var target *shell.Target
	if spec.needsTarget {
		t, err := shell.ParseTarget(p.positional[0], p.port)
		if err != nil {
			log.Error("%v", err)
			os.Exit(spec.exitCode)
		}
		target = t
	}
	if spec.needsProbe {
		log.Step("Probing connection to %s", target.UserHost())
		if err := shell.ProbeConnection(target); err != nil {
			fatalProbeError(log, target, err, spec.exitCode)
			return
		}
	}

	spec.run(p, log, target)
}
