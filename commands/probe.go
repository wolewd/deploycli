package commands

import (
	"os"

	"deploycli/internal/output"
	"deploycli/internal/shell"
	"deploycli/internal/static"
)

// fatalProbeError maps a ProbeConnection error to an actionable message
// and exits with the given code. It never returns.
func fatalProbeError(log *output.Logger, target *shell.Target, err error, code int) {
	switch err {
	case shell.ErrPubkeyRejected:
		log.Error("Cannot access %s using the RSA key. Register the key with:\n  ssh-copy-id %s", target.UserHost(), target.UserHost())
	case shell.ErrConnectionFailed:
		log.Error("Cannot connect to %s. Check the host, port, and network/firewall settings.", target.UserHost())
	case shell.ErrDockerGroupMissing:
		log.Error("User %s on the remote host is not in the docker group. Fix it with:\n  sudo usermod -aG docker %s && newgrp docker", target.User, target.User)
	case shell.ErrDockerNotInstalled:
		log.Error("Docker is not installed on the remote host %s. Install it first: "+static.URLDockerInstall, target.Host)
	default:
		log.Error("Failed to reach %s: %v", target.UserHost(), err)
	}
	os.Exit(code)
}
