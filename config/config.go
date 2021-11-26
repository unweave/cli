package config

import "github.com/spf13/viper"

type Config struct {
	viper *viper.Viper
}

func New() *Config {
	return &Config{
		viper: viper.New(),
	}
}
