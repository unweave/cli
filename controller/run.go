package controller

import (
	"context"
	"fmt"
	ignore "github.com/sabhiram/go-gitignore"
	"github.com/unweave/cli/entity"
	"github.com/unweave/cli/info"
	"github.com/unweave/cli/pkg/compress"
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
func (c *Controller) Run(ctx context.Context, command string) error {
	path, err := c.cfg.GetProjectPath()
	if err != nil {
		return err
	}

	pid, err := c.cfg.GetProjectIDFromPath(path)
	if err != nil {
		fmt.Println(info.ProjectNotFoundMsg())
		return err
	}

	zepl, err := c.api.CreateZepl(ctx, pid, command)
	if err != nil {
		return err
	}
	fmt.Printf("Created zepl '%s' with ID '%s'\n", zepl.Name, zepl.ID)

	// Walk the filesystem the repo root and zip up the files
	gatherFunc := gatherContext(path)
	if err = c.api.UploadZeplContext(ctx, zepl.ID, gatherFunc); err != nil {
		return err
	}

	if err = c.api.LaunchZepl(ctx, zepl.ID); err != nil {
		return err
	}

	// Logs to get logs
	if err = c.api.TailZeplLogs(ctx, zepl.ID); err != nil {
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
