package config

import (
	"fmt"
	"github.com/spf13/viper"
	"github.com/unweave/cli/entity"
	"os"
	"path/filepath"
)

type Config struct {
	viper *viper.Viper
	Root  *entity.RootConfig
	Path  string
}

func New() *Config {
	cfgViper := viper.New()
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	path := filepath.Join(home, ".unweave", "config.json")
	cfgViper.SetConfigFile(path)
	config := &Config{
		viper: cfgViper,
		Root:  nil,
		Path:  path,
	}

	// Create the empty config if it doesn't exist
	if err := viper.ReadInConfig(); os.IsNotExist(err) {
		err = MarshalAndWrite(config, entity.RootConfig{
			User:     nil,
			Projects: nil,
		})
		if err != nil {
			panic(err)
		}
	} else if err != nil {
		fmt.Printf("Fialed to read config file: %s\n", err)
		panic(err)
	}

	return config
}
