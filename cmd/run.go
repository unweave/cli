package cmd

import (
	"context"
	"github.com/spf13/cobra"
	"github.com/unweave/cli/entity"
)

func (h *Handler) Run(ctx context.Context, cmd *entity.Command) error {
	return h.ctrl.Run(ctx)
}

func RunCmd(cmd *cobra.Command, args []string) error {
	h := New()
	ctx := context.Background()
	return h.Run(ctx, &entity.Command{
		Cmd:  cmd,
		Args: args,
	})
}
