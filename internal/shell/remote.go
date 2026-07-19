package shell

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"deploycli/internal/static"
)

// Target is a parsed "user@ip:path" remote.
type Target struct {
	User string
	Host string
	Path string
	Port string
}

// ParseTarget parses "user@ip:path" into its components.
func ParseTarget(raw string, port string) (*Target, error) {
	atIdx := strings.Index(raw, "@")
	colonIdx := strings.Index(raw, ":")
	if atIdx < 0 || colonIdx < 0 || colonIdx < atIdx {
		return nil, fmt.Errorf("invalid target %q, expected format user@ip:path", raw)
	}
	user := raw[:atIdx]
	host := raw[atIdx+1 : colonIdx]
	path := raw[colonIdx+1:]
	if user == "" || host == "" || path == "" {
		return nil, fmt.Errorf("invalid target %q, expected format user@ip:path", raw)
	}
	if port == "" {
		port = static.DefaultSSHPort
	}
	return &Target{User: user, Host: host, Path: path, Port: port}, nil
}

func (t *Target) UserHost() string {
	return fmt.Sprintf("%s@%s", t.User, t.Host)
}

// sshOpts returns the shared -o flags for non-interactive SSH/SCP.
func sshOpts() []string {
	return []string{
		"-o", static.SSHOptBatchMode,
		"-o", static.SSHOptStrictHostKeyChecking,
	}
}

// sshBaseArgs builds a full SSH argument list from sshOpts plus the port.
func sshBaseArgs(port string) []string {
	return append(sshOpts(), "-p", port)
}

// RunRemote runs a single command over SSH, streaming output to stderr.
func RunRemote(t *Target, remoteCmd string) error {
	args := append(sshBaseArgs(t.Port), t.UserHost(), remoteCmd)
	cmd := exec.Command(static.BinSSH, args...)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	cmd.Stdin = nil
	return cmd.Run()
}

// RunRemoteCapture is like RunRemote but returns captured combined output.
func RunRemoteCapture(t *Target, remoteCmd string) (string, error) {
	args := append(sshBaseArgs(t.Port), t.UserHost(), remoteCmd)
	cmd := exec.Command(static.BinSSH, args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func RunRemoteSession(t *Target, cmds []string) error {
	joined := strings.Join(cmds, " && ")
	return RunRemote(t, joined)
}

// Sentinel errors categorise what went wrong with a remote target so
// callers can print specific, actionable messages.
var (
	ErrPubkeyRejected     = errors.New("ssh key not accepted by remote host")
	ErrConnectionFailed   = errors.New("could not connect to remote host")
	ErrDockerGroupMissing = errors.New("remote user cannot access docker without sudo")
	ErrDockerNotInstalled = errors.New("docker is not installed on remote host")
)

// ProbeConnection verifies reachability, key acceptance, and docker access.
// BatchMode=yes is critical: it makes SSH fail immediately on a rejected
// key instead of hanging on a password prompt when called from a subprocess.
func ProbeConnection(t *Target) error {
	out, err := RunRemoteCapture(t, "docker ps")
	if err == nil {
		return nil
	}
	lower := strings.ToLower(out)
	switch {
	case strings.Contains(lower, "permission denied") && strings.Contains(lower, "publickey"):
		return ErrPubkeyRejected
	case strings.Contains(lower, "connection refused"),
		strings.Contains(lower, "connection timed out"),
		strings.Contains(lower, "no route to host"),
		strings.Contains(lower, "operation timed out"),
		strings.Contains(lower, "name or service not known"):
		return ErrConnectionFailed
	case strings.Contains(lower, "command not found") && strings.Contains(lower, "docker"):
		return ErrDockerNotInstalled
	default:
		// SSH connected fine but `docker ps` failed for some other reason;
		// the most common cause is the remote user not being in the docker
		// group (permission denied on the docker socket).
		return ErrDockerGroupMissing
	}
}

func ScpUpload(t *Target, localPath, remoteDir string) error {
	dest := fmt.Sprintf("%s:%s/", t.UserHost(), strings.TrimSuffix(remoteDir, "/"))
	args := append(sshOpts(), "-P", t.Port, localPath, dest)
	cmd := exec.Command(static.BinSCP, args...)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
