package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/unweave/cli/model"
)

func createDir(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, os.ModePerm)
	} else if err != nil {
		return err
	}
	return nil
}

// ReadAndUnmarshal reads the config file and unmarshals it into the RootConfig struct
func ReadAndUnmarshal(config *Config, rc *model.RootConfig) error {
	buf, err := ioutil.ReadFile(config.Path)
	if err != nil {
		return err
	}
	return json.Unmarshal(buf, rc)
}

// MarshalAndWrite marshals a RootConfig struct and writes it to disk. It reloads the
// config variable after writing.
func MarshalAndWrite(config *Config, rc *model.RootConfig) error {
	if err := createDir(filepath.Dir(config.Path)); err != nil {
		return err
	}
	buf, err := json.MarshalIndent(rc, "", "  ")
	if err != nil {
		return err
	}

	if err = ioutil.WriteFile(config.Path, buf, os.ModePerm); err != nil {
		return err
	}
	return config.Reload()
}
