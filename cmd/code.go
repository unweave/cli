package cmd

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/unweave/cli/config"
	"github.com/unweave/cli/ui"
	"github.com/unweave/unweave/api/types"
	"os"
	"os/exec"
)

func Code(cmd *cobra.Command, args []string) error {
	execRef, _ := parseArgsToExecRefAndSSHArgs(args)
	prvKey := config.SSHPrivateKeyPath
	execCh, isNew, errCh := getOrCreateExec(cmd, execRef)
	ctx := cmd.Context()

	for {
		select {
		case e := <-execCh:
			if e.Status == types.StatusRunning {
				ensureHosts(e)
				defer cleanupHosts(e)
				prvKey := getUnweavePrivateKeyOrDefault(ctx, e, prvKey)

				err := handleCopySourceDir(isNew, e, prvKey)
				if err != nil {
					return err
				}

				ui.Infof("🔧 Setting up VS Code ...")
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
