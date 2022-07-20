package cmd

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/unweave/cli/model"
)

func (h *Handler) Logout(ctx context.Context, cmd *model.Command) error {
	return h.ctrl.Logout(ctx)
}

func LogoutCmd(cmd *cobra.Command, args []string) error {
	h := New()
	ctx := context.Background()
	cmd.SilenceUsage = true
	return h.Logout(ctx, &model.Command{
		Cmd:  cmd,
		Args: args,
	})
}
