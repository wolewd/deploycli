// Package exitcode defines the process exit codes deploycli uses so a
// calling REST API can tell what category of failure happened without
// parsing text output.
package exitcode

const (
	// OK means the command completed successfully.
	OK = 0
	// GeneralErr covers usage errors / anything not otherwise categorized.
	GeneralErr = 1
	// GitErr covers fetch/pull failures.
	GitErr = 10
	// DockerErr covers local docker build failures.
	DockerErr = 20
	// TransferErr covers docker save / scp transfer failures.
	TransferErr = 30
	// RemoteErr covers ssh / remote docker compose failures.
	RemoteErr = 40
)
