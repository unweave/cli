package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/unweave/unweave/api/types"
)

func BoxUp(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true
	params := types.ExecCtx{}

	var filesystemID *string

	if len(args) > 0 {
		filesystemID = &args[0]
	}
	if _, err := sessionCreate(cmd.Context(), params, true, filesystemID); err != nil {
		os.Exit(1)
		return nil
	}

	return nil
}
