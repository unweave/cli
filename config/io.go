package config

import (
	"github.com/unweave/cli/entity"
	"os"
	"path/filepath"
	"reflect"
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
func ReadAndUnmarshal(config *Config, rc *entity.RootConfig) error {
	if err := config.viper.ReadInConfig(); err != nil {
		return err
	}
	return config.viper.Unmarshal(rc)
}

// MarshalAndWrite marshals a RootConfig struct and writes it to disk
func MarshalAndWrite(config *Config, rc *entity.RootConfig) error {
	fields := reflect.ValueOf(*rc)
	for i := 0; i < fields.NumField(); i++ {
		k := fields.Type().Field(i).Name
		v := fields.Field(i).Interface()

		config.viper.Set(k, v)
	}

	if err := createDir(filepath.Dir(config.Path)); err != nil {
		return err
	}
	return config.viper.WriteConfig()
}
