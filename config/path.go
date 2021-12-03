package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// GetActiveProjectDir returns the active project directory by recursively going up the
// directory tree until it finds a directory that's configured inside the unweave root
// config file.
func (c *Config) GetActiveProjectDir() (string, error) {
	var activeProjectDir string
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	var walk func(path string)
	walk = func(path string) {
		if _, ok := c.Root.Projects[path]; ok {
			activeProjectDir = path
			return
		}
		parent := filepath.Dir(path)
		if parent == "." || parent == "/" {
			return
		}
		walk(parent)

	}
	walk(pwd)

	if activeProjectDir == "" {
		return "", fmt.Errorf("no active project found")
	}
	return activeProjectDir, nil
}
