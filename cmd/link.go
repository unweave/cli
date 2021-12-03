package cmd

import (
	"context"
	"github.com/spf13/cobra"
	"github.com/unweave/cli/entity"
	"os"
	"path/filepath"
)

func (h *Handler) Link(ctx context.Context, cmd *entity.Command) error {
	relPath := "."
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	projectId := cmd.Args[0]
	if len(cmd.Args) > 1 {
		relPath = cmd.Args[1]
	}
	path, err := filepath.Abs(filepath.Join(pwd, relPath))
	if err != nil {
		panic(err)
	}

	return h.ctrl.Link(ctx, projectId, path)
}

func LinkCmd(cmd *cobra.Command, args []string) error {
	h := New()
	ctx := context.Background()
	return h.Link(ctx, &entity.Command{
		Cmd:  cmd,
		Args: args,
	})
}
