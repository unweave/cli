package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/unweave/cli/model"
)

func findLinkedPath(projects map[string]model.ProjectConfig, projectID string) (
	string, bool,
) {
	for path, project := range projects {
		if project.ID == projectID {
			return path, true
		}
	}
	return "", false
}

func (h *Handler) List(ctx context.Context, cmd *model.Command) error {
	projects, err := h.ctrl.GetProjects(ctx)
	if err != nil {
		return err
	}

	fmt.Println("Projects:")
	for _, p := range projects {
		if path, ok := findLinkedPath(h.cfg.Root.Projects, p.ID); ok {
			fmt.Printf("  %s \n\tID: %s \n\tPath: %s\n", p.Name, p.ID, path)
		} else {
			fmt.Printf("  %s \n\tID: %s \n\tPath: Not linked\n", p.Name, p.ID)
		}
	}

	if (h.cfg.Root.Projects == nil) || (len(h.cfg.Root.Projects) == 0) {
		return nil
	}
	return nil
}

func ListCmd(cmd *cobra.Command, args []string) error {
	h := New()
	ctx := context.Background()
	cmd.SilenceUsage = true
	return h.List(ctx, &model.Command{
		Cmd:  cmd,
		Args: args,
	})
}
