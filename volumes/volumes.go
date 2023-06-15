package volumes

import (
	"context"
	"os"

	"github.com/unweave/cli/config"
	"github.com/unweave/cli/ui"
	"github.com/unweave/unweave/api/types"
)

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

	client := config.InitUnweaveClient()
	projectOwner, projectName := config.GetProjectOwnerAndName()
	projectProvider := config.Provider

	created, err := client.Volume.Create(ctx, projectOwner, projectName, types.VolumeCreateRequest{
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

// Delete deletes a volume
func Delete(ctx context.Context, name string) {
	defer os.Exit(1)

	if name == "" {
		ui.Attentionf("❌ Please provide a valid volume name")
		os.Exit(1)
	}

	client := config.InitUnweaveClient()
	projectOwner, projectName := config.GetProjectOwnerAndName()

	err := client.Volume.Delete(ctx, projectOwner, projectName, name)

	if err != nil {
		ui.Attentionf("❌ Failed to delete the volume: " + err.Error())
		os.Exit(1)
	}

	ui.Successf("✅ Volume deleted successfully")
}

// List lists all volumes for a given project or default project if none is specified
func List(ctx context.Context) {
	defer os.Exit(1)

	ownerID, projectID := config.GetProjectOwnerAndName()

	var project string
	if projectID != "" {
		project = projectID
	}

	client := config.InitUnweaveClient()
	volumes, err := client.Volume.List(ctx, ownerID, project)

	if err != nil {
		ui.Attentionf("❌ Failed to list the volumes: " + err.Error())
		os.Exit(1)
	}

	renderVolumes(volumes)
}

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

	client := config.InitUnweaveClient()
	err := client.Volume.Update(ctx, userID, projectID, name, types.VolumeResizeRequest{
		IDOrName: name,
		Size:     newSize,
	})

	if err != nil {
		ui.Attentionf("❌ Failed to update the volume: " + err.Error())
		os.Exit(1)
	}

	ui.Successf("✅ Volume updated successfully")
}
