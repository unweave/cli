package cmd

import (
	"context"
	"github.com/spf13/cobra"
	"github.com/unweave/cli/entity"
)

func (h *Handler) Logout(ctx context.Context, cmd *entity.Command) error {
	return h.ctrl.Logout(ctx)
}

func LogoutCmd(cmd *cobra.Command, args []string) error {
	h := New()
	ctx := context.Background()
	return h.Logout(ctx, &entity.Command{
		Cmd:  cmd,
		Args: args,
	})
}
