package cmd

import (
	"context"
	"github.com/unweave/cli/ui"
)

// renderCobraSelection invokes Cobra to select an exec with a prompt, returns the index of the selected ID
func renderCobraSelection(ctx context.Context, options []string, optIdByOptionIdx map[int]string, prompt string) (string, error) {
	selected, err := ui.Select(prompt, options)
	if err != nil {
		return "", err
	}
	return optIdByOptionIdx[selected], nil
}
