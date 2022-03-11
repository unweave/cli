package entity

type RootConfig struct {
	User     *UserConfig              `json:"user"`
	Projects map[string]ProjectConfig `json:"projects"`
}

type ProjectConfig struct {
	Id string `json:"id"`
}

type UserConfig struct {
	Token string `json:"token"`
}

type ZeplConfig struct {
	InstanceType string `json:"instanceType"`
}
