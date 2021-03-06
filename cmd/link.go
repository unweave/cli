package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/unweave/cli/model"
)

func (h *Handler) Link(ctx context.Context, cmd *model.Command) error {
	relPath := "."
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	projectID := cmd.Args[0]
	if len(cmd.Args) > 1 {
		relPath = cmd.Args[1]
	}
	path, err := filepath.Abs(filepath.Join(pwd, relPath))
	if err != nil {
		panic(err)
	}
	if _, err = os.Stat(path); os.IsNotExist(err) {
		fmt.Printf("Path %s does not exist\n", path)
		return err
	}

	return h.ctrl.Link(ctx, projectID, path)
}

func LinkCmd(cmd *cobra.Command, args []string) error {
	h := New()
	ctx := context.Background()
	cmd.SilenceUsage = true
	return h.Link(ctx, &model.Command{
		Cmd:  cmd,
		Args: args,
	})
}
