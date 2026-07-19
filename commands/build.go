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

// Build runs a single-image docker build. Each call builds exactly one
// image; multi-service orchestration is the caller's responsibility.
func Build(args []string) {
	runCommand(args, commandSpec{
		name:          "build",
		minPositional: 3,
		usage:         "deploycli build [project_path] [image_name] [image_tag]",
		exitCode:      exitcode.DockerErr,
		binaries:      []string{static.BinDocker},
		checkDaemon:   true,
		run: func(p parsedArgs, log *output.Logger, _ *shell.Target) {
			projectPath := p.positional[0]
			imageName := p.positional[1]
			imageTag := p.positional[2]
			image := fmt.Sprintf("%s:%s", imageName, imageTag)
			log.Step("Building %s from %s (no cache)", image, projectPath)
			start := time.Now()
			err := shell.RunLocal("", static.BinDocker, "build", "--no-cache", "--rm", "-t", image, projectPath)
			if err != nil {
				log.StepResult("build", false, time.Since(start), map[string]any{"image": image})
				log.Error("docker build failed: %v", err)
				os.Exit(exitcode.DockerErr)
			}
			log.StepResult("build", true, time.Since(start), map[string]any{"image": image})
		},
	})
}
