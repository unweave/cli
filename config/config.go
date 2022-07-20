package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/unweave/cli/constants"
	"github.com/unweave/cli/model"
)

type Config struct {
	Path    string            `json:"path"`
	IsDebug bool              `json:"debug"`
	Api     *model.ApiConfig  `json:"api"`
	Root    *model.RootConfig `json:"root"`
	Zepl    *model.ZeplConfig `json:"zepl"`
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
	rootCfg := model.RootConfig{
		User:     &model.UserConfig{},
		Projects: make(map[string]model.ProjectConfig),
	}

	config := Config{
		Path:    getConfigPath(),
		IsDebug: false,
		Api: &model.ApiConfig{
			ApiUrl:       getApiUrl(),
			AppUrl:       getAppUrl(),
			GqlUrl:       getGqlUrl(),
			WorkbenchUrl: getWorkbenchUrl(),
		},
		Root: &rootCfg,
		Zepl: &model.ZeplConfig{
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
