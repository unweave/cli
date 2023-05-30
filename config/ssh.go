package config

import (
	"os"
	"path/filepath"

	"github.com/unweave/cli/ui"
)

func GetUnweaveSSHKeysFolder() string {
	home, err := os.UserHomeDir()
	if err != nil {
		ui.Errorf("Unable to find home directory: %s", err)
		os.Exit(1)
	}
	path := filepath.Join(home, UnweaveSSHKeysDir)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0700); err != nil {
			ui.Errorf("Unable to create %s directory: %s", UnweaveSSHKeysDir, err)
			os.Exit(1)
		}
	}
	return path
}
