package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/unweave/cli/entity"
	"gopkg.in/gookit/color.v1"
	"os"
	"path/filepath"
)

func (h *Handler) Run(ctx context.Context, cmd *entity.Command) error {
	var path string
	relPath := "."
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	if len(cmd.Args) > 0 {
		relPath = cmd.Args[0]
		p, err := filepath.Abs(filepath.Join(pwd, relPath))
		if err != nil {
			panic(err)
		}
		path = filepath.Clean(p)
		if _, err = os.Stat(path); os.IsNotExist(err) {
			fmt.Printf("Path %s does not exist\n", path)
			return err
		}
	} else {
		path, err = h.cfg.GetActiveProjectDir()
		if err != nil {
			msg := "Ow snap! Looks like you don't have a currently active Unweave project. \n" +
				"Either switch to a unweave project folder or create a new one by running: \n" +
				color.Blue.Render("unweave init")
			fmt.Println(msg)
			return err
		}
	}

	return h.ctrl.Run(ctx, path)
}

func RunCmd(cmd *cobra.Command, args []string) error {
	h := New()
	ctx := context.Background()
	return h.Run(ctx, &entity.Command{
		Cmd:  cmd,
		Args: args,
	})
}
