package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/unweave/cli/config"
	"github.com/unweave/cli/entity"
)

func (h *Handler) Root(ctx context.Context, cmd *entity.Command) error {
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
	return h.Root(ctx, &entity.Command{
		Cmd:  cmd,
		Args: args,
	})
}
