package cmd

import (
	"context"
	"fmt"
	"os"
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
			ui.HandleError(err)
			os.Exit(1)
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
			ui.HandleError(err)
			os.Exit(1)
		}

		scpArgs = append(scpArgs, formattedArg)
	}
	if targetExec == nil {
		return fmt.Errorf("At least one remote host must be specified")
	}
	if targetExec.Connection == nil {
		return fmt.Errorf("Target session must have an active connection")
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

	ui.Infof(fmt.Sprintf("🔄 Copying %s => %s", scpArgs[0], scpArgs[1]))

	// If the first argument of the SCP args points towards a logical directory assume local => remote
	pathInfo, _ := os.Stat(scpArgs[0])
	if pathInfo.IsDir() {
		return copySourceAndUnzip(targetExec.ID, scpArgs[0], splitSessFromDirpath(args[1]), *targetExec.Connection, publicKeyPath)
	} else {
		err := copySourceSCP(scpArgs[0], scpArgs[1], publicKeyPath)
		if err != nil {
			ui.Infof("❎ Unsuccessful copy %s => %s", scpArgs[0], scpArgs[1])
			return nil
		}
	}

	ui.Infof("✅  Copied %s => %s", scpArgs[0], scpArgs[1])
	return nil
}

func formatCopyArgToScpArgs(ctx context.Context, argExecID string, exec *types.Exec, arg string) (string, error) {
	if argExecID == "" {
		if arg == "." {
			return os.Getwd()
		}

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

	return fmt.Sprintf("%s@%s:%s", connectionInfo.User, connectionInfo.Host, splitSessFromDirpath(arg)), nil
}

// splitSessFromDirpath takes a qualified cp argument e.g. sess:<execId>/path and returns /path
func splitSessFromDirpath(arg string) string {
	parts := strings.SplitN(arg, "/", 2)
	return fmt.Sprintf("/%s", filepath.Clean(parts[1]))
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
