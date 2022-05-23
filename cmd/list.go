package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/unweave/cli/entity"
)

func findLinkedPath(projects map[string]entity.ProjectConfig, projectId string) (
	string, bool,
) {
	for path, project := range projects {
		if project.Id == projectId {
			return path, true
		}
	}
	return "", false
}

func (h *Handler) List(ctx context.Context, cmd *entity.Command) error {
	projects, err := h.ctrl.GetProjects(ctx)
	if err != nil {
		return err
	}

	fmt.Println("Projects:")
	for _, p := range projects {
		if path, ok := findLinkedPath(h.cfg.Root.Projects, p.Id); ok {
			fmt.Printf("  %s \n\tID: %s \n\tPath: %s\n", p.Name, p.Id, path)
		} else {
			fmt.Printf("  %s \n\tID: %s \n\tPath: Not linked\n", p.Name, p.Id)
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
	return h.List(ctx, &entity.Command{
		Cmd:  cmd,
		Args: args,
	})
}
