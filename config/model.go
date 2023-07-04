package config

import (
	_ "embed"

	"github.com/pelletier/go-toml/v2"
	"github.com/unweave/cli/ui"
)

type (
	user struct {
		ID    string `toml:"id"`
		Email string `toml:"email"`
		Token string `toml:"token"`
	}

	Secrets struct {
		ProjectToken string `env:"UNWEAVE_PROJECT_TOKEN"`
		SSHKeyPath   string `env:"UNWEAVE_SSH_KEY_PATH"`
		SSHKeyName   string `env:"UNWEAVE_SSH_KEY_NAME"`
	}

	provider struct {
	}

	Spec struct {
		Name string        `toml:"name"`
		CPU  specResources `toml:"cpu"`
		GPU  specResources `toml:"gpu"`
		HDD  specHDD       `toml:"hdd"`
	}

	specResources struct {
		Type   string `toml:"type"`
		Count  int    `toml:"count"`
		Memory int    `toml:"memory"`
	}

	specHDD struct {
		Size int `toml:"size"`
	}

	Project struct {
		URI             string              `toml:"project_uri"`
		Env             *Secrets            `toml:"env"`
		Providers       map[string]provider `toml:"provider"`
		Specs           []Spec              `toml:"specs"`
		DefaultProvider string              `toml:"default_provider"`
	}

	unweave struct {
		UnwEnv string `toml:"unweave_env" env:"UNWEAVE_ENV"`
		ApiURL string `toml:"api_url" env:"UNWEAVE_API_URL"`
		AppURL string `toml:"app_url" env:"UNWEAVE_APP_URL"`
		User   *user  `toml:"user"`
	}

	config struct {
		Unweave *unweave `toml:"unweave"`
		Project *Project `toml:"project"`
	}
)

func (c *config) String() string {
	buf, err := toml.Marshal(c)
	if err != nil {
		ui.Errorf("Failed to marshal config: %s", err)
	}
	return string(buf)
}

func (c *unweave) Save() error {
	return marshalAndWrite(unweaveConfigPath, c)
}

func (c *Project) String() string {
	buf, err := toml.Marshal(c)
	if err != nil {
		ui.Errorf("Failed to marshal config: %s", err)
	}
	return string(buf)
}

func (c *Project) Save() error {
	return marshalAndWrite(projectConfigPath, c)
}
