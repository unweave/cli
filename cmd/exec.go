package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/unweave/cli/config"
	"github.com/unweave/cli/ui"
	"github.com/unweave/unweave/api/types"
)

func Exec(cmd *cobra.Command, args []string) error {
	return runSSHConnectionCommand(cmd, args, &execCommandFlow{})
}

type execCommandFlow struct{}

func (e *execCommandFlow) parseArgs(cmd *cobra.Command, args []string) execCmdArgs {
	execArgs := execCmdArgs{
		execRef:              "",
		sshConnectionOptions: config.SSHConnectionOptions,
	}

	doubleDashIdx := cmd.ArgsLenAtDash()
	argsBeforeDoubleDash := []string{}

	if doubleDashIdx == -1 {
		const errMsg = "❌ Invalid arguments. You must pass the -- flag. " +
			"See `unweave exec --help` for more information"
		ui.Errorf(errMsg)
		os.Exit(1)
	}

	if doubleDashIdx >= 0 {
		argsBeforeDoubleDash = args[:doubleDashIdx]
		execArgs.userCommand = args[doubleDashIdx:]
	}

	if len(execArgs.userCommand) == 0 {
		const errMsg = "❌ Invalid arguments. You must pass a command after the -- flag. " +
			"See `unweave exec --help` for more information"
		ui.Errorf(errMsg)
		os.Exit(1)
	}

	if len(argsBeforeDoubleDash) > 0 {
		const errMsg = "❌ Invalid arguments. You may not pass argment before the -- flag. " +
			"See `unweave exec --help` for more information"
		ui.Errorf(errMsg)
		os.Exit(1)
	}

	execArgs.execCommand = execArgs.userCommand

	if !config.ExecAttach {
		escaped := strings.ReplaceAll(strings.Join(execArgs.userCommand, " "), "\"", "\\\"")

		execArgs.execCommand = []string{"nohup", "bash", "-c", "\"", escaped, "\""}
		execArgs.execCommand = append(execArgs.execCommand, ">", execLogFile, "2>&1", "&", "echo", "$!", ">", "./pid.nohup", "&&", "sleep", "1")
	}

	return execArgs
}

func (e *execCommandFlow) getExec(cmd *cobra.Command, execCmd execCmdArgs) (chan types.Exec, bool, chan error) {
	ctx := cmd.Context()

	sendError := func(err error) chan error {
		errCh := make(chan error, 1)
		errCh <- err
		return errCh
	}

	ui.Infof("Initializing session...")

	execCh, errCh, err := execCreateAndWatch(ctx, types.ExecConfig{Command: execCmd.userCommand}, types.GitConfig{})
	if err != nil {
		return nil, false, sendError(err)
	}

	const alwaysNewExec = true

	return execCh, alwaysNewExec, errCh
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
