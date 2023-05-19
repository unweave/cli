package ssh

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/unweave/cli/ui"
	"github.com/unweave/unweave/api/types"
)

func Connect(ctx context.Context, connectionInfo types.ConnectionInfo, privKeyPath string) error {
	var sshArgs []string

	// TODO, we want to allow options to override the UserKnownHostsFile and StrictHostKeyChecking
	sshArgs = append(sshArgs, "-o", "UserKnownHostsFile=/dev/null")
	sshArgs = append(sshArgs, "-o", "StrictHostKeyChecking=no")
	if privKeyPath != "" {
		sshArgs = append(sshArgs, "-i", privKeyPath)
	}

	sshCommand := exec.Command(
		"ssh",
		append(sshArgs, fmt.Sprintf("%s@%s", connectionInfo.User, connectionInfo.Host))...,
	)

	ui.Debugf("Running SSH command: %s", strings.Join(sshCommand.Args, " "))

	sshCommand.Stdin = os.Stdin
	sshCommand.Stdout = os.Stdout
	sshCommand.Stderr = os.Stderr

	if err := sshCommand.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			// Exited with non-zero exit code
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				if status.ExitStatus() == 255 {
					ui.Infof("The remote host closed the connection.")
					return nil
				}
			}
			return err
		}
		return fmt.Errorf("SSH command failed: %v", err)
	}
	return nil
}
