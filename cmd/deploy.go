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
	"github.com/unweave/unweave/api/types"
)

func Deploy(cmd *cobra.Command, args []string) error {
	name := strings.ReplaceAll(config.EndpointName, "_", "-")

	return runSSHConnectionCommand(cmd, args, &deployCommandFlow{endpointName: name})
}

type deployCommandFlow struct {
	endpointName string
	execCommandFlow
}

func (d *deployCommandFlow) parseArgs(cmd *cobra.Command, args []string) execCmdArgs {
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

	execArgs := execCmdArgs{
		sshConnectionOptions: config.SSHConnectionOptions,
		skipCopy:             len(args) == 0,
	}

	if !execArgs.skipCopy {
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

		execArgs.copyDir = path
	}

	execArgs.userCommand = strings.Split(config.Command, " ")

	execArgs.execCommand = execArgs.userCommand
	execArgs.execCommand = wrapCommandNoHupLogging(execArgs.userCommand)

	return execArgs
}

func (d *deployCommandFlow) onSshCommandFinish(ctx context.Context, execID string) error {
	uwc := config.InitUnweaveClient()
	owner, project := config.GetProjectOwnerAndName()

	endpoints, err := uwc.Endpoints.List(ctx, owner, project)
	if err != nil {
		return fmt.Errorf("list endpoints: %w", err)
	}

	end, ok := findEndpoint(d.endpointName, endpoints)

	if !ok {
		ui.Debugf("endpoint not found, creating new, name: %q", d.endpointName)

		endpoint, err := uwc.Endpoints.Create(ctx, owner, project, execID, d.endpointName)
		if err != nil {
			return fmt.Errorf("deploy: %w", err)
		}

		ui.Infof("✅ Created endpint %q for exec", endpoint.ID)
		ui.Infof("🔗 Access endpoint at:")
		ui.Infof("    https://%s", endpoint.HTTPEndpoint)

		return nil
	}

	ui.Debugf("endpoint found, creating version, name: %q, id: %q", d.endpointName, end.ID)

	version, err := uwc.Endpoints.CreateVersion(ctx, owner, project, end.ID, execID)
	if err != nil {
		return fmt.Errorf("create version: %w", err)
	}

	ui.Infof("✅ Created endpint version %q for exec", version.ID)
	ui.Infof("🔗 Access endpoint at:")
	ui.Infof("    https://%s", end.HTTPEndpoint)
	ui.Infof("🔗 Access version at:")
	ui.Infof("    https://%s", version.HTTPEndpoint)

	return nil
}

func findEndpoint(name string, ends []types.Endpoint) (types.Endpoint, bool) {
	for _, end := range ends {
		if strings.EqualFold(name, end.Name) || strings.EqualFold(name, end.ID) {
			return end, true
		}
	}

	return types.Endpoint{}, false
}
