package config

import "github.com/unweave/cli/entity"

func (c *Config) AddProject(path string, config entity.ProjectConfig) error {
	c.Root.Projects[path] = config
	return MarshalAndWrite(c, c.Root)
}
