package volumes

import (
	"context"
	"os"

	"github.com/unweave/cli/config"
	"github.com/unweave/cli/ui"
	"github.com/unweave/unweave/api/types"
)

// Update updates an existing volume
func Update(ctx context.Context, name string, newSize int) {
	defer os.Exit(1)

	if name == "" {
		ui.Attentionf("❌ Please provide a valid volume name")
		os.Exit(1)
	}

	if newSize <= 0 {
		ui.Attentionf("❌ Please provide a valid volume size")
		os.Exit(1)
	}

	userID, projectID := config.GetProjectOwnerAndName()
	if name == "" {
		ui.Attentionf("❌ Please provide a valid volume name")
		os.Exit(1)
	}

	err := rest.Volume.Update(ctx, userID, projectID, name, types.VolumeResizeRequest{
		IDOrName: name,
		Size:     newSize,
	})

	if err != nil {
		ui.Attentionf("❌ Failed to update the volume: " + err.Error())
		os.Exit(1)
	}

	ui.Successf("✅ Volume updated successfully")
}
