package cmd

import (
	"context"
	"github.com/spf13/cobra"
	"github.com/unweave/cli/entity"
)

func (h *Handler) Init(ctx context.Context, cmd *entity.Command) error {
	return nil
}

func InitCmd(cmd *cobra.Command, args []string) error {
	h := New()
	ctx := context.Background()
	cmd.SilenceUsage = true
	return h.Init(ctx, &entity.Command{
		Cmd:  cmd,
		Args: args,
	})
}
