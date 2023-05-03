package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/unweave/cli/config"
	"github.com/unweave/cli/ui"
	"github.com/unweave/unweave/api/types"
)

func Code(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true
	ctx := cmd.Context()

	errMsg := "❌ Invalid arguments. If you want to pass arguments to the ssh command, " +
		"use the -- flag. See `unweave ssh --help` for  more information."

	execCh := make(chan types.Exec)
	errCh := make(chan error)

	var err error
	var sshArgs []string
	var execRef string // Can be execID or name

	// If the number of args is great than one, we always expect the first arg to be
	// the separator flag "--". If the number of args is one, we expect it to be the
	// execID or name
	if len(args) > 1 {
		sshArgs = args[1:]
		if sshArgs[0] != "--" {
			ui.Errorf(errMsg)
			os.Exit(1)
		}

		if len(args) == 1 {
			execRef = args[0]
		} else {
			execRef = args[0]
		}
	}
	if config.CreateExec {

		// If the flag to create a new exec is passed, any arguments must be forwarded to
		// the ssh command
		if len(args) > 0 && args[0] != "--" {
			ui.Errorf(errMsg)
			os.Exit(1)
		}

		ui.Infof("Initializing node...")

		execCh, errCh, err = execCreateAndWatch(ctx, types.ExecConfig{}, types.GitConfig{})
		if err != nil {
			return err
		}

	} else {

		if execRef == "" {
			var execs []types.Exec

			execRef, execs, err = selectExec(cmd.Context(), "Select a session to connect to")
			if err != nil {
				return err
			}
			if len(execs) == 0 {
				ui.Errorf("❌ No active sessions found and no session name or ID provided. If " +
					"you want to create a new session, use the --new flag.")
				os.Exit(1)
			}
		}

		execCh, errCh, err = execWaitTillReady(ctx, execRef)
		if err != nil {
			return err
		}
	}
	for {
		select {
		case e := <-execCh:
			if e.Status == types.StatusRunning {

				if e.Connection == nil {
					ui.Errorf("❌ Something unexpected happened. No connection info found for session %q", e.ID)
					ui.Infof("Run `unweave session ls` to see the status of your session and try connecting manually.")
					os.Exit(1)
				}
				ui.Infof("🚀 Session %q up and running", e.ID)
				ui.Infof("🔧 Setting up VS Code ...")

				if err := removeKnownHostsEntry(e.Connection.Host); err != nil {
					// Log and continue anyway. Most likely the entry is not there.
					ui.Debugf("Failed to remove known_hosts entry: %v", err)
				}

				if err := addHost("uw:"+e.ID, e.Connection.Host, e.Connection.User, e.Connection.Port); err != nil {
					ui.Debugf("Failed to add host to ssh config: %v", err)
				}

				defer func() {
					if e := removeHost("uw:" + e.ID); e != nil {
						ui.Debugf("Failed to remove host from ssh config: %v", e)
					}
				}()

				// TODO we should wait until port is open

				arg := fmt.Sprintf("vscode-remote://ssh-remote+%s@%s/home/ubuntu", e.Connection.User, e.Connection.Host)

				codeCmd := exec.Command("code", "--folder-uri="+arg)
				codeCmd.Stdout = os.Stdout
				codeCmd.Stderr = os.Stderr

				if e := codeCmd.Run(); e != nil {
					ui.Errorf("Failed to start VS Code: %v", e)
					os.Exit(1)
				}
				ui.Successf("✅ VS Code is ready!")
				return nil
			}

		case err := <-errCh:
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
