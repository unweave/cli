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
