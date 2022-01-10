package config

import (
	"fmt"
	"github.com/unweave/cli/constants"
	"github.com/unweave/cli/entity"
	"os"
	"path/filepath"
)

type Config struct {
	Root    *entity.RootConfig
	Path    string
	IsDebug bool
}

func (c *Config) Reload() error {
	if err := ReadAndUnmarshal(c, c.Root); err != nil {
		return err
	}
	return nil
}

func (c *Config) IsLoggedIn() (bool, error) {
	if err := c.Reload(); err != nil {
		return false, err
	}
	if c.Root.User == nil || c.Root.User.Token == "" {
		return false, nil
	}
	return true, nil
}

func (c *Config) GetApiUrl() string {
	url := os.Getenv("UNWEAVE_API_URL")
	if url == "" {
		url = constants.UnweaveApiUrl
	}
	return url
}

func (c *Config) GetAppUrl() string {
	url := os.Getenv("UNWEAVE_APP_URL")
	if url == "" {
		url = constants.UnweaveAppUrl
	}
	return url
}

func (c *Config) GetWorkbenchUrl() string {
	return c.GetApiUrl() + "/workbench"
}

func (c *Config) GetGqlUrl() string {
	return c.GetApiUrl() + "/graphql"
}

func New() *Config {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	path := filepath.Join(home, ".unweave", "config.json")

	// Init empty
	rootCfg := entity.RootConfig{
		User:     &entity.UserConfig{},
		Projects: make(map[string]entity.ProjectConfig),
	}
	config := Config{
		Root: &rootCfg,
		Path: path,
	}

	// Create the empty config if it doesn't exist
	if err := ReadAndUnmarshal(&config, &rootCfg); os.IsNotExist(err) {
		err = MarshalAndWrite(&config, &rootCfg)
		if err != nil {
			panic(err)
		}
	} else if err != nil {
		fmt.Printf("Fialed to read config file: %s\n", err)
		panic(err)
	}

	return &config
}
