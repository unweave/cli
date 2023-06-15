package volumes

import (
	"fmt"
	"time"

	"github.com/unweave/cli/ui"
	"github.com/unweave/unweave/api/types"
)

func renderVolumes(volumes []types.Volume) {
	cols := []ui.Column{ui.Column{
		Title: "ID",
		Width: 5 + ui.MaxFieldLength(volumes, func(volume types.Volume) string {
			return volume.ID
		})},
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
		}}

	rows := make([]ui.Row, 0, len(volumes))

	for _, volume := range volumes {
		row := []string{
			volume.ID,
			volume.Name,
			fmt.Sprintf("%d GB", volume.Size),
			volume.State.CreatedAt.Format(time.RFC3339),
			volume.Provider.DisplayName(),
		}
		rows = append(rows, row)
	}

	ui.Table("Volumes", cols, rows)
}
