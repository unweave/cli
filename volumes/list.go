package volumes

import (
	"context"
	"os"

	"github.com/unweave/cli/config"
	"github.com/unweave/cli/ui"
)

// List lists all volumes for a given project or default project if none is specified
func List(ctx context.Context) {
	defer os.Exit(1)

	ownerID, projectID := config.GetProjectOwnerAndName()

	var project string
	if projectID != "" {
		project = projectID
	}

	volumes, err := rest.Volume.List(ctx, ownerID, project)

	if err != nil {
		ui.Attentionf("❌ Failed to list the volumes: " + err.Error())
		os.Exit(1)
	}

	renderVolumes(volumes)
}
