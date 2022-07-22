package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/unweave/cli/model"
)

// Link links a project id to a given directory in the filesystem.
func (h *Handler) Link(ctx context.Context, cmd *model.Command) error {
	path := ""
	projectID := cmd.Args[0]
	if len(cmd.Args) > 1 {
		path = cmd.Args[1]
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		panic(err)
	}
	if _, err = os.Stat(absPath); os.IsNotExist(err) {
		fmt.Printf("Path %s does not exist\n", path)
		return err
	}

	return h.ctrl.Link(ctx, projectID, absPath)
}

func LinkCmd(cmd *cobra.Command, args []string) error {
	h := New()
	cmd.SilenceUsage = true
	return h.Link(cmd.Context(), &model.Command{
		Cmd:  cmd,
		Args: args,
	})
}
