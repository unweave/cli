package controller

import (
	"context"
	"fmt"
	ignore "github.com/sabhiram/go-gitignore"
	"github.com/unweave/cli/entity"
	"github.com/unweave/cli/pkg/compress"
	"gopkg.in/gookit/color.v1"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const defaultGitIgnore = `
.git
**/.DS_Store
`

// Run runs the user's latest changes and environment with Unweave. It uploads the users
// code to the server and runs it. Any files/patterns in the .gitignore file will are
// from the upload.
func (c *Controller) Run(ctx context.Context, path string) error {
	pid, err := c.cfg.GetProjectIdFromPath(path)
	if err != nil {
		msg := "Ow snap! Looks like you don't have a currently active Unweave project. \n" +
			"Either switch to a unweave project folder or create a new one by running: \n" +
			color.Blue.Render("unweave init")
		fmt.Println(msg)
		return err
	}

	rid, err := c.api.CreateZepl(ctx, pid)
	if err != nil {
		return err
	}
	fmt.Println("Created zepl:", rid)

	// Walk the filesystem the repo root and zip up the files
	gatherFunc := gatherContext(path)
	if err = c.api.UploadZeplContext(ctx, pid, rid, gatherFunc); err != nil {
		return err
	}

	// Connect to get logs
	if err = c.api.ConnectToZepl(ctx, pid, rid); err != nil {
		return err
	}
	return nil
}

// gatherContext zips up the user's code and environment and write it to a buffer to be
// uploaded to the server.
func gatherContext(rootDir string) entity.GatherContextFunc {
	giPath := filepath.Join(rootDir, ".gitignore")
	lines := strings.Split(defaultGitIgnore, "\n")

	// Compile ignore pattern - use GitIgnore if it exists
	var gi *ignore.GitIgnore
	if _, err := os.Stat(giPath); os.IsNotExist(err) {
		gi = ignore.CompileIgnoreLines(lines...)
	} else {
		gi, err = ignore.CompileIgnoreFileAndLines(giPath, lines...)
		if err != nil {
			fmt.Println("Error compiling .gitignore file:", err)
			fmt.Println("Ignoring .gitignore file")
		}
	}
	return func(w io.Writer) error {
		return compress.Zip(rootDir, w, gi)
	}
}
