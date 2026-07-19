package commands

import (
	"os"
	"time"

	"deploycli/internal/exitcode"
	"deploycli/internal/output"
	"deploycli/internal/shell"
	"deploycli/internal/static"
)

func Fetch(args []string) {
	runCommand(args, commandSpec{
		name:          "fetch",
		minPositional: 1,
		usage:         "deploycli fetch [project_path]",
		exitCode:      exitcode.GitErr,
		binaries:      []string{"git"},
		run: func(p parsedArgs, log *output.Logger, _ *shell.Target) {
			projectPath := p.positional[0]
			log.Step("Fetching all branches from origin in %s", projectPath)
			start := time.Now()
			err := shell.RunLocal("", static.BinGit, "-C", projectPath, "fetch", "origin")
			if err != nil {
				log.StepResult("fetch", false, time.Since(start), nil)
				log.Error("git fetch failed: %v", err)
				os.Exit(exitcode.GitErr)
			}
			log.StepResult("fetch", true, time.Since(start), nil)
		},
	})
}
