package config

import (
	"encoding/json"
	"fmt"
	"github.com/unweave/cli/constants"
	"github.com/unweave/cli/entity"
	"os"
)

type Config struct {
	Path    string             `json:"path"`
	IsDebug bool               `json:"debug"`
	Api     *entity.ApiConfig  `json:"api"`
	Root    *entity.RootConfig `json:"root"`
	Zepl    *entity.ZeplConfig `json:"zepl"`
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

func (c *Config) Reload() error {
	if err := ReadAndUnmarshal(c, c.Root); err != nil {
		return err
	}
	return nil
}

func (c *Config) ToJson() ([]byte, error) {
	buf, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func New() *Config {
	// Init empty
	rootCfg := entity.RootConfig{
		User:     &entity.UserConfig{},
		Projects: make(map[string]entity.ProjectConfig),
	}

	config := Config{
		Path:    getConfigPath(),
		IsDebug: false,
		Api: &entity.ApiConfig{
			ApiUrl:        getApiUrl(),
			AppUrl:        getAppUrl(),
			UnweaveDomain: getUnweaveDomain(),
			GqlUrl:        getGqlUrl(),
			WorkbenchUrl:  getWorkbenchUrl(),
		},
		Root: &rootCfg,
		Zepl: &entity.ZeplConfig{
			IsGpu: IsGpu,
		},
	}

	// Create the empty config if it doesn't exist
	if err := ReadAndUnmarshal(&config, &rootCfg); os.IsNotExist(err) {
		err = MarshalAndWrite(&config, &rootCfg)
		if err != nil {
			panic(err)
		}
	} else if err != nil {
		fmt.Printf("Failed to read config file: %s\n", err)
		panic(err)
	}

	// Override auth token if set manually at runtime
	if constants.AuthToken != "" {
		config.Root.User.Token = constants.AuthToken
	}
	return &config
}
