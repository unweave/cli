package cmd

import (
	"context"
	"github.com/spf13/cobra"
	"github.com/unweave/cli/entity"
)

func (h *Handler) Login(ctx context.Context, cmd *entity.Command) error {
	// check if user token already exists
	// if not, check for token flag
	// if token, validate token with backend
	// else, execute browser login flow
	return nil
}

func LoginCmd(cmd *cobra.Command, args []string) error {
	h := New()
	ctx := context.Background()
	return h.Login(ctx, &entity.Command{
		Cmd:  cmd,
		Args: args,
	})
}
