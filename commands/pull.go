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
	p := parseArgs(args)
	log := output.NewLogger(p.json)

	if len(p.positional) < 2 {
		log.Error("usage: deploycli pull [project_path] [branch_name]")
		os.Exit(exitcode.GitErr)
	}
	projectPath := p.positional[0]
	branch := p.positional[1]

	if err := shell.CheckBinary(static.BinGit); err != nil {
		log.Error("git not found. Install and configure it first: " + static.URLGitInstall)
		os.Exit(exitcode.GitErr)
	}

	log.Step("Pulling %s from origin in %s", branch, projectPath)
	start := time.Now()
	err := shell.RunLocal("", static.BinGit, "-C", projectPath, "pull", "origin", branch)
	if err != nil {
		log.StepResult("pull", false, time.Since(start), nil)
		log.Error("git pull failed: %v", err)
		os.Exit(exitcode.GitErr)
	}
	log.StepResult("pull", true, time.Since(start), nil)
}
