package cmd

import (
	"github.com/spf13/cobra"
)

func Code(cmd *cobra.Command, args []string) error {
	return sessionConnect(cmd, true, args)
}
