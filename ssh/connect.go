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

func Connect(ctx context.Context, connectionInfo types.ConnectionInfo, prvKeyPath string, sshArgs []string) error {
	overrideUserKnownHostsFile := false
	overrideStrictHostKeyChecking := false

	for _, arg := range sshArgs {
		if strings.Contains(arg, "UserKnownHostsFile") {
			overrideUserKnownHostsFile = true
		}
		if strings.Contains(arg, "StrictHostKeyChecking") {
			overrideStrictHostKeyChecking = true
		}
	}

	if prvKeyPath != "" {
		sshArgs = append(sshArgs, "-i", prvKeyPath)
	}

	if !overrideUserKnownHostsFile {
		sshArgs = append(sshArgs, "-o", "UserKnownHostsFile=/dev/null")
	}
	if !overrideStrictHostKeyChecking {
		sshArgs = append(sshArgs, "-o", "StrictHostKeyChecking=no")
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
