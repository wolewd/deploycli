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
	p := parseArgs(args)
	log := output.NewLogger(p.json)

	if len(p.positional) < 1 {
		log.Error("usage: deploycli fetch [project_path]")
		os.Exit(exitcode.GitErr)
	}
	projectPath := p.positional[0]

	if err := shell.CheckBinary(static.BinGit); err != nil {
		log.Error("git not found. Install and configure it first: " + static.URLGitInstall)
		os.Exit(exitcode.GitErr)
	}

	log.Step("Fetching all branches from origin in %s", projectPath)
	start := time.Now()
	err := shell.RunLocal("", static.BinGit, "-C", projectPath, "fetch", "origin")
	if err != nil {
		log.StepResult("fetch", false, time.Since(start), nil)
		log.Error("git fetch failed: %v", err)
		os.Exit(exitcode.GitErr)
	}
	log.StepResult("fetch", true, time.Since(start), nil)
}
