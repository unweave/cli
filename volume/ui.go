package volume

import (
	"fmt"
	"time"

	"github.com/unweave/cli/ui"
	"github.com/unweave/unweave/api/types"
)

func RenderVolumesList(volumes []types.Volume) {
	cols := []ui.Column{
		{
			Title: "ID",
			Width: 5 + ui.MaxFieldLength(volumes, func(volume types.Volume) string {
				return volume.ID
			}),
		},
		{
			Title: "Name",
			Width: 5 + ui.MaxFieldLength(volumes, func(volume types.Volume) string {
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

	rows := make([]ui.Row, len(volumes))

	if len(volumes) == 0 {
		ui.Infof("No existing volumes")
		return
	}

	for idx, volume := range volumes {
		rows[idx] = []string{
			volume.ID,
			volume.Name,
			fmt.Sprintf("%d GB", volume.Size),
			volume.State.CreatedAt.Format(time.RFC3339),
			volume.Provider.DisplayName(),
		}
	}

	ui.Table("Volumes", cols, rows)
}
