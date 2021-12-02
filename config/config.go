package config

import (
	"fmt"
	"github.com/spf13/viper"
	"github.com/unweave/cli/entity"
	"os"
	"path/filepath"
)

type Config struct {
	viper   *viper.Viper
	Root    *entity.RootConfig
	Path    string
	IsDev   bool
	IsDebug bool
}

func (c *Config) Reload() error {
	c.viper.SetConfigFile(c.Path)
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
	return "http://localhost:4000"
}

func (c *Config) GetAppUrl() string {
	return "http://localhost:3000"
}

func (c *Config) GetGqlUrl() string {
	return c.GetApiUrl() + "/"
}

func (c *Config) GetRestUrl() string {
	return "http://localhost:8000/api"
}

func New() *Config {
	cfgViper := viper.New()
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	path := filepath.Join(home, ".unweave", "config.json")
	cfgViper.SetConfigFile(path)

	// Init empty
	rootCfg := entity.RootConfig{}
	config := Config{
		viper:   cfgViper,
		Root:    &rootCfg,
		Path:    path,
		IsDev:   os.Getenv("UNWEAVE_ENV") == "dev",
		IsDebug: os.Getenv("UNWEAVE_DEBUG") == "true",
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
