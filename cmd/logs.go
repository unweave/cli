package cmd

import (
	"errors"
	"os"

	"github.com/spf13/cobra"
	"github.com/unweave/cli/config"
	"github.com/unweave/cli/ssh"
	"github.com/unweave/cli/ui"
)

const execLogFile = "/tmp/exec.log"

func Logs(cmd *cobra.Command, args []string) error {

	if len(args) == 0 {
		const errMsg = "❌ Invalid arguments. Missing session-id or name. " +
			"See `unweave logs --help` for more information"
		ui.Errorf(errMsg)
		os.Exit(1)
	}

	if len(args) > 1 {
		const errMsg = "❌ Invalid arguments. You must pass the session-id or name only. " +
			"See `unweave logs --help` for more information"
		ui.Errorf(errMsg)
		os.Exit(1)
	}

	execRef := args[0]
	ctx := cmd.Context()

	e, err := getExecByNameOrID(ctx, execRef)
	if err != nil {
		return errors.New("Could not find session by name or ID")
	}

	command := []string{"tail"}

	if config.FollowLogs {
		command = append(command, "-f")
	}

	command = append(command, execLogFile)

	prvKey := config.SSHPrivateKeyPath
	if err := ssh.Connect(ctx, e.Network, prvKey, config.SSHConnectionOptions, command); err != nil {
		ui.Errorf("%s", err)
		os.Exit(1)
	}

	return nil
}
