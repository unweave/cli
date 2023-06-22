package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/unweave/cli/config"
	"github.com/unweave/cli/session"
	"github.com/unweave/cli/ui"
	"github.com/unweave/unweave/api/types"
)

func Exec(cmd *cobra.Command, args []string) error {
	return runSSHConnectionCommand(cmd, args, &execCommandFlow{})
}

type execCommandFlow struct{}

func (e *execCommandFlow) parseArgs(cmd *cobra.Command, args []string) (execRef string, sshConnectionOptions []string, command []string) {
	var execArgs []string

	doubleDashIdx := cmd.ArgsLenAtDash()

	if doubleDashIdx == -1 {
		const errMsg = "❌ Invalid arguments. You must pass the -- flag. " +
			"See `unweave exec --help` for more information"
		ui.Errorf(errMsg)
		os.Exit(1)
	}

	if doubleDashIdx >= 0 {
		execArgs = args[:doubleDashIdx]
		command = args[doubleDashIdx:]
	} else {
		execArgs = args
	}

	if len(command) == 0 {
		const errMsg = "❌ Invalid arguments. You must pass a command after the -- flag. " +
			"See `unweave exec --help` for more information"
		ui.Errorf(errMsg)
		os.Exit(1)
	}

	if len(execArgs) > 1 {
		const errMsg = "❌ Invalid arguments. You may only pass one session-name or id to the exec command " +
			"before the -- flag. See `unweave exec --help` for more information"
		ui.Errorf(errMsg)
		os.Exit(1)
	}

	if len(execArgs) == 1 {
		execRef = execArgs[0]
	}

	if !config.ExecAttach {
		command = append([]string{"nohup"}, command...)
		command = append(command, ">", "exec.log", "2>&1", "&", "echo", "$!", ">", "./pid.nohup", "&&", "sleep", "1")
	}

	return execRef, sshConnectionOptions, command
}

func (e *execCommandFlow) getExec(cmd *cobra.Command, execRef string) (chan types.Exec, bool, chan error) {
	ctx := cmd.Context()

	sendError := func(err error) chan error {
		errCh := make(chan error, 1)
		errCh <- err
		return errCh
	}

	exec, err := getExecByNameOrID(ctx, execRef)
	if err != nil {
		return nil, false, sendError(err)
	}

	execCh, errCh, err := session.Wait(ctx, exec.ID)
	if err != nil {
		return nil, false, sendError(err)
	}

	return execCh, false, errCh
}

func (e *execCommandFlow) onTerminate(ctx context.Context, execID string) error {
	ui.Infof("Session %q exited. Use 'unweave terminate' to stop the session.", execID)

	return nil
}

// getExecs invokes the UnweaveClient and returns all container executions. Does not list terminated sessions by default
func getExecs(ctx context.Context) ([]types.Exec, error) {
	uwc := config.InitUnweaveClient()
	listTerminated := config.All
	owner, projectName := config.GetProjectOwnerAndName()
	return uwc.Exec.List(ctx, owner, projectName, listTerminated)
}

// getExecByNameOrID invokes the UnweaveClient and returns the container execution associated a name or ID that matches a given string
func getExecByNameOrID(ctx context.Context, ref string) (*types.Exec, error) {
	execs, err := getExecs(ctx)
	if err != nil {
		return nil, err
	}

	for _, e := range execs {
		// handle a case where the user accidentally input the execRef ID as the name
		if ref == e.Name {
			return &e, nil
		}
		if ref == e.ID {
			return &e, nil
		}
	}

	return nil, fmt.Errorf("session %s does not exist", ref)
}
