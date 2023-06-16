package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/unweave/cli/config"
	"github.com/unweave/cli/ui"
	"github.com/unweave/cli/volume"
)

func VolumeCreate(cmd *cobra.Command, args []string) error {
	var name = args[0]
	var sizeStr *string

	if len(args) > 1 {
		sizeStr = &args[1]
	}
	if name == "" {
		ui.Fatal("Invalid volume name", nil)
	}

	var size = config.DefaultVolumeSize
	if sizeStr != nil {
		sizei, err := strconv.ParseInt(*sizeStr, 10, 64)
		if err != nil {
			ui.Fatal("Size must be a valid integer ", nil)
		}
		size = int(sizei)
	}

	vol, err := volume.Create(cmd.Context(), name, size)
	if err != nil {
		ui.Debugf("Failed to create volume: %s", err.Error())
		ui.Fatal("Failed to create volume", err)
	}

	ui.Successf("✅ Volume created successfully")

	volumes, err := volume.List(cmd.Context())
	if err != nil {
		ui.Fatal("There was a problem rendering the newly created volume", err)
	}
	volume.RenderVolumesList(volumes, &vol)

	return nil
}

func VolumeDelete(cmd *cobra.Command, args []string) error {
	var name string

	if len(args) > 0 {
		name = args[0]
	}

	if name == "" {
		ui.Errorf("Invalid volume name")
		os.Exit(1)
	}

	msg := fmt.Sprintf("Are you sure you want to delete volume %q? "+
		"This will permanently delete any data in the volume", name)

	confirm := ui.Confirm(msg, "n")
	if !confirm {
		return nil
	}

	err := volume.Delete(cmd.Context(), name)
	if err != nil {
		ui.Debugf("Failed to delete volume: %s", err.Error())
		ui.Errorf("Failed to delete volume")
		os.Exit(1)
	}

	ui.Successf("✅ Volume deleted successfully")

	return nil
}

func VolumeList(cmd *cobra.Command, args []string) error {
	volumes, err := volume.List(cmd.Context())
	if err != nil {
		ui.Debugf("Failed to list volumes: %s", err.Error())
		ui.Errorf("Failed to list volumes")
		os.Exit(1)
	}

	volume.RenderVolumesList(volumes, nil)

	return nil
}

func VolumeResize(cmd *cobra.Command, args []string) error {
	var name = args[0]
	var sizeStr *string

	if len(args) > 1 {
		sizeStr = &args[1]
	}
	if name == "" {
		ui.Fatal("Invalid volume name", nil)
	}

	var size = config.DefaultVolumeSize
	if sizeStr != nil {
		sizei, err := strconv.ParseInt(*sizeStr, 10, 64)
		if err != nil {
			ui.Fatal("Size must be a valid integer", nil)
		}
		size = int(sizei)
	}

	err := volume.Update(cmd.Context(), name, size)
	if err != nil {
		ui.Debugf("Failed to update the volume: %s", err.Error())
		ui.Errorf("Failed to update the volume")
	}

	ui.Successf("✅ Volume updated successfully")

	return nil
}
