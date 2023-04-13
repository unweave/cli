package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/unweave/unweave/api/types"
)

func BoxUp(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	params := types.ExecCtx{}

	if _, err := sessionCreate(cmd.Context(), params, true); err != nil {
		os.Exit(1)
		return nil
	}

	return nil
}
