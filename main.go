// Command deploycli automates the PC-to-VPS Docker deploy flow.
//
// It is designed as a subprocess for a REST API backend (like ffmpeg):
// progress on stderr, machine-parseable results on stdout, categorised exit
// codes so the caller can distinguish failure types without parsing text.
package main

import (
	"fmt"
	"os"

	"deploycli/commands"
	"deploycli/internal/exitcode"
	"deploycli/internal/static"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(exitcode.GeneralErr)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "fetch":
		commands.Fetch(args)
	case "pull":
		commands.Pull(args)
	case "build":
		commands.Build(args)
	case "send":
		commands.Send(args)
	case "remote-rebuild":
		commands.RemoteRebuild(args)
	case "remote-status":
		commands.RemoteStatus(args)
	case "remote-clean":
		commands.RemoteClean(args)
	case "prune-builder":
		commands.PruneBuilder(args)
	case "-h", "--help", "help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "ERROR: unknown command %q\n\n", cmd)
		printUsage()
		os.Exit(exitcode.GeneralErr)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "deploycli %s\n\n", static.Version)
	fmt.Fprintln(os.Stderr, `Usage:
  deploycli fetch [project_path]
  deploycli pull [project_path] [branch_name]
  deploycli build [project_path] [image_name] [image_tag]
  deploycli send [user]@[ip]:[deploy_path] --image [image_name]:[image_tag] [--image ...] [--port 22] [--json]
  deploycli remote-rebuild [user]@[ip]:[deploy_path] --image [image_name]:[image_tag] [--image ...] [--port 22] [--json]
  deploycli remote-status [user]@[ip]:[deploy_path] [--port 22] [--json]
  deploycli remote-clean [user]@[ip]:[deploy_path] [--port 22] [--json]
  deploycli prune-builder

Notes:
  - SSH auth uses ~/.ssh/id_rsa (the default key), no config file is read.
  - --image may be repeated for send/remote-rebuild (multi-image docker-compose stacks).
  - --port defaults to 22.
  - --json makes each step emit a single-line JSON object on stdout instead of human-readable text.`)
}
