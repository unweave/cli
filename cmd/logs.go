package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/unweave/cli/model"
)

func (h *Handler) Logs(ctx context.Context, cmd *model.Command) error {
	zeplID := cmd.Args[0]
	fmt.Printf("Fetching logs from zepl %s\n", zeplID)
	return h.ctrl.Logs(ctx, zeplID)
}

func LogsCmd(cmd *cobra.Command, args []string) error {
	h := New()
	ctx := context.Background()
	cmd.SilenceUsage = true
	return h.Logs(ctx, &model.Command{
		Cmd:  cmd,
		Args: args,
	})
}
