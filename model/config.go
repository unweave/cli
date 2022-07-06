package model

type ApiConfig struct {
	ApiUrl       string `json:"unweaveApiUrl"`
	AppUrl       string `json:"unweaveAppUrl"`
	GqlUrl       string `json:"unweaveGqlUrl"`
	WorkbenchUrl string `json:"unweaveWorkbenchUrl"`
}

type ProjectConfig struct {
	ID string `json:"id"`
}

type RootConfig struct {
	User     *UserConfig              `json:"user"`
	Projects map[string]ProjectConfig `json:"projects"`
}

type UserConfig struct {
	Token string `json:"token"`
}

type ZeplConfig struct {
	IsGpu bool `json:"IsGpu"`
}
