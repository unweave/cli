package volumes

import (
	"context"
	"os"

	"github.com/unweave/cli/config"
	"github.com/unweave/cli/ui"
)

// Delete deletes a volume
func Delete(ctx context.Context, name string) {
	defer os.Exit(1)

	if name == "" {
		ui.Attentionf("❌ Please provide a valid volume name")
		os.Exit(1)
	}

	projectOwner, projectName := config.GetProjectOwnerAndName()

	err := rest.Volume.Delete(ctx, projectOwner, projectName, name)

	if err != nil {
		ui.Attentionf("❌ Failed to delete the volume: " + err.Error())
		os.Exit(1)
	}

	ui.Successf("✅ Volume deleted successfully")
}
