package commands

import (
	"os"
	"time"

	"deploycli/internal/exitcode"
	"deploycli/internal/output"
	"deploycli/internal/shell"
	"deploycli/internal/static"
)

func Pull(args []string) {
	runCommand(args, commandSpec{
		name:          "pull",
		minPositional: 2,
		usage:         "deploycli pull [project_path] [branch_name]",
		exitCode:      exitcode.GitErr,
		binaries:      []string{static.BinGit},
		run: func(p parsedArgs, log *output.Logger, _ *shell.Target) {
			projectPath := p.positional[0]
			branch := p.positional[1]
			log.Step("Pulling %s from origin in %s", branch, projectPath)
			start := time.Now()
			err := shell.RunLocal("", static.BinGit, "-C", projectPath, "pull", "origin", branch)
			if err != nil {
				log.StepResult("pull", false, time.Since(start), nil)
				log.Error("git pull failed: %v", err)
				os.Exit(exitcode.GitErr)
			}
			log.StepResult("pull", true, time.Since(start), nil)
		},
	})
}
