package cmd

import (
	"context"
	"github.com/spf13/cobra"
	"github.com/unweave/cli/entity"
	"strings"
)

func (h *Handler) Run(ctx context.Context, cmd *entity.Command) error {
	command := strings.Join(cmd.Args, " ")
	return h.ctrl.Run(ctx, command)
}

func RunCmd(cmd *cobra.Command, args []string) error {
	h := New()
	ctx := context.Background()
	cmd.SilenceUsage = true
	return h.Run(ctx, &entity.Command{
		Cmd:  cmd,
		Args: args,
	})
}
