package shell

import (
	"fmt"
	"os"
	"os/exec"

	"deploycli/internal/static"
)

// RunLocal executes a local command, streaming its own stdout/stderr through
// to this process's stderr (command chatter counts as progress, so it never
// pollutes the machine-parseable stdout channel).
func RunLocal(dir string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	cmd.Stdin = nil
	return cmd.Run()
}

func CheckBinary(name string) error {
	if _, err := exec.LookPath(name); err != nil {
		return fmt.Errorf("%s not found on PATH", name)
	}
	return nil
}

func CheckDockerDaemon() error {
	cmd := exec.Command(static.BinDocker, "info")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker daemon is not reachable")
	}
	return nil
}
