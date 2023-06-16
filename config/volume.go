package config

import (
	"fmt"
	"strings"

	"github.com/unweave/unweave/api/types"
)

const DefaultVolumeSize = 4

// GetVolumeAttachParams reads the existing config and
func GetVolumeAttachParams() ([]types.VolumeAttachParams, error) {
	if len(Volumes) == 0 {
		return nil, nil
	}

	var params = make([]types.VolumeAttachParams, len(Volumes))
	for idx, volume := range Volumes {
		ref, mntPath, exists := strings.Cut(volume, ":")
		if !exists {
			return nil, fmt.Errorf("volume name %s is an invalid format", volume)
		}

		params[idx] = types.VolumeAttachParams{
			VolumeRef: ref,
			MountPath: mntPath,
		}
	}

	return params, nil
}
