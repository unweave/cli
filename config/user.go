package config

import "github.com/unweave/cli/entity"

func (c *Config) UpdateUserConfig(userConfig entity.UserConfig) error {
	c.Root.User = &entity.UserConfig{
		Id:    userConfig.Id,
		Token: userConfig.Token,
	}
	return MarshalAndWrite(c, c.Root)
}
