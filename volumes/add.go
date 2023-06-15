package volumes

import (
	"context"
	"os"

	"github.com/unweave/cli/config"
	"github.com/unweave/cli/ui"
	"github.com/unweave/unweave/api/types"
)

var rest = config.InitUnweaveClient()

// Add creates a new volume, size in GB
func Add(ctx context.Context, name string, size int) {
	defer os.Exit(1)

	if name == "" {
		ui.Attentionf("❌ Please provide a valid volume name")
		os.Exit(1)
	}

	if size <= 0 {
		size = config.DefaultVolumeSize
	}
	projectOwner, projectName := config.GetProjectOwnerAndName()
	projectProvider := config.Provider

	created, err := rest.Volume.Create(ctx, projectOwner, projectName, types.VolumeCreateRequest{
		Size:     size,
		Name:     name,
		Provider: types.Provider(projectProvider),
	})

	if err != nil {
		ui.Attentionf("❌ Failed to create the volume: " + err.Error())
		os.Exit(1)
	}

	ui.Successf("✅ Volume created successfully")
	renderVolumes([]types.Volume{created})
}
