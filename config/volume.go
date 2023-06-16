package config

import (
	"strings"

	"github.com/unweave/unweave/api/types"
)

const DefaultVolumeSize = 4

// GetVolumeAttachParams reads the existing config and
func GetVolumeAttachParams() []types.VolumeAttachParams {
	if len(Volumes) == 0 {
		return nil
	}

	var params = make([]types.VolumeAttachParams, 0, len(Volumes))
	for _, volume := range Volumes {
		ref, mntPath, exists := strings.Cut(volume, ":")
		if !exists {
			continue
		}

		params = append(params, types.VolumeAttachParams{
			VolumeRef: ref,
			MountPath: mntPath,
		})
	}

	return params
}
