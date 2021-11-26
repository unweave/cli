package config

import (
	"github.com/unweave/cli/entity"
	"os"
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

// ReadAndUnmarshal reads the cfg file and unmarshals it into the RootConfig struct
func ReadAndUnmarshal(config *Config, rc entity.RootConfig) error {
	if err := config.viper.ReadInConfig(); err != nil {
		return err
	}
	return config.viper.Unmarshal(&rc)
}

// MarshalAndWrite write marshals a RootConfig structs and writes it to disk
func MarshalAndWrite(config *Config, rc entity.RootConfig) error {
	rt := reflect.TypeOf(rc)
	fields := reflect.VisibleFields(reflect.TypeOf(rc))

	for _, field := range fields {
		config.viper.Set(field.Name, rt.FieldByIndex(field.Index))
	}

	if err := createDir(config.Path); err != nil {
		return err
	}
	return config.viper.WriteConfig()
}
