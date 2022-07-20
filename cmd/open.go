package cmd

import (
	"context"

	"github.com/skratchdot/open-golang/open"
	"github.com/spf13/cobra"
	"github.com/unweave/cli/model"
)

func (h *Handler) Open(ctx context.Context, cmd *model.Command) error {
	// TODO: check if project exists
	if len(cmd.Args) > 0 {
		return open.Run(h.cfg.Api.AppUrl + "/project/" + cmd.Args[0])
	}
	return open.Run(h.cfg.Api.AppUrl)
}

func OpenCmd(cmd *cobra.Command, args []string) error {
	h := New()
	ctx := context.Background()
	cmd.SilenceUsage = true
	return h.Open(ctx, &model.Command{
		Cmd:  cmd,
		Args: args,
	})
}
