package entity

type RootConfig struct {
	User     *UserConfig              `json:"user"`
	Projects map[string]ProjectConfig `json:"projects"`
}

type ProjectConfig struct {
	Id   string `json:"id"`
	Path string `json:"path"`
}

type UserConfig struct {
	Id    string `json:"id"`
	Token string `json:"token"`
}
