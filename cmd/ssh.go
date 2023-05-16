package cmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/unweave/cli/config"
	"github.com/unweave/cli/session"
	"github.com/unweave/cli/ssh"
	"github.com/unweave/cli/ui"
	"github.com/unweave/unweave/api/types"
)

func copySource(execID, rootDir, dstPath string, connectionInfo types.ConnectionInfo, keyPath string) error {
	name := fmt.Sprintf("uw-context-%s.tar.gz", execID)
	tmpFile, err := os.CreateTemp(os.TempDir(), name)
	if err != nil {
		return err
	}

	ui.Infof("🧳 Gathering context from %q", rootDir)

	if err := gatherContext(rootDir, tmpFile, "tar"); err != nil {
		return fmt.Errorf("failed to gather context: %v", err)
	}

	tmpDstPath := filepath.Join("/tmp", name)
	scpCommand := exec.Command(
		"scp",
		"-r",
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		tmpFile.Name(),
		fmt.Sprintf("%s@%s:%s", connectionInfo.User, connectionInfo.Host, tmpDstPath),
	)
	if keyPath != "" {
		scpCommand.Args = append(scpCommand.Args, "-i", keyPath)
	}

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

	// Unzip on remote host
	sshCommand := exec.Command(
		"ssh",
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-i", keyPath,
		fmt.Sprintf("%s@%s", connectionInfo.User, connectionInfo.Host),
		fmt.Sprintf("tar -xzf %s -C %s && rm -rf %s", tmpDstPath, dstPath, tmpDstPath),
	)

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

	ui.Infof("✅  Successfully copied source directory to remote host")

	return nil
}

func SSH(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true
	ctx := cmd.Context()

	// TODO: parse args and forward to ssh command

	errMsg := "❌ Invalid arguments. If you want to pass arguments to the ssh command, " +
		"use the -- flag. See `unweave ssh --help` for  more information."

	execCh := make(chan types.Exec)
	errCh := make(chan error)
	isNew := false

	var err error
	var sshArgs []string
	var execRef string // Can be execID or name

	// If the number of args is great than one, we always expect the first arg to be
	// the separator flag "--". If the number of args is one, we expect it to be the
	// execID or name
	if len(args) > 1 {
		sshArgs = args[1:]
		if sshArgs[0] != "--" {
			ui.Errorf(errMsg)
			os.Exit(1)
		}

		if len(args) == 1 {
			execRef = args[0]
		} else {
			execRef = args[0]
		}
	}

	watchCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	if config.CreateExec {
		// If the flag to create a new exec is passed, any arguments must be forwarded to
		// the ssh command
		if len(args) > 0 && args[0] != "--" {
			ui.Errorf(errMsg)
			os.Exit(1)
		}

		ui.Infof("Initializing node...")

		execCh, errCh, err = execCreateAndWatch(watchCtx, types.ExecConfig{}, types.GitConfig{})
		if err != nil {
			return err
		}
		isNew = true

	} else {

		if execRef == "" {
			var execs []types.Exec

			execRef, execs, err = selectExec(cmd.Context(), "Select a session to connect to")
			if err != nil {
				return err
			}
			if len(execs) == 0 {
				ui.Errorf("❌ No active sessions found and no session name or ID provided. If " +
					"you want to create a new session, use the --new flag.")
				os.Exit(1)
			}
		}

		execCh, errCh, err = session.Wait(watchCtx, execRef)
		if err != nil {
			return err
		}
	}

	for {
		select {
		case e := <-execCh:
			if e.Status == types.StatusRunning {
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

				defer func() {
					if e := ssh.RemoveHost("uw:" + e.ID); e != nil {
						ui.Debugf("Failed to remove host from ssh config: %v", e)
					}
				}()

				if !config.NoCopySource && isNew {
					dir, err := config.GetActiveProjectPath()
					if err != nil {
						ui.Errorf("Failed to get active project path. Skipping copying source directory")
						return fmt.Errorf("failed to get active project path: %v", err)
					}

					if err := copySource(e.ID, dir, "/home/ubuntu", *e.Connection, ""); err != nil {
						fmt.Println(err)
					}
				} else {
					ui.Infof("Skipping copying source directory")
				}

				if err := ssh.Connect(ctx, *e.Connection, args); err != nil {
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
