package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/unweave/cli/model"
)

func (h *Handler) Login(ctx context.Context, cmd *model.Command) error {
	// Check if user token already exists
	if h.cfg.Root.User != nil && h.cfg.Root.User.Token != "" {
		fmt.Println("You are already logged in. Run `unweave logout` first for a fresh login")
		return nil
	}

	// Login with token if provided
	token, err := cmd.Cmd.Flags().GetString("token")
	if err != nil {
		return err
	}

	if token != "" {
		return h.ctrl.LoginWithToken(ctx, token)
	}
	return h.ctrl.LoginWithBrowser(ctx)
}

func LoginCmd(cmd *cobra.Command, args []string) error {
	h := New()
	ctx := context.Background()
	cmd.SilenceUsage = true
	return h.Login(ctx, &model.Command{
		Cmd:  cmd,
		Args: args,
	})
}
