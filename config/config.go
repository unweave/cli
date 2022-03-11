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
	Zepl    *entity.ZeplConfig
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

func (c *Config) GetUnweaveDomain() string {
	url := os.Getenv("UNWEAVE_DOMAIN")
	if url == "" {
		url = constants.UnweaveDomain
	}
	return url
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
	return "https://workbench." + c.GetUnweaveDomain()
}

func (c *Config) GetGqlUrl() string {
	return c.GetApiUrl() + "/graphql"
}

func getUnweaveEnv() string {
	env := os.Getenv("UNWEAVE_ENV")
	if env == "" {
		env = constants.UnweaveEnv
	}
	return env
}

func getConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	if env := getUnweaveEnv(); env == "production" {
		return filepath.Join(home, ".unweave", "config.json")
	}
	return filepath.Join(home, ".unweave", getUnweaveEnv()+"-config.json")

}

func New() *Config {
	it := "cpu"
	if UseGpu {
		it = "gpu"
	}

	// Init empty
	rootCfg := entity.RootConfig{
		User:     &entity.UserConfig{},
		Projects: make(map[string]entity.ProjectConfig),
	}

	config := Config{
		Root: &rootCfg,
		Path: getConfigPath(),
		Zepl: &entity.ZeplConfig{
			InstanceType: it,
		},
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
