package cmd

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"github.com/unweave/cli/config"
	"github.com/unweave/cli/ui"
	"github.com/unweave/unweave/api/types"
)

var getSessionIDRegex = regexp.MustCompile(`^sess:([^/]+)`)

func Copy(cmd *cobra.Command, args []string) error {
	scpArgs := make([]string, 0, len(args))
	var targetExec *types.Exec

	for _, arg := range args {
		argExecID, err := getExecIDFromCopyArgs(arg)
		if err != nil {
			return err
		}

		if argExecID != "" && targetExec != nil {
			return fmt.Errorf("Copying between multiple sessions is not supported")
		}

		if argExecID != "" {
			targetExec, err = getExecByNameOrID(cmd.Context(), argExecID)
			if err != nil {
				return fmt.Errorf("Could not find session by name or ID")
			}
		}

		formattedArg, err := formatCopyArgToScpArgs(cmd.Context(), argExecID, targetExec, arg)
		if err != nil {
			return err
		}

		scpArgs = append(scpArgs, formattedArg)
	}

	if targetExec.SSHKey.PublicKey == nil && config.SSHPublicKeyPath == "" {
		return fmt.Errorf("Failed to identify public key, check your Unweave config file or specify it manually")
	}

	publicKeyPath := ""
	if config.SSHPrivateKeyPath != "" {
		publicKeyPath = config.SSHPublicKeyPath
	} else {
		publicKeyPath = *targetExec.SSHKey.PublicKey
	}

	ui.Infof(fmt.Sprintf("🔄 Copying %s to %s", scpArgs[0], scpArgs[1]))
	err := copySourceSCP(scpArgs, publicKeyPath)
	ui.Infof("✅  Copied %s to %s", scpArgs[0], scpArgs[1])
	return err
}

func formatCopyArgToScpArgs(ctx context.Context, argExecID string, exec *types.Exec, arg string) (string, error) {
	if argExecID == "" {
		return arg, nil
	}

	if exec == nil {
		return "", fmt.Errorf("Assertion failed, please file an issue with the Unweave team." +
			"Please provide steps to reproduce")
	}

	connectionInfo := exec.Connection
	if connectionInfo == nil {
		return "", fmt.Errorf("Could not get connection from session")
	}

	parts := strings.SplitN(arg, "/", 2)
	dir := filepath.Clean(parts[1])

	return fmt.Sprintf("%s@%s:/%s", connectionInfo.User, connectionInfo.Host, dir), nil
}

func getExecIDFromCopyArgs(input string) (string, error) {
	matches := getSessionIDRegex.FindStringSubmatch(input)

	if len(matches) == 2 {
		return matches[1], nil
	} else if input == "sess:" {
		return "", fmt.Errorf("ExecId not specified")
	} else {
		return "", nil
	}
}
