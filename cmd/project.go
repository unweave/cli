package cmd

import (
	"github.com/spf13/cobra"
	"github.com/unweave/cli/project"
	"github.com/unweave/cli/ui"
)

func ProjectCreate(cmd *cobra.Command, args []string) error {
	var name = args[0]

	if name == "" {
		ui.Fatal("Invalid project name", nil)
	}

	project, err := project.Create(cmd.Context(), name, "")
	if err != nil {
		ui.Fatal("Failed to create project", err)
	}

	ui.Successf("✅ Project %q created successfully", project.Name)

	return nil
}