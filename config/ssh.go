package config

import (
	"os"
	"path/filepath"

	"github.com/unweave/cli/ui"
)

func GetSSHKeysFolder() string {
	home, err := os.UserHomeDir()
	if err != nil {
		ui.Errorf("Unable to find home directory: %s", err)
		os.Exit(1)
	}
	dotSSHPath := filepath.Join(home, ".ssh")
	if _, err := os.Stat(dotSSHPath); os.IsNotExist(err) {
		if err := os.MkdirAll(dotSSHPath, 0700); err != nil {
			ui.Errorf(".ssh directory not found and attempt to create it failed: %s", err)
			os.Exit(1)
		}
	}
	return dotSSHPath
}
