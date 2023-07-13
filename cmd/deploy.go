package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/unweave/cli/config"
	"github.com/unweave/cli/ui"
)

func Deploy(cmd *cobra.Command, args []string) error {
	return runSSHConnectionCommand(cmd, args, &deployCommandFlow{})
}

type deployCommandFlow struct {
	execCommandFlow
}

func (d *deployCommandFlow) parseArgs(cmd *cobra.Command, args []string) execCmdArgs {
	if len(args) == 0 {
		const errMsg = "❌ Invalid arguments. You must pass a file path to deploy.\n" +
			"See `unweave deploy --help` for more information"
		ui.Errorf(errMsg)
		os.Exit(1)
	}

	if len(args) > 1 {
		const errMsg = "❌ Invalid arguments. You can only pass a single file path to deploy.\n" +
			"See `unweave deploy --help` for more information"
		ui.Errorf(errMsg)
		os.Exit(1)
	}

	if config.Command == "" {
		const errMsg = "❌ Invalid arguments. You must pass the --cmd flag to deploy.\n" +
			"See `unweave deploy --help` for more information"
		ui.Errorf(errMsg)
		os.Exit(1)
	}

	p := args[0]

	path, err := filepath.Abs(args[0])
	if err != nil {
		const errMsg = "❌ Invalid arguments. Problem with filepath: %q.\n%s\n" +
			"See `unweave deploy --help` for more information"
		ui.Errorf(errMsg, p, err)
		os.Exit(1)
	}

	_, err = os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			const errMsg = "❌ Invalid arguments. Filepath does not exist: %q.\n" +
				"See `unweave deploy --help` for more information"
			ui.Errorf(errMsg, path)
			os.Exit(1)

		}

		const errMsg = "❌ Invalid arguments. Problem with filepath: %q.\n%s\n" +
			"See `unweave deploy --help` for more information"
		ui.Errorf(errMsg, path, err)
		os.Exit(1)
	}

	execArgs := execCmdArgs{
		sshConnectionOptions: config.SSHConnectionOptions,
		copyDir:              path,
	}

	execArgs.userCommand = strings.Split(config.Command, " ")

	execArgs.execCommand = execArgs.userCommand
	execArgs.execCommand = wrapCommandNoHupLogging(execArgs.userCommand)

	return execArgs
}

func (d *deployCommandFlow) onSshCommandFinish(ctx context.Context, execID string) error {
	uwc := config.InitUnweaveClient()

	owner, project := config.GetProjectOwnerAndName()

	endpoint, err := uwc.Endpoints.Create(ctx, owner, project, execID)
	if err != nil {
		return fmt.Errorf("deploy: %w", err)
	}

	ui.Infof("✅ Created endpint %q for exec", endpoint.ID)
	ui.Infof("🔗 Access endpoint at:")
	ui.Infof("  https://%s", endpoint.HTTPEndpoint)

	return nil
}
