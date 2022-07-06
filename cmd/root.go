package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/unweave/cli/config"
	"github.com/unweave/cli/model"
)

func (h *Handler) Root(ctx context.Context, cmd *model.Command) error {
	if config.ShowConfig {
		cfg := config.New()
		bytes, err := cfg.ToJson()
		if err != nil {
			return err
		}
		fmt.Println(string(bytes[:]))
		return nil
	}
	return cmd.Cmd.Usage()
}

func RootCmd(cmd *cobra.Command, args []string) error {
	h := New()
	ctx := context.Background()
	cmd.SilenceUsage = true
	return h.Root(ctx, &model.Command{
		Cmd:  cmd,
		Args: args,
	})
}
