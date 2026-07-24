package commands

import (
	"os"
	"time"

	"deploycli/internal/exitcode"
	"deploycli/internal/output"
	"deploycli/internal/shell"
	"deploycli/internal/static"
)

func PruneBuilder(args []string) {
	runCommand(args, commandSpec{
		name:          "prune-builder",
		minPositional: 0,
		usage:         "deploycli prune-builder",
		exitCode:      exitcode.DockerErr,
		binaries:      []string{static.BinDocker},
		run: func(p parsedArgs, log *output.Logger, _ *shell.Target) {
			log.Step("Pruning Docker build cache")
			start := time.Now()
			if err := shell.RunLocal("", static.BinDocker, "builder", "prune", "-af"); err != nil {
				log.StepResult("prune-builder", false, time.Since(start), nil)
				log.Error("docker builder prune failed: %v", err)
				os.Exit(exitcode.DockerErr)
			}
			log.StepResult("prune-builder", true, time.Since(start), nil)
		},
	})
}
