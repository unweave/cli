package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/unweave/cli/ui"
	"github.com/unweave/unweave/api/types"
)

func Code(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true
	ctx := cmd.Context()

	// TODO: parse args and forward to ssh command

	ui.Infof("Initializing node...")

	execch, errch, err := execCreateAndWatch(ctx, types.ExecConfig{}, types.GitConfig{})
	if err != nil {
		return err
	}

	for {
		select {
		case e := <-execch:
			if e.Status == types.StatusRunning {
				if e.Connection == nil {
					ui.Errorf("❌ Something unexpected happened. No connection info found for session %q", e.ID)
					ui.Infof("Run `unweave session ls` to see the status of your session and try connecting manually.")
					os.Exit(1)
				}
				ui.Infof("🚀 Session %q up and running", e.ID)
				ui.Infof("Setting up VS Code ...")

				if err := removeKnownHostsEntry(e.Connection.Host); err != nil {
					// Log and continue anyway. Most likely the entry is not there.
					ui.Debugf("Failed to remove known_hosts entry: %v", err)
				}
				if err := addHost("uw:"+e.ID, e.Connection.Host, e.Connection.User, e.Connection.Port); err != nil {
					ui.Debugf("Failed to add host to ssh config: %v", err)
				}

				arg := fmt.Sprintf("vscode-remote://ssh-remote+%s@%s/home/ubuntu", e.Connection.User, e.Connection.Host)
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
