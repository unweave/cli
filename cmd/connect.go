package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/unweave/cli/entity"
)

func (h *Handler) Connect(ctx context.Context, cmd *entity.Command) error {
	projectID := cmd.Args[0]
	zeplID := cmd.Args[1]

	fmt.Printf("Connecting to zepl %s for project %s\n", projectID, zeplID)
	return h.ctrl.Connect(ctx, projectID, zeplID)
}

func ConnectCmd(cmd *cobra.Command, args []string) error {
	h := New()
	ctx := context.Background()
	return h.Connect(ctx, &entity.Command{
		Cmd:  cmd,
		Args: args,
	})
}
