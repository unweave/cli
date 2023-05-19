package cmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/unweave/cli/config"
	"github.com/unweave/cli/session"
	"github.com/unweave/cli/ssh"
	"github.com/unweave/cli/ui"
	"github.com/unweave/unweave/api/types"
)

// SSH handles the Cobra command for SSH
func SSH(cmd *cobra.Command, args []string) error {
	return sessionConnect(cmd, false, args)
}

// sessionConnect handles the flow to spawn a new SSH connection to an exec.
func sessionConnect(cmd *cobra.Command, withOpenVSCode bool, args []string) error {
	cmd.SilenceUsage = true
	ctx := cmd.Context()

	sshArgsByKey, err := parseSSHArgs(args)
	if err != nil {
		const errMsg = "❌ Invalid arguments. If you want to pass arguments to the ssh command, " +
			"use the -- flag. See `unweave ssh --help` for more information"
		ui.Errorf(errMsg)
		os.Exit(1)
	}

	// Lays the ground to fix the following:
	// TODO Fix a bug where unweave ssh {exec name / exec id} was not implemented/working
	execRef, err := getExecRefFromArgs(cmd.Context(), sshArgsByKey)
	if err != nil {
		ui.Errorf(err.Error())
		os.Exit(1)
	}

	execCh := make(chan types.Exec)
	errCh := make(chan error)
	isNew := false

	watchCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	if config.CreateExec {
		ui.Infof("Initializing node...")

		execCh, errCh, err = execCreateAndWatch(watchCtx, types.ExecConfig{}, types.GitConfig{})
		if err != nil {
			return err
		}
		isNew = true
	} else {
		var createNewExec bool
		if execRef == "" {
			execRef, createNewExec, err = sessionSelectSSHExecRef(cmd, execRef, false)
			if err != nil {
				return err
			}
		}

		if createNewExec {
			execCh, errCh, err = execCreateAndWatch(ctx, types.ExecConfig{}, types.GitConfig{})
			if err != nil {
				return err
			}
			isNew = true
		} else {
			execCh, errCh, err = session.Wait(ctx, execRef)
			if err != nil {
				return err
			}
		}
	}

	if withOpenVSCode {
		return handleSSHConnWithOpenVSCode(ctx, execCh, isNew, sshArgsByKey, errCh)
	} else {
		return handleSSHConn(ctx, execCh, isNew, sshArgsByKey, errCh)
	}
}

// parseSSHArgs dynamically parses the SSH connection arguments and returns them as a map.
func parseSSHArgs(args []string) (map[string]string, error) {
	sshArgs := make(map[string]string)

	if len(args) < 1 {
		return sshArgs, nil
	}
	if len(args) == 1 {
		sshArgs["execRef"] = args[0]
		return sshArgs, nil
	}
	if args[0] != "--" {
		return nil, errors.New("invalid arguments format. Use '--' as the separator")
	}

	// Iterate over the arguments starting from the second one
	for i := 1; i < len(args); i += 2 {
		key := strings.TrimPrefix(args[i], "--")
		if i+1 >= len(args) {
			return nil, errors.New("invalid arguments format. Missing value for flag: " + key)
		}
		value := args[i+1]
		sshArgs[key] = value
	}

	return sshArgs, nil
}

// getExecRefFromArgs returns the execRef from SSH Arguments, checks if it exists
func getExecRefFromArgs(ctx context.Context, sshArgsByKey map[string]string) (execRef string, err error) {
	execRef = sshArgsByKey["execRef"]
	if execRef == "" {
		return "", nil
	}

	e, err := getExecByNameOrID(ctx, execRef)
	if err != nil {
		return "", err
	}
	if e != nil {
		return e.ID, nil
	}

	return "", fmt.Errorf("session ID %s does not exist", execRef)
}

func handleSSHConn(ctx context.Context, execCh chan types.Exec, isNew bool, sshArgsByKey map[string]string, errCh chan error) error {
	for {
		select {
		case e := <-execCh:
			if e.Status == types.StatusRunning {
				ensureHosts(e)
				defer cleanupHosts(e)
				privKey := getPrivateKeyPathFromArgs(ctx, e, sshArgsByKey)

				err := handleCopySourceDir(isNew, e, privKey)
				if err != nil {
					return err
				}

				if err := ssh.Connect(ctx, *e.Connection, privKey); err != nil {
					ui.Errorf("%s", err)
					os.Exit(1)
				}

				if terminate := ui.Confirm("SSH session terminated. Do you want to terminate the session?", "n"); terminate {
					if err := sessionTerminate(ctx, e.ID); err != nil {
						return err
					}
					ui.Infof("Session %q terminated.", e.ID)
				}
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

func handleSSHConnWithOpenVSCode(ctx context.Context, execCh chan types.Exec, isNew bool, sshArgsByKey map[string]string, errCh chan error) error {
	for {
		select {
		case e := <-execCh:
			if e.Status == types.StatusRunning {
				ensureHosts(e)
				defer cleanupHosts(e)
				privKey := getPrivateKeyPathFromArgs(ctx, e, sshArgsByKey)

				err := handleCopySourceDir(isNew, e, privKey)
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

func cleanupHosts(e types.Exec) {
	if err := ssh.RemoveHost("uw:" + e.ID); err != nil {
		ui.Debugf("Failed to remove host from ssh config: %v", err)
	}
}

func ensureHosts(e types.Exec) {
	if e.Connection == nil {
		ui.Errorf("❌ Something unexpected happened. No connection info found for session %q", e.ID)
		ui.Infof("Run `unweave ls` to see the status of your session and try connecting manually.")
		os.Exit(1)
	}

	ui.Infof("🚀 Session %q up and running", e.ID)

	if err := ssh.RemoveKnownHostsEntry(e.Connection.Host); err != nil {
		// Log and continue anyway. Most likely the entry is not there.
		ui.Debugf("Failed to remove known_hosts entry: %v", err)
	}

	if err := ssh.AddHost("uw:"+e.ID, e.Connection.Host, e.Connection.User, e.Connection.Port); err != nil {
		ui.Debugf("Failed to add host to ssh config: %v", err)
	}
}

func handleCopySourceDir(isNew bool, e types.Exec, privKey string) error {
	// TODO: Wait until port is open before cleaning up the source code

	if !config.NoCopySource && isNew {
		dir, err := config.GetActiveProjectPath()
		if err != nil {
			ui.Errorf("Failed to get active project path. Skipping copying source directory")
			return fmt.Errorf("failed to get active project path: %v", err)
		}
		if err := copySource(e.ID, dir, "/home/ubuntu", *e.Connection, privKey); err != nil {
			fmt.Println(err)
		}
	} else {
		ui.Infof("Skipping copying source directory")
	}
	return nil
}

func copySource(execID, rootDir, dstPath string, connectionInfo types.ConnectionInfo, privKeyPath string) error {
	tmpFile, err := createTempContextFile(execID)
	if err != nil {
		return err
	}

	ui.Infof("🧳 Gathering context from %q", rootDir)

	if err := gatherContext(rootDir, tmpFile, "tar"); err != nil {
		return fmt.Errorf("failed to gather context: %v", err)
	}

	tmpDstPath := filepath.Join("/tmp", fmt.Sprintf("uw-context-%s.tar.gz", execID))

	if err := executeSCP(tmpFile.Name(), tmpDstPath, dstPath, connectionInfo, privKeyPath); err != nil {
		return fmt.Errorf("failed to copy source: %w", err)
	}

	if err := executeSSH(tmpDstPath, dstPath, connectionInfo, privKeyPath); err != nil {
		return fmt.Errorf("failed to extract source: %w", err)
	}

	ui.Infof("✅  Successfully copied source directory to remote host")

	return nil
}

func createTempContextFile(execID string) (*os.File, error) {
	name := fmt.Sprintf("uw-context-%s.tar.gz", execID)
	tmpFile, err := os.CreateTemp(os.TempDir(), name)
	if err != nil {
		return nil, err
	}
	return tmpFile, nil
}

func executeSCP(srcPath, tmpDstPath, dstPath string, connectionInfo types.ConnectionInfo, privKeyPath string) error {
	scpCommandArgs := []string{"-r",
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
	}
	if privKeyPath != "" {
		scpCommandArgs = append(scpCommandArgs, "-i", privKeyPath)
	}
	scpCommandArgs = append(scpCommandArgs, srcPath, fmt.Sprintf("%s@%s:%s", connectionInfo.User, connectionInfo.Host, tmpDstPath))

	scpCommand := exec.Command("scp", scpCommandArgs...)

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	scpCommand.Stdout = stdout
	scpCommand.Stderr = stderr

	ui.Infof("🔄 Copying source to %q", dstPath)

	if err := scpCommand.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			// Exited with non-zero exit code
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				if status.ExitStatus() == 255 {
					ui.Infof("The remote host closed the connection.")
					return fmt.Errorf("failed to copy source: %w", err)
				}
			}
			ui.Infof("Failed to copy source directory to remote host: %s", stderr.String())
			return err
		}
		return fmt.Errorf("scp command failed: %v", err)
	}

	return nil
}

func executeSSH(srcPath, dstPath string, connectionInfo types.ConnectionInfo, privKeyPath string) error {
	sshCommand := exec.Command(
		"ssh",
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-i", privKeyPath,
		fmt.Sprintf("%s@%s", connectionInfo.User, connectionInfo.Host),
		fmt.Sprintf("tar -xzf %s -C %s && rm -rf %s", srcPath, dstPath, srcPath),
	)

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	sshCommand.Stdout = stdout
	sshCommand.Stderr = stderr

	if err := sshCommand.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			// Exited with non-zero exit code
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				if status.ExitStatus() == 255 {
					ui.Infof("The remote host closed the connection.")
					return fmt.Errorf("failed to copy source: %w", err)
				}
			}
			ui.Infof("Failed to extract source directory on remote host: %s", stderr.String())
			return err
		}
		return fmt.Errorf("failed to unzip on remote host: %v", err)
	}

	return nil
}

// getPrivateKeyPathFromArgs returns the first best private key path from SSH arguments for an Exec
func getPrivateKeyPathFromArgs(ctx context.Context, e types.Exec, sshArgsByKey map[string]string) string {
	keysFolder := config.GetUnweaveSSHKeysFolder()

	// Case where a local key can be parsed from the pub flag
	if localKey, ok := sshArgsByKey["prv"]; ok {
		return localKey
	}

	// Case where we know what the public key was when the session was created
	dirEntries, err := os.ReadDir(keysFolder)
	if err != nil {
		ui.HandleError(err)
		os.Exit(1)
	}

	list := filterPublicKeys(dirEntries)
	for _, key := range list {
		// Unweave private keys are represented by the public key name
		name := strings.TrimSuffix(key.Name(), ".pub")
		if e.SSHKey.Name == name {
			privateKeyPath := filepath.Join(keysFolder, name)
			return privateKeyPath
		}
	}

	return ""
}
