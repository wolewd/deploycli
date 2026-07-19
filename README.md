# deploycli

## Overview

deploycli builds Docker images locally and ships them to a VPS. The idea is simple, your VPS never sees source code or build toolchains. Images get built on your machine (or a dedicated build box), then uploaded. Production stays clean. 

It's meant to run as a subprocess behind a backend, like `ffmpeg`. Progress on stderr, results on stdout, exit codes for decisions. No text parsing. This program Only works for Linux btw.

## How to build

```
go build -o deploycli .
```

Stamp a version for releases:

```
go build -ldflags "-X deploycli/internal/static.Version=v1.0.0" -o deploycli .
```

Needs Go 1.22+. No dependencies outside the standard library.

## Requirements

**On your machine:**

| Tool | Why |
|---|---|
| `git` | pull source |
| `docker` (daemon running) | build and save images |
| `ssh`, `scp` | connect to the VPS, ship bundles |

**On the VPS:**

| Tool | Why |
|---|---|
| `sshd` | accept connections |
| `docker` + `docker compose` v2 | load images, run containers |

deploycli assumes the VPS already has a working docker-compose setup (or swarm, or whatever orchestrator you use). It does not provision anything. `remote-rebuild` runs `docker compose down`, loads new images, and runs `docker compose up -d`. The `docker-compose.yml` is yours to maintain.

**Before you start:**

- Your SSH key must be in the VPS's `authorized_keys`. The CLI never prompts for passwords.
- The SSH user on the VPS needs to be in the `docker` group. The CLI never runs `sudo` because a sudo prompt would hang a non-interactive subprocess.
- It uses `~/.ssh/id_rsa`. There's no config file and no `--key` flag.

Each command checks only the tools it actually needs. `fetch` checks for `git`, `send` checks for `docker` + `scp` + `ssh`, and so on.

## Commands

### fetch

```
deploycli fetch [project_path]
```

Runs `git fetch origin` in the given project directory. Use this before `pull` if you want to see what's new without merging.

### pull

```
deploycli pull [project_path] [branch_name]
```

Pulls the named branch from origin. Typically the first step in a deploy pipeline.

### build

```
deploycli build [project_path] [image_name] [image_tag]
```

Builds a single Docker image (no cache, cleans up intermediate containers). One image per call. If you have three services, call it three times. The caller owns the orchestration.

### send

```
deploycli send [user]@[ip]:[deploy_path] --image [image]:[tag] [--image ...] [--port 22] [--json]
```

Saves one or more images into a single tar archive and uploads it to the VPS via SCP. The bundle is atomic. If the connection drops, the VPS never sees a half-baked set of images.

### remote-rebuild

```
deploycli remote-rebuild [user]@[ip]:[deploy_path] --image [image]:[tag] [--image ...] [--port 22] [--json]
```

Over SSH: stops the stack, removes old images, loads the new bundle, and brings everything back up. All steps run in a single session so they share the same working directory.

### remote-status

```
deploycli remote-status [user]@[ip]:[deploy_path] [--port 22] [--json]
```

Runs `docker compose ps` on the VPS and prints the output. Useful as a sanity check after a deploy.

### remote-clean

```
deploycli remote-clean [user]@[ip]:[deploy_path] [--port 22] [--json]
```

Deletes leftover `bundle.tar` from the deploy path. `remote-rebuild` removes it automatically on success, but if a rebuild fails (or never runs), the file sits there. Use this to clean up.

### Flags

| Flag | What it does |
|---|---|
| `--image` | Pass an `image:tag` pair. Repeat for multiple images. Only `send` and `remote-rebuild` use it. |
| `--port` | SSH port. Defaults to 22. |
| `--json` | Each step writes one JSON line to stdout instead of human-readable text to stderr. |

## Exit codes

| Code | Meaning |
|---|---|
| 0 | success |
| 1 | usage error or something unexpected |
| 10 | git failed (fetch or pull) |
| 20 | docker build failed |
| 30 | transfer failed (docker save or scp) |
| 40 | remote failed (ssh, docker compose, cleanup) |

A calling backend can branch on these numbers without parsing any text.

## Example

Full deploy of a three-service docker-compose project:

```sh
# 1. Pull master
deploycli pull ~/projects/myapp master

# 2. Build each service
deploycli build ~/projects/myapp/api      api      latest
deploycli build ~/projects/myapp/worker   worker   latest
deploycli build ~/projects/myapp/scheduler scheduler latest

# 3. Bundle and ship to the VPS
deploycli send root@1.2.3.4:/opt/deploy/myapp \
  --image api:latest --image worker:latest --image scheduler:latest

# 4. Rebuild the remote stack
deploycli remote-rebuild root@1.2.3.4:/opt/deploy/myapp \
  --image api:latest --image worker:latest --image scheduler:latest

# 5. Check everything came up
deploycli remote-status root@1.2.3.4:/opt/deploy/myapp

# 6. If remote-rebuild fails, clean up the leftover bundle before retrying
deploycli remote-clean root@1.2.3.4:/opt/deploy/myapp
```

## For developers

### How it works

Each command is a standalone function that parses its own args, checks its own dependencies, does the work, and exits with a code. The idea is that a backend calls it via `os/exec` and knows the outcome from the exit code alone. Zero string matching.

Arguments are parsed with a hand-rolled loop instead of `flag` because flags like `--image` can appear after positional args (`deploycli send user@ip:/path --image foo:latest`), and `flag` stops at the first non-flag token.

### Design choices

**Explicit, not config-driven.** No config file, no env vars, no dotfiles. Every argument is on the command line. Predictable and easy to reason about from the caller side.

**Non interactive only.** SSH uses `BatchMode=yes` so a rejected key fails immediately instead of hanging. `StrictHostKeyChecking=accept-new` means first-time connections don't prompt. No `sudo` anywhere. The remote user touches the docker socket directly.

**Atomic transfer.** `send` saves all images into one tar. If SCP drops mid-transfer, the VPS has nothing new. The old stack keeps running until `remote-rebuild` finishes cleanly.

**One image per build, many per send.** `build` does one image because the caller might build them in parallel. `send` and `remote-rebuild` handle multiple images because the atomic-bundle guarantee requires it.

**Per-invocation temp dirs.** Uses `os.MkdirTemp`, not a fixed directory. Parallel deploys never collide.

**Centralised constants.** `internal/static` holds every hardcoded string (binary names, SSH flags, default port, install URLs). Nothing is duplicated across the codebase.

### Building a service on top

deploycli is designed to be embedded. Import the packages you need and wire them into your own `main.go`. The commands are plain exported functions:

```go
package main

import "deploycli/commands"

func main() {
    // your own flag parsing, config loading, auth, etc.
    // then just call the commands directly:
    commands.Pull([]string{"/home/deploy/projects/myapp", "main"})
    commands.Build([]string{"/home/deploy/projects/myapp/api", "api", "latest"})
    commands.Send([]string{"root@1.2.3.4:/opt/myapp", "--image", "api:latest"})
    commands.RemoteRebuild([]string{"root@1.2.3.4:/opt/myapp", "--image", "api:latest"})
    commands.RemoteClean([]string{"root@1.2.3.4:/opt/myapp"})
}
```

You get progress on stderr, structured results on stdout, and categorised exit codes for free. The `commands` package is the public API. Everything in `internal/` is implementation detail.

Common patterns:

- **HTTP handler that triggers a deploy.** Accept a webhook, validate it, call `commands.Pull` + `commands.Build` + `commands.Send` + `commands.RemoteRebuild` in sequence. Stream stderr to the HTTP response.
- **Parallel builds.** Call `commands.Build` in separate goroutines since each invocation builds exactly one image and they don't share state.
- **Custom logging.** The `output` package is internal, but you can wrap `os.Stderr` before calling commands to capture progress into your own log pipeline.

The CLI itself (`main.go`) is just a thin dispatcher that maps `os.Args` to command functions. You can replace it entirely with your own entry point that adds flags, config, or whatever your service needs.

### Contributing

Open a pull request. Keep it focused, don't add external dependencies unless there's a strong reason, and make sure `go build ./... && go vet ./...` passes.
