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
	execRef := ""
	if len(args) == 1 {
		execRef = args[0]
	}

	prvKey := config.SSHPrivateKeyPath
	execCh, isNew, errCh := getOrCreateExec(cmd, execRef)
	ctx := cmd.Context()

	for {
		select {
		case e := <-execCh:
			if e.Status == types.StatusRunning {
				prvKey, err := getDefaultKey(ctx, e, prvKey)
				if prvKey == "" {
					ui.Errorf("Expected private key to be none empty string")
					os.Exit(1)
				}
				if err != nil {
					ui.Errorf("Failed to get private key: %s", err)
					os.Exit(1)
				}

				ensureHosts(e, prvKey)
				err = handleCopySourceDir(isNew, e, prvKey)
				if err != nil {
					ui.HandleError(err)
					os.Exit(1)
				}

				ui.Infof("🔧 Setting up VS Code ...")
				arg := fmt.Sprintf("vscode-remote://ssh-remote+%s@%s%s", e.Network.User, e.Network.Host, config.ProjectHostDir())

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
