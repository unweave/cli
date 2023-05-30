package config

import (
	"path/filepath"

	"github.com/unweave/cli/ui"
)

// UnweaveHostDir is the filesystem location where your code gets stored on the host. Adjust this as you need.
const UnweaveHostDir = "/home/unweave"

// ProjectHostDir is the location where project files get copied to
func ProjectHostDir() string {
	absPath, err := filepath.Abs(".")
	if err != nil {
		ui.HandleError(err)
	}

	_, rootDir := filepath.Split(absPath)

	return filepath.Join(UnweaveHostDir, rootDir)
}
