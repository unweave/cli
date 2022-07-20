package config

import "github.com/unweave/cli/model"

func (c *Config) UpdateUserConfig(userConfig model.UserConfig) error {
	c.Root.User = &model.UserConfig{
		Token: userConfig.Token,
	}
	return MarshalAndWrite(c, c.Root)
}
