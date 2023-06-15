package cmd

import (
	"github.com/spf13/cobra"
	"github.com/unweave/cli/config"
	"github.com/unweave/cli/volumes"
)

func VolumeAdd(cmd *cobra.Command, args []string) error {
	var name string

	if len(args) > 0 {
		name = args[0]
	}

	volumes.Add(cmd.Context(), name, config.VolumeSize)

	return nil
}

func VolumeUpdate(cmd *cobra.Command, args []string) error {
	var name string

	if len(args) > 0 {
		name = args[0]
	}

	volumes.Update(cmd.Context(), name, config.VolumeSize)

	return nil
}

func VolumeList(cmd *cobra.Command, args []string) error {
	volumes.List(cmd.Context())

	return nil
}
