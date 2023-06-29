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
	"time"

	"github.com/spf13/cobra"
	"github.com/unweave/cli/config"
	"github.com/unweave/cli/session"
	"github.com/unweave/cli/ssh"
	"github.com/unweave/cli/ui"
	"github.com/unweave/unweave/api/types"
)

// SSH handles the Cobra command for SSH
func SSH(cmd *cobra.Command, args []string) error {
	return runSSHConnectionCommand(cmd, args, &sshCommandFlow{})
}

type execCmdArgs struct {
	execRef              string
	sshConnectionOptions []string

	// userCommand is the string command
	// parts given by the user. This should
	// be recorded as what the user ran.
	userCommand []string

	// executeCommand is the string command
	// parts that will be executed on the exec
	// this should be used in the SSH connection.
	execCommand []string

	// attached denotes if the command will
	// run in attached mode, or detach.
	attached bool
}

type sshConnectionCommandFlow interface {
	parseArgs(cmd *cobra.Command, args []string) execCmdArgs
	getExec(cmd *cobra.Command, command execCmdArgs) (chan types.Exec, bool, chan error)
	onTerminate(ctx context.Context, execID string) error
}

func runSSHConnectionCommand(cmd *cobra.Command, args []string, flow sshConnectionCommandFlow) error {
	commandArgs := flow.parseArgs(cmd, args)

	prvKey := config.SSHPrivateKeyPath
	execCh, isNew, errCh := flow.getExec(cmd, commandArgs)
	ctx := cmd.Context()

	for {
		select {
		case e := <-execCh:
			if e.Status == types.StatusRunning {
				defer cleanupHosts(e)
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

				if err := ssh.Connect(ctx, e.Network, prvKey, commandArgs.sshConnectionOptions, commandArgs.execCommand); err != nil {
					ui.Errorf("%s", err)
					os.Exit(1)
				}

				if err := flow.onTerminate(ctx, e.ID); err != nil {
					return err
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

type sshCommandFlow struct{}

func (s *sshCommandFlow) parseArgs(cmd *cobra.Command, args []string) execCmdArgs {
	command := execCmdArgs{}

	if len(args) == 1 {
		command.execRef = args[0]
	}

	// If the number of args is great than one, we always expect the first arg to be
	// the separator flag "--". If the number of args is one, we expect it to be the
	// execID or name
	if len(args) > 1 {
		command.execRef = args[0]
		command.sshConnectionOptions = args[1:]

		if command.sshConnectionOptions[0] != "--" {
			const errMsg = "❌ Invalid arguments. If you want to pass arguments to the ssh command, " +
				"use the -- flag. See `unweave ssh --help` for more information"
			ui.Errorf(errMsg)
			os.Exit(1)
		}
	}

	return command
}

func (s *sshCommandFlow) getExec(cmd *cobra.Command, command execCmdArgs) (chan types.Exec, bool, chan error) {
	return getOrCreateExec(cmd, command.execRef)
}

func (s *sshCommandFlow) onTerminate(ctx context.Context, execID string) error {
	if terminate := ui.Confirm("SSH session terminated. Do you want to terminate the session?", "n"); terminate {
		if err := sessionTerminate(ctx, execID); err != nil {
			return err
		}
		ui.Infof("Session %q terminated.", execID)
	}

	return nil
}

// getOrCreateExec handles the flow to spawn a new Exec or get an existing one, returns whether to expect a new Exec
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
			execRef, createNewExec, err = sessionSelectSSHExecRef(ctx, execRef, false)
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

	return execCh, isNew, nil
}

func cleanupHosts(e types.Exec) {
	if err := ssh.RemoveHost("uw:" + e.ID); err != nil {
		ui.Debugf("Failed to remove host from ssh config: %v", err)
	}
}

func ensureHosts(e types.Exec, identityFile string) {
	if e.Network.Host == "" {
		ui.Errorf("❌ Something unexpected happened. No connection info found for session %q", e.ID)
		ui.Infof("Run `unweave ls` to see the status of your session and try connecting manually.")
		os.Exit(1)
	}
	if e.Network.Port == 0 {
		ui.Errorf("❌ Something unexpected happened. No port info found for session %q", e.ID)
		ui.Infof("Run `unweave ls` to see the status of your session. If this problem persists please contact an administrator.")
		os.Exit(1)
	}

	ui.Infof("🚀 Session %q up and running", e.ID)

	if err := ssh.RemoveKnownHostsEntry(e.Network.Host); err != nil {
		// Log and continue anyway. Most likely the entry is not there.
		ui.Debugf("Failed to remove known_hosts entry: %v", err)
	}

	if err := ssh.AddHost("uw:"+e.ID, e.Network.Host, e.Network.User, e.Network.Port, identityFile); err != nil {
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
		if err := copyDirFromLocalAndUnzip(e.ID, dir, config.ProjectHostDir(), e.Network, privKey); err != nil {
			return err
		}
	} else {
		ui.Infof("Skipping copying source directory")
	}
	return nil
}

func copyDirFromLocalAndUnzip(execID, rootDir, dstPath string, connectionInfo types.ExecNetwork, privKeyPath string) error {
	ui.Infof("🧳 Gathering context from %q", rootDir)

	tmpFile, err := createTempContextFile(execID)
	if err != nil {
		return err
	}
	if err := gatherContext(rootDir, tmpFile, "tar"); err != nil {
		return fmt.Errorf("failed to gather context: %v", err)
	}

	tmpDstPath := filepath.Join("/tmp", fmt.Sprintf("uw-context-%s.tar.gz", execID))

	ui.Infof("🔄 Copying source to %q", dstPath)

	// target to copy to i.e. /home/user/Desktop/ user@your.server.example.com:/path/to/foo
	remoteTarget := fmt.Sprintf("%s@%s:%s", connectionInfo.User, connectionInfo.Host, tmpDstPath)
	if err := copySourceSCP(tmpFile.Name(), remoteTarget, privKeyPath); err != nil {
		return fmt.Errorf("failed to copy source: %w", err)
	}

	if err := copySourceUnTar(tmpDstPath, dstPath, connectionInfo, privKeyPath); err != nil {
		return fmt.Errorf("failed to extract source: %w", err)
	}

	ui.Infof("✅  Successfully copied source directory to remote host")

	return nil
}

func copyDirFromRemoteAndUnzip(sshTarget, localDirectory, privateKey string) error {
	ui.Infof("🧳 Gathering context from %q", sshTarget)

	remotePath, err := tarRemoteDirectory(sshTarget, privateKey)
	if err != nil {
		return fmt.Errorf("Failed to zip the remote directory. Expected both a remote target and directory in %s", sshTarget)
	}

	ui.Infof("📦 Copying the archive of the remote path to the host...")

	sshTargetAndDir := strings.Split(sshTarget, ":")
	if len(sshTargetAndDir) != 2 {
		return fmt.Errorf("Expected target to be in the format 'user@host:directory'")
	}
	sshTargetDirectory := sshTargetAndDir[0] + ":" + remotePath
	remoteFilename := filepath.Base(remotePath)
	archiveLocalTargetDir := config.GetGlobalConfigPath()
	archiveLocalTarget := filepath.Join(archiveLocalTargetDir, remoteFilename)

	err = copySourceSCP(sshTargetDirectory, config.GetGlobalConfigPath(), privateKey)
	if err != nil {
		return fmt.Errorf("Failed to copy the archive of your remote path to the host. "+
			"Please check if %s exists on the remote and ensure Unweave has the necessary permissions to access %s",
			sshTargetDirectory, archiveLocalTargetDir)
	}

	ui.Infof("🗜️ Unzipping the copied archive to the local directory...")

	cmd := exec.Command("tar", "-xf", archiveLocalTarget, "-C", localDirectory, "--strip-components=1")

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("Failed to unzip the copied archive of your code from %s to %s. "+
			"Please ensure that Unweave has the necessary permissions to access %s and perform this operation manually",
			archiveLocalTarget, localDirectory, archiveLocalTarget)
	}

	return nil
}

// tarRemoteDirectory takes an ssh target, and zips up the contents of that target to a returned in the remote /tmp
func tarRemoteDirectory(sshTarget, privateKeyPath string) (remoteArchiveLoc string, err error) {
	sshTargetAndDir := strings.Split(sshTarget, ":")
	if len(sshTargetAndDir) != 2 {
		return "", fmt.Errorf("Failed to zip remote directory, expected both a remote target and directory in %s", sshTarget)
	}

	timestamp := time.Now().Unix()
	remoteArchiveLoc = fmt.Sprintf("/tmp/uw-context-%d.tar.gz", timestamp)
	tarCmd := fmt.Sprintf("tar -czf %s -C %s .", remoteArchiveLoc, sshTargetAndDir[1])

	sshCommand := exec.Command(
		"ssh",
		"-tt",
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-i", privateKeyPath,
		sshTargetAndDir[0],
		tarCmd,
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
					return "", fmt.Errorf("failed to copy source: %w", err)
				}
			}
			ui.Infof("Failed to extract source directory on remote host: %s", stderr.String())
			return "", err
		}
		return "", fmt.Errorf("failed to unzip on remote host: %v", err)
	}

	return remoteArchiveLoc, nil
}

func createTempContextFile(execID string) (*os.File, error) {
	name := fmt.Sprintf("uw-context-%s.tar.gz", execID)
	tmpFile, err := os.CreateTemp(os.TempDir(), name)
	if err != nil {
		return nil, err
	}
	return tmpFile, nil
}

func copySourceSCP(from, to string, privKeyPath string) error {
	scpCommandArgs := []string{
		"-r",
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
	}
	if privKeyPath != "" {
		scpCommandArgs = append(scpCommandArgs, "-i", privKeyPath)
	}

	scpCommandArgs = append(scpCommandArgs, []string{from, to}...)

	scpCommand := exec.Command("scp", scpCommandArgs...)

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	scpCommand.Stdout = stdout
	scpCommand.Stderr = stderr

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

func copySourceUnTar(srcPath, dstPath string, connectionInfo types.ExecNetwork, prvKeyPath string) error {
	sshCommand := exec.Command(
		"ssh",
		"-tt",
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-i", prvKeyPath,
		fmt.Sprintf("%s@%s", connectionInfo.User, connectionInfo.Host),
		// ensure dstPath exist and root logs into that path
		fmt.Sprintf("mkdir -p %s && echo 'cd %s' > ~/.bashrc &&", dstPath, dstPath),
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
			ui.Infof("Failed to extract source directory on remote host")
			ui.Infof(stderr.String())
			ui.Infof(stdout.String())
			return err
		}
		return fmt.Errorf("failed to unzip on remote host: %v", err)
	}

	return nil
}

// getDefaultKey tries to find the private key for this Exec, or relies on
// a default key if it can't be found
func getDefaultKey(ctx context.Context, e types.Exec, defaultKey string) (string, error) {
	// if the user has specified their own private key location
	if defaultKey != "" {
		return defaultKey, nil
	}
	if len(e.Keys) == 0 {
		return defaultKey, fmt.Errorf("There are no SSH keys for exec ID %s, please contact an administrator if this is an error", e.ID)
	}
	execKeyName := e.Keys[0].Name

	keysFolder := config.GetUnweaveSSHKeysFolder()
	dirEntries, err := os.ReadDir(keysFolder)
	if err != nil {
		return "", fmt.Errorf("failed to read SSH keys folder: %w", err)
	}

	publicKeys := filterPublicKeys(dirEntries)

	// Ensure default key is never ""
	// Unweave private keys are trimmed public ones
	for _, key := range publicKeys {
		name := strings.TrimSuffix(key.Name(), ".pub")
		defaultKey = filepath.Join(keysFolder, name)
		if execKeyName == name {
			return defaultKey, nil
		}
	}

	if defaultKey == "" {
		privKeyPath, _, err := generateSSHKey(ctx)
		if err != nil {
			return "", fmt.Errorf("failed to generate SSH key: %w", err)
		}
		return privKeyPath, nil
	}

	return defaultKey, nil
}
