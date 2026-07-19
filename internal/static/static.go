// Package static centralises every hardcoded string (binary names, SSH
// flags, default values, external URLs) so they are defined once.
package static

const (
	BinDocker = "docker"
	BinGit    = "git"
	BinSSH    = "ssh"
	BinSCP    = "scp"
)

const (
	SSHOptBatchMode             = "BatchMode=yes"
	SSHOptStrictHostKeyChecking = "StrictHostKeyChecking=accept-new"
)

const DefaultSSHPort = "22"

// Version is set at build time via -ldflags, e.g.
// go build -ldflags "-X deploycli/internal/static.Version=v1.2.3"
// Falls back to "dev" when built without a version stamp.
var Version = "dev"

const (
	URLDockerInstall = "https://docs.docker.com/engine/install/"
	URLGitInstall    = "https://git-scm.com/"
)
