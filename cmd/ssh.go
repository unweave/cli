package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/unweave/cli/config"
	"github.com/unweave/cli/ui"
	"github.com/unweave/unweave/api/types"
)

func ssh(ctx context.Context, connectionInfo types.ConnectionInfo) error {
	sshCommand := exec.Command(
		"ssh",
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		fmt.Sprintf("%s@%s", connectionInfo.User, connectionInfo.Host),
	)

	sshCommand.Stdin = os.Stdin
	sshCommand.Stdout = os.Stdout
	sshCommand.Stderr = os.Stderr

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	go func() {
		<-signalChan
		fmt.Println("\nReceived an interrupt, do you really want to exit? (y/n)")
		var response string
		fmt.Scanf("%s", &response)
		response = strings.ToLower(response)
		if response == "y" || response == "yes" {
			os.Exit(0)
		}
	}()

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

	sessionID, err := sessionCreate(cmd.Context(), types.ExecCtx{})
	if err != nil {
		os.Exit(1)
		return nil
	}

	uwc := InitUnweaveClient()
	listTerminated := config.All
	owner, projectName := config.GetProjectOwnerAndName()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			sessions, err := uwc.Session.List(ctx, owner, projectName, listTerminated)
			if err != nil {
				var e *types.Error
				if errors.As(err, &e) {
					uie := &ui.Error{Error: e}
					fmt.Println(uie.Verbose())
					os.Exit(1)
				}
				return err
			}

			for _, s := range sessions {
				if s.ID == sessionID {
					if s.Status == types.StatusRunning {
						if s.Connection == nil {
							ui.Errorf("❌ Something unexpected happened. No connection info found for session %q", sessionID)
							ui.Infof("Run `unweave session ls` to see the status of your session and try connecting manually.")
							os.Exit(1)
						}
						ui.Infof("🚀 Session %q up and running", sessionID)
						if err := ssh(ctx, *s.Connection); err != nil {
							ui.Errorf("%s", err)
							os.Exit(1)
						}

						// Ask to terminate session
						if terminate := ui.Confirm("SSH session terminated. Do you want to terminate the session?", "n"); terminate {
							if err := sessionTerminate(ctx, sessionID); err != nil {
								return err
							}
							ui.Infof("Session %q terminated.", sessionID)
							os.Exit(0)
						}

					}
					if s.Status == types.StatusError {
						ui.Errorf("❌ Session %s failed to start", sessionID)
						os.Exit(1)
					}
					if s.Status == types.StatusTerminated {
						ui.Errorf("Session %q is terminated.", sessionID)
						os.Exit(1)
					}
				}
			}

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
