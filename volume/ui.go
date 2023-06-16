package volume

import (
	"fmt"
	"time"

	"github.com/unweave/cli/ui"
	"github.com/unweave/unweave/api/types"
)

func RenderVolumesList(volumes []types.Volume, highlight *types.Volume) {
	const highlighted = " *"
	cols := []ui.Column{
		{
			Title: "Name",
			Width: 5 + ui.MaxFieldLength(volumes, func(volume types.Volume) string {
				if highlight != nil {
					return volume.Name + highlighted
				}
				return volume.Name
			}),
		}, {
			Title: "Size",
			Width: 5 + ui.MaxFieldLength(volumes, func(volume types.Volume) string {
				return fmt.Sprintf("%v", volume.Size)
			}),
		},
		{
			Title: "Created At",
			Width: 5 + ui.MaxFieldLength(volumes, func(volume types.Volume) string {
				return volume.State.CreatedAt.Format(time.RFC3339)
			}),
		}, {
			Title: "Provider",
			Width: 5 + ui.MaxFieldLength(volumes, func(volume types.Volume) string {
				return volume.Provider.DisplayName()
			}),
		},
	}

	rows := make([]ui.Row, 0, len(volumes))

	if len(volumes) == 0 {
		ui.Infof("No existing volumes")
		return
	}

	if highlight != nil {
		rows = append(rows, []string{
			highlight.Name + highlighted,
			fmt.Sprintf("%d GB", highlight.Size),
			highlight.State.CreatedAt.Format(time.RFC3339),
			highlight.Provider.DisplayName(),
		})
	}

	for _, volume := range volumes {
		if highlight != nil {
			if volume.ID == highlight.ID {
				continue
			}
		}
		rows = append(rows, []string{
			volume.Name,
			fmt.Sprintf("%d GB", volume.Size),
			volume.State.CreatedAt.Format(time.RFC3339),
			volume.Provider.DisplayName(),
		})
	}

	ui.Table("Volumes", cols, rows)
}
