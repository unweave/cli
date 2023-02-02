package config

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

func createDir(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, os.ModePerm)
	} else if err != nil {
		return err
	}
	return nil
}

func unmarshalDotEnv(path string, config interface{}) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	data := make(map[string]string)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "=")
		if len(parts) == 2 {
			data[parts[0]] = parts[1]
		}
	}

	configDataValue := reflect.ValueOf(config).Elem()
	configDataType := configDataValue.Type()

	for i := 0; i < configDataValue.NumField(); i++ {
		field := configDataType.Field(i)
		envKey, ok := field.Tag.Lookup("env")
		if ok {
			configDataValue.Field(i).SetString(data[envKey])
		}
	}

	return nil
}

// readAndUnmarshal reads the config file and unmarshals it into the config struct
func readAndUnmarshal[T any](path string, config *T) error {
	if filepath.Base(path) == ".env" {
		return unmarshalDotEnv(path, config)
	}

	buf, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if strings.HasSuffix(path, ".toml") {
		return toml.Unmarshal(buf, config)
	}
	if strings.HasSuffix(path, ".json") {
		return json.Unmarshal(buf, config)
	}
	return nil
}

// marshalAndWrite marshals a RootConfig struct and writes it to disk. It reloads the
// config variable after writing.
func marshalAndWrite[T any](path string, config *T) error {
	if err := createDir(filepath.Dir(path)); err != nil {
		return err
	}
	buf, err := toml.Marshal(config)
	if err != nil {
		return err
	}

	if err = os.WriteFile(path, buf, os.ModePerm); err != nil {
		return err
	}
	return nil
}
