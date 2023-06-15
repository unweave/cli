package volume

import (
	"context"
	"fmt"
	"os"

	"github.com/unweave/cli/config"
	"github.com/unweave/cli/ui"
	"github.com/unweave/unweave/api/types"
)

// Create creates a new volume, size in GB
func Create(ctx context.Context, name string, size int) (types.Volume, error) {
	if size <= 0 {
		size = config.DefaultVolumeSize
	}

	client := config.InitUnweaveClient()
	projectOwner, projectName := config.GetProjectOwnerAndName()
	projectProvider := config.Provider

	volume, err := client.Volume.Create(ctx, projectOwner, projectName, types.VolumeCreateRequest{
		Size:     size,
		Name:     name,
		Provider: types.Provider(projectProvider),
	})

	if err != nil {
		return types.Volume{}, fmt.Errorf("failed to create volume: %s", err)
	}
	return volume, nil
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
func List(ctx context.Context) ([]types.Volume, error) {
	ownerID, projectID := config.GetProjectOwnerAndName()

	var project string
	if projectID != "" {
		project = projectID
	}

	client := config.InitUnweaveClient()
	volumes, err := client.Volume.List(ctx, ownerID, project)
	if err != nil {
		return nil, fmt.Errorf("failed to list volumes: %s", err)
	}
	return volumes, nil
}

// Update updates an existing volume
func Update(ctx context.Context, name string, newSize int) error {
	userID, projectID := config.GetProjectOwnerAndName()

	client := config.InitUnweaveClient()
	err := client.Volume.Update(ctx, userID, projectID, name, types.VolumeResizeRequest{
		IDOrName: name,
		Size:     newSize,
	})

	if err != nil {
		return fmt.Errorf("failed to update volume: %s", err)
	}
	return nil
}
