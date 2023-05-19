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
	execRef, sshArgs := parseArgsToExecRefAndSSHArgs(args)
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

				if err := ssh.Connect(ctx, *e.Connection, prvKey, sshArgs); err != nil {
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

// getOrCreateExec handles the flow to spawn a new Exec or get an existing one, passes it to a Channel
func getOrCreateExec(cmd *cobra.Command, execRef string) (chan types.Exec, bool, chan error) {
	cmd.SilenceUsage = true
	ctx := cmd.Context()

	execCh := make(chan types.Exec)
	errCh := make(chan error)
	isNew := false
	var err error

	watchCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	if config.CreateExec {
		ui.Infof("Initializing node...")

		_, errCh, err := execCreateAndWatch(watchCtx, types.ExecConfig{}, types.GitConfig{})
		if err != nil {
			errCh <- err
			return nil, false, errCh
		}
		isNew = true
	} else {
		var createNewExec bool
		if execRef == "" {
			execRef, createNewExec, err = sessionSelectSSHExecRef(cmd, execRef, false)
			if err != nil {
				errCh <- err
				return nil, false, errCh
			}
		}

		if createNewExec {
			execCh, errCh, err = execCreateAndWatch(ctx, types.ExecConfig{}, types.GitConfig{})
			if err != nil {
				errCh <- err
				return nil, false, errCh
			}
			isNew = true
		} else {
			execCh, errCh, err = session.Wait(ctx, execRef)
			if err != nil {
				errCh <- err
				return nil, false, errCh
			}
		}
	}

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
