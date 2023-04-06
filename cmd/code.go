package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/spf13/cobra"
	"github.com/unweave/cli/config"
	"github.com/unweave/cli/ui"
	"github.com/unweave/unweave/api/types"
)

func Code(cmd *cobra.Command, args []string) error {
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

	ticketCount := 0
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
						ui.Infof("Setting up VS Code ...")

						arg := fmt.Sprintf("vscode-remote://ssh-remote+%s@%s/home/ubuntu", s.Connection.User, s.Connection.Host)
						codeCmd := exec.Command("code", "--folder-uri="+arg)
						codeCmd.Stdout = os.Stdout
						codeCmd.Stderr = os.Stderr

						if e := codeCmd.Run(); e != nil {
							ui.Errorf("Failed to start VS Code: %v", e)
							os.Exit(1)
						}
						ui.Successf("VS Code is ready!")
						return nil
					}

					if s.Status == types.StatusError {
						ui.Errorf("❌ Session %s failed to start", sessionID)
						os.Exit(1)
					}
					if s.Status == types.StatusTerminated {
						ui.Errorf("Session %q is terminated.", sessionID)
						os.Exit(1)
					}

					if ticketCount%10 == 0 {
						ui.Infof("Waiting for session %q to start...", sessionID)
					}
					ticketCount++
				}
			}

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
