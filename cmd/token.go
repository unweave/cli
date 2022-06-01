package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/unweave/cli/entity"
)

func (h *Handler) CreateUserToken(ctx context.Context, cmd *entity.Command) error {
	token, err := h.ctrl.CreateUserToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to create new user token, %s", err)
	}

	fmt.Println("Created new user token:")
	fmt.Printf("\t%s\n", token)
	return nil
}

func (h *Handler) GetUserTokens(ctx context.Context, cmd *entity.Command) error {
	tokens, err := h.ctrl.GetUserTokens(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch user tokens, %s", err)
	}

	for _, token := range tokens {
		fmt.Printf("\t%s\n", token)
	}
	return nil
}

func CreateUserTokenCmd(cmd *cobra.Command, args []string) error {
	h := New()
	ctx := context.Background()
	cmd.SilenceUsage = true
	return h.CreateUserToken(ctx, &entity.Command{
		Args: args,
		Cmd:  cmd,
	})
}

func GetUserTokensCmd(cmd *cobra.Command, args []string) error {
	h := New()
	ctx := context.Background()
	cmd.SilenceUsage = true
	return h.GetUserTokens(ctx, &entity.Command{
		Args: args,
		Cmd:  cmd,
	})
}
