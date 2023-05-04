package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	ignore "github.com/sabhiram/go-gitignore"
	"github.com/spf13/cobra"
	"github.com/unweave/cli/config"
	"github.com/unweave/cli/tools"
	"github.com/unweave/cli/ui"
	"github.com/unweave/unweave/api/types"
)

const defaultGitIgnore = `
.git
**/.DS_Store
`

type gatherContextFunc func(w io.Writer) error

// gatherContext zips up the user's code and environment and write it to a buffer to be
// uploaded to the server.
func gatherContext(rootDir string, w io.Writer, archiveType string) error {
	giPath := filepath.Join(rootDir, ".gitignore")
	lines := strings.Split(defaultGitIgnore, "\n")

	// Compile ignore pattern - use GitIgnore if it exists
	var gi *ignore.GitIgnore
	if _, err := os.Stat(giPath); os.IsNotExist(err) {
		gi = ignore.CompileIgnoreLines(lines...)
	} else {
		gi, err = ignore.CompileIgnoreFileAndLines(giPath, lines...)
		if err != nil {
			ui.Errorf("Error compiling .gitignore file:", err)
			ui.Errorf("Ignoring .gitignore file")
		}
	}

	if archiveType == "zip" {
		return tools.Zip(rootDir, w, gi)
	}
	return tools.Tar(rootDir, w, gi)
}

func Build(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	dir := ""
	if len(args) > 0 {
		dir = args[0]
	} else {
		var err error
		dir, err = config.GetActiveProjectPath()
		if err != nil {
			ui.Errorf("Couldn't get active project path. Make sure you're in a project " +
				"directory or supply a path: \n" + err.Error())
			os.Exit(1)
		}
	}

	if s, err := os.Stat(dir); err != nil || !s.IsDir() {
		ui.Errorf("Couldn't find directory %q", dir)
		os.Exit(1)
	}

	uwc := InitUnweaveClient()
	buf := &bytes.Buffer{}

	owner, projectName := config.GetProjectOwnerAndName()

	if err := gatherContext(dir, buf, "zip"); err != nil {
		return err
	}
	ui.Infof("Starting build for project '%s/%s'", owner, projectName)

	params := types.BuildsCreateParams{
		Builder:      "docker",
		BuildContext: io.NopCloser(buf),
	}
	buildID, err := uwc.Build.Create(cmd.Context(), owner, projectName, params)
	if err != nil {
		var e *types.Error
		if errors.As(err, &e) {
			uie := &ui.Error{Error: e}
			fmt.Println(uie.Verbose())
			os.Exit(1)
		}
	}
	ui.Successf("Build %q is under way!", buildID)
	return nil
}
