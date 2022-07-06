package config

import (
	"os"
	"path/filepath"

	"github.com/unweave/cli/constants"
)

func getApiUrl() string {
	url := os.Getenv("UNWEAVE_API_URL")
	if url == "" {
		url = constants.UnweaveApiUrl
	}
	return url
}

func getAppUrl() string {
	url := os.Getenv("UNWEAVE_APP_URL")
	if url == "" {
		url = constants.UnweaveAppUrl
	}
	return url
}

func getConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	if env := getUnweaveEnv(); env == "production" {
		return filepath.Join(home, ".unweave", "config.json")
	}
	return filepath.Join(home, ".unweave", getUnweaveEnv()+"-config.json")

}

func getGqlUrl() string {
	return getApiUrl() + "/graphql"
}

func getWorkbenchUrl() string {
	url := os.Getenv("UNWEAVE_WORKBENCH_URL")
	if url == "" {
		url = constants.UnweaveWorkbenchUrl
	}
	return url
}

func getUnweaveEnv() string {
	env := os.Getenv("UNWEAVE_ENV")
	if env == "" {
		env = constants.UnweaveEnv
	}
	return env
}
