package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/unweave/cli/entity"
)

func (h *Handler) List(ctx context.Context, cmd *entity.Command) error {
	projects, err := h.ctrl.GetProjects(ctx)
	if err != nil {
		return err
	}

	fmt.Println("Projects:")
	for _, p := range projects {
		fmt.Println("\t", p)
	}

	if (h.cfg.Root.Projects == nil) || (len(h.cfg.Root.Projects) == 0) {
		return nil
	}
	return nil
}

func ListCmd(cmd *cobra.Command, args []string) error {
	h := New()
	ctx := context.Background()
	return h.List(ctx, &entity.Command{
		Cmd:  cmd,
		Args: args,
	})
}
