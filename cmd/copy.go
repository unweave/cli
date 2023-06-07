package cmd

import (
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

var (
	getSessionIDRegex = regexp.MustCompile(`^sess:([^/]+)`)
	regExValidDirpath = regexp.MustCompile(`^\/(?:[^\/]+\/)*[^\/]+$`)
)

func Copy(cmd *cobra.Command, args []string) error {
	exec, err := getTargetExec(cmd, args)
	if err != nil {
		return err
	}

	if exec == nil {
		return fmt.Errorf("At least one remote host must be specified")
	}
	if exec.Connection == nil {
		return fmt.Errorf("Target session must have an active connection")
	}
	if exec.SSHKey.PublicKey == nil && config.SSHPublicKeyPath == "" {
		return fmt.Errorf("Failed to identify public key, check your Unweave config file or specify it manually")
	}

	scpArgs, err := formatPaths(exec, args)
	if err != nil {
		ui.Infof("❌ Unsuccessful copy: %s", err.Error())
	}
	ui.Infof(fmt.Sprintf("🔄 Copying %s => %s", scpArgs[0], scpArgs[1]))

	privateKey, err := getDefaultKey(cmd.Context(), *exec, config.SSHPrivateKeyPath)
	if err != nil {
		ui.Infof("❌ Unsuccessful copy, failed to get private key")
	}

	switch {
	case shouldCopyLocalDirToRemote(args[0], args[1]):
		// Eventually simplify this to talk in terms of scpArgs, too many dependants for now
		err = copyDirFromLocalAndUnzip(exec.ID, scpArgs[0], splitSessFromDirpath(args[1]), *exec.Connection, privateKey)
	case shouldCopyRemoteDirToLocal(args[0], args[1]):
		err = copyDirFromRemoteAndUnzip(scpArgs[0], scpArgs[1], privateKey)
	default:
		err = copySourceSCP(scpArgs[0], scpArgs[1], privateKey)
	}

	if err != nil {
		fmt.Println(err.Error())
		ui.Infof("❌ Unsuccessful copy %s => %s", scpArgs[0], scpArgs[1])
		return nil
	}

	ui.Infof("✅  Copied %s => %s", scpArgs[0], scpArgs[1])
	return nil
}

func shouldCopyLocalDirToRemote(localPath1, remotePath2 string) bool {
	if strings.Contains(localPath1, "sess:") {
		return false
	}

	pathInfo, err := os.Stat(localPath1)
	if err != nil || pathInfo == nil {
		return false
	}

	return pathInfo.IsDir()
}

func shouldCopyRemoteDirToLocal(remotePath1, localPath2 string) bool {
	if !strings.Contains(remotePath1, "sess:") {
		return false
	}

	sessFromDirpath := splitSessFromDirpath(remotePath1)
	return regExValidDirpath.MatchString(sessFromDirpath)
}

func getTargetExec(cmd *cobra.Command, args []string) (*types.Exec, error) {
	var targetExec *types.Exec
	for _, arg := range args {
		argExecID, err := getExecIDFromCopyArgs(arg)
		if err != nil {
			return nil, err
		}

		if argExecID != "" && targetExec != nil {
			return nil, fmt.Errorf("Copying between multiple sessions is not supported")
		}

		if argExecID != "" {
			targetExec, err = getExecByNameOrID(cmd.Context(), argExecID)
			if err != nil {
				return nil, fmt.Errorf("Could not find session by name or ID")
			}
		}
	}
	return targetExec, nil
}

func formatLocalPath(arg string) (string, error) {
	if arg == "." {
		arg, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("Could not get local directory")
		}
		return arg, nil
	}
	return arg, nil
}

func formatRemotePath(exec *types.Exec, arg string) (string, error) {
	if exec == nil {
		return "", fmt.Errorf("Assertion failed, please file an issue with the Unweave team. Please provide steps to reproduce")
	}

	connectionInfo := exec.Connection
	if connectionInfo == nil {
		return "", fmt.Errorf("Could not get connection from session")
	}

	return fmt.Sprintf("%s@%s:%s", connectionInfo.User, connectionInfo.Host, splitSessFromDirpath(arg)), nil
}

func formatPaths(exec *types.Exec, args []string) ([]string, error) {
	formattedArgs := make([]string, len(args))

	for i, arg := range args {
		if strings.HasPrefix(arg, "sess:") {
			formattedArg, err := formatRemotePath(exec, arg)
			if err != nil {
				return nil, err
			}
			formattedArgs[i] = formattedArg
		} else {
			formattedArg, err := formatLocalPath(arg)
			if err != nil {
				return nil, err
			}
			formattedArgs[i] = formattedArg
		}
	}

	return formattedArgs, nil
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
