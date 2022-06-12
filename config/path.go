package config

import (
	"fmt"
	"github.com/unweave/cli/info"
	"os"
	"path/filepath"
)

func getAbsPath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}

	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	p, err := filepath.Abs(filepath.Join(pwd, path))
	if err != nil {
		panic(err)
	}
	return p
}

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

func (c *Config) GetProjectIDFromPath(projectDir string) (string, error) {
	cfg, ok := c.Root.Projects[projectDir]
	if !ok {
		return "", fmt.Errorf("no active project found")
	}
	return cfg.ID, nil
}

// GetProjectPath parses the path to the project directory used as context for the
// current execution. It checks to see if the project path was passed as a flag argument.
// If not, it checks to see if the current directory or a parent is the project directory.
// It returns the absolute path to the project directory.
func (c *Config) GetProjectPath() (string, error) {
	if ZeplProjectPath != "" {
		p := getAbsPath(ZeplProjectPath)
		path := filepath.Clean(p)

		if _, err := os.Stat(path); os.IsNotExist(err) {
			fmt.Printf("Path %s does not exist\n", path)
			return "", err
		}
		return path, nil
	}

	path, err := c.GetActiveProjectDir()
	if err != nil {
		fmt.Println(info.ProjectNotFoundMsg())
		return "", err
	}

	return path, nil
}
