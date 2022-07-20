package config

import "github.com/unweave/cli/model"

func (c *Config) AddProject(path string, config model.ProjectConfig) error {
	c.Root.Projects[path] = config
	return MarshalAndWrite(c, c.Root)
}
