package ui

import (
	"fmt"
	"strings"

	"github.com/muesli/reflow/indent"
	"github.com/unweave/unweave/api/types"
)

type ResultEntry struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func ResultTitle(title string) {
	Successf(title)
}

func Result(entries []ResultEntry, indentation uint) {
	str := ""
	maxWidth := 0
	// This probably a better way to do this but this is quick and easy,
	for _, entry := range entries {
		if len(entry.Key) > maxWidth {
			maxWidth = len(entry.Key)
		}
	}
	for _, entry := range entries {
		padding := maxWidth - len(entry.Key) + 1
		str += fmt.Sprintf("%s:%*s%s\n", entry.Key, -padding, "", entry.Value)
	}
	str = indent.String(str, indentation)
	Successf(str)
}

func FormatVolumes(volumes []types.ExecVolume) string {
	if len(volumes) == 0 {
		return "-"
	}

	var b strings.Builder

	for i, v := range volumes {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(fmt.Sprintf("%s:%s", v.VolumeID, v.MountPath))
	}

	return b.String()
}
