package config

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/unweave/cli/ui"
	"github.com/unweave/unweave/tools/gonfig"
)

var (
	// Version is added at build time
	Version = "dev"

	//go:embed templates/config.toml
	configEmbed       string
	configTemplate, _ = template.New("config.toml").Parse(configEmbed)

	//go:embed templates/env
	envEmbed       string
	envTemplate, _ = template.New("env").Parse(envEmbed)

	//go:embed templates/gitignore
	gitignoreEmbed string

	GlobalConfigDirName  = ".unweave_global"
	ProjectConfigDirName = ".unweave"
	unweaveConfigPath    = ""
	projectConfigPath    = ProjectConfigDirName + "/config.toml"
	envConfigPath        = ProjectConfigDirName + "/.env"

	Config = &config{
		Unweave: &unweave{
			ApiURL: "https://api.unweave.io",
			AppURL: "https://app.unweave.io",
			User:   &user{},
		},
		Project: &project{
			Env:       &secrets{},
			Providers: map[string]provider{},
		},
	}
)

func GetProjectOwnerAndName() (string, string) {
	uri := Config.Project.URI
	parts := strings.Split(uri, "/")

	if len(parts) != 2 {
		ui.Errorf("Invalid project URI: %q. Should be of type '<owner>/<project>", uri)
		os.Exit(1)
	}

	owner := strings.Split(uri, "/")[0]
	name := strings.Split(uri, "/")[1]
	return owner, name
}

// GetActiveProjectPath returns the active project directory by recursively going up the
// directory tree until it finds a directory that's contains the .unweave/config.toml file
func GetActiveProjectPath() (string, error) {
	var activeProjectDir string
	pwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	var walk func(path string)
	walk = func(path string) {
		cfgDir := filepath.Join(path, ProjectConfigDirName)
		if _, err = os.Stat(cfgDir); err == nil {
			if _, err = os.Stat(filepath.Join(cfgDir, "config.toml")); err == nil {
				activeProjectDir = path
				return
			}
		}

		parent := filepath.Dir(path)
		if parent == "." || parent == "/" {
			return
		}
		walk(parent)
	}
	walk(pwd)

	if activeProjectDir == "" {
		return "", fmt.Errorf("no active project found")
	}
	return activeProjectDir, nil
}

func GetGlobalConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		ui.Errorf("Could not get user home directory")
		os.Exit(1)
	}
	return filepath.Join(home, GlobalConfigDirName)
}

func Init() {
	// ----- ProjectConfig -----
	envConfig := &secrets{}
	projectConfig := &project{}
	projectDir, err := GetActiveProjectPath()
	if err == nil {
		projectConfigPath = filepath.Join(projectDir, projectConfigPath)
		envConfigPath = filepath.Join(projectDir, envConfigPath)

		if err = readAndUnmarshal(projectConfigPath, &projectConfig); err != nil {
			ui.Infof("Failed to read project config at path %q", projectConfigPath)
		}
		if err = readAndUnmarshal(envConfigPath, envConfig); err != nil {
			ui.Infof("Failed to read environment config at path %q", envConfigPath)
		}
	}
	projectConfig.Env = envConfig

	// ----- Unweave Config -----

	env := "production"
	if e, ok := os.LookupEnv("UNWEAVE_ENV"); ok {
		env = e
	}
	apiURL := Config.Unweave.ApiURL
	appURL := Config.Unweave.AppURL

	switch env {
	case "staging", "stg":
		unweaveConfigPath = filepath.Join(GetGlobalConfigPath(), "stg-config.toml")
		apiURL = "https://api.staging-unweave.io"
		appURL = "https://app.staging-unweave.io"
	case "development", "dev":
		unweaveConfigPath = filepath.Join(GetGlobalConfigPath(), "dev-config.toml")
		apiURL = "http://localhost:4000"
		appURL = "http://localhost:3000"
	case "production", "prod":
		unweaveConfigPath = filepath.Join(GetGlobalConfigPath(), "config.toml")
	default:
		// If anything else, assume production
		fmt.Println("Unrecognized environment. Assuming production.")
	}

	// Load saved config - create the empty config if it doesn't exist
	if err = readAndUnmarshal(unweaveConfigPath, Config.Unweave); os.IsNotExist(err) {
		err = Config.Unweave.Save()
		if err != nil {
			ui.Errorf("Failed to create config file: %v", err)
		}
	} else if err != nil {
		ui.Errorf("Failed to read config file: %v", err)
	}

	// Need to set these after reading the config file so that they can be overridden
	Config.Unweave.ApiURL = apiURL
	Config.Unweave.AppURL = appURL
	Config.Project = projectConfig

	// Override with environment variables
	gonfig.GetFromEnvVariables(Config)
}
