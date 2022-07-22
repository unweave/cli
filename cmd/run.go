package cmd

import (
	"context"
	"strings"

	"github.com/spf13/cobra"
	"github.com/unweave/cli/model"
)

func (h *Handler) Run(ctx context.Context, cmd *model.Command) error {
	command := strings.Join(cmd.Args, " ")
	return h.ctrl.Run(ctx, command)
}

func RunCmd(cmd *cobra.Command, args []string) error {
	h := New()
	cmd.SilenceUsage = true
	return h.Run(cmd.Context(), &model.Command{
		Cmd:  cmd,
		Args: args,
	})
}
