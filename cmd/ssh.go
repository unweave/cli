package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/unweave/cli/ui"
	"github.com/unweave/unweave/api/types"
)

func removeKnownHostsEntry(hostname string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %v", err)
	}
	knownHostsFile := fmt.Sprintf("%s/.ssh/known_hosts", home)
	removeHostKeyCmd := exec.Command("ssh-keygen", "-R", hostname, "-f", knownHostsFile)

	ui.Debugf("Removing host key from known_hosts: %s", strings.Join(removeHostKeyCmd.Args, " "))

	if err = removeHostKeyCmd.Run(); err != nil {
		return fmt.Errorf("failed to remove host key from known_hosts: %v", err)
	}
	return nil
}

func ssh(ctx context.Context, connectionInfo types.ConnectionInfo, prvKeyPath *string) error {
	sshCommand := exec.Command(
		"ssh",
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		fmt.Sprintf("%s@%s", connectionInfo.User, connectionInfo.Host),
	)
	if prvKeyPath != nil {
		sshCommand.Args = append(sshCommand.Args, "-i", *prvKeyPath)
	}

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

func SSH(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true
	ctx := cmd.Context()

	// TODO: parse args and forward to ssh command

	ui.Infof("Initializing node...")

	execch, errch, err := sessionCreateAndWatch(ctx, types.ExecCtx{})
	if err != nil {
		return err
	}

	for {
		select {
		case e := <-execch:
			if e.Status == types.StatusRunning {
				if e.Status == types.StatusRunning {
					if e.Connection == nil {
						ui.Errorf("❌ Something unexpected happened. No connection info found for session %q", e.ID)
						ui.Infof("Run `unweave session ls` to see the status of your session and try connecting manually.")
						os.Exit(1)
					}
					ui.Infof("🚀 Session %q up and running", e.ID)

					if err := removeKnownHostsEntry(e.Connection.Host); err != nil {
						// Log and continue anyway. Most likely the entry is not there.
						ui.Debugf("Failed to remove known_hosts entry: %v", err)
					}

					if err := ssh(ctx, *e.Connection, nil); err != nil {
						ui.Errorf("%s", err)
						os.Exit(1)
					}

					if terminate := ui.Confirm("SSH session terminated. Do you want to terminate the session?", "n"); terminate {
						if err := sessionTerminate(ctx, e.ID); err != nil {
							return err
						}
						ui.Infof("Session %q terminated.", e.ID)
						os.Exit(0)
					}
				}
				return nil
			}

		case err := <-errch:
			var e *types.Error
			if errors.As(err, &e) {
				uie := &ui.Error{Error: e}
				fmt.Println(uie.Verbose())
				os.Exit(1)
			}
			return err

		case <-ctx.Done():
			return nil
		}
	}
}
