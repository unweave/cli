package config

import (
	"bytes"
	_ "embed"
	"os"
	"path/filepath"
)

func IsProjectLinked() bool {
	_, err := os.Stat(projectConfigPath)
	return err == nil
}

func WriteProjectConfig(projectURI string, providers []string) error {
	buf := &bytes.Buffer{}

	vars := struct {
		ProjectURI string
		Providers  []struct {
			Name string
		}
	}{
		ProjectURI: projectURI,
		Providers:  []struct{ Name string }{},
	}

	for _, p := range providers {
		// We already have the unweave config in the template.
		if p == "unweave" {
			continue
		}
		vars.Providers = append(vars.Providers, struct{ Name string }{Name: p})
	}

	if err := configTemplate.Execute(buf, vars); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(projectConfigPath), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(projectConfigPath, buf.Bytes(), 0644); err != nil {
		return err
	}
	return nil
}

func WriteEnvConfig() error {
	buf := &bytes.Buffer{}
	if err := envTemplate.Execute(buf, Config.Project.Env); err != nil {
		return err
	}
	dir := filepath.Dir(envConfigPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	if err := os.WriteFile(envConfigPath, buf.Bytes(), 0644); err != nil {
		return err
	}
	return nil
}

func WriteGitignore() error {
	dir := filepath.Dir(projectConfigPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	path := filepath.Join(dir, ".gitignore")
	if err := os.WriteFile(path, []byte(gitignoreEmbed), 0644); err != nil {
		return err
	}
	return nil
}
