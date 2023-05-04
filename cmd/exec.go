package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/unweave/cli/config"
	"github.com/unweave/cli/ui"
	"github.com/unweave/unweave/api/types"
)

func gitIsStatusClean() (bool, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return false, fmt.Errorf("failed to get git status: %s %s", stderr.String(), stdout.String())
	}
	return strings.TrimSpace(stdout.String()) == "", nil
}

func getGitRemoteURL() (string, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("git", "config", "--get", "remote.origin.url")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to get git remote URL: %s", stderr.String())
	}
	url := strings.TrimSuffix(strings.TrimSpace(stdout.String()), ".git")
	return url, nil
}

func gitAdd() error {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("git", "add", ".")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add: %s %s", stderr.String(), stdout.String())
	}
	return nil
}

func gitCommit() (string, error) {
	var stdout, stderr bytes.Buffer

	clean, err := gitIsStatusClean()
	if err != nil {
		return "", err
	}

	if !clean {
		cmd := exec.Command("git", "commit", "-m", "uw:auto-commit") // TODO: add commit message
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("failed to commit: %s %s", stderr.String(), stdout.String())
		}
	}

	// Get Commit ID
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to get commit ID: %s %s", stderr.String(), stdout.String())
	}
	return strings.TrimSpace(stdout.String()), nil
}

func gitPush() error {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("git", "push")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to push: %s %s", stderr.String(), stdout.String())
	}
	return nil
}

func Exec(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	dir, err := config.GetActiveProjectPath()
	if err != nil {
		ui.Errorf("Couldn't get active project path. Make sure you're in a project " +
			"directory or supply a path: \n" + err.Error())
		os.Exit(1)
	}

	if s, err := os.Stat(dir); err != nil || !s.IsDir() {
		ui.Errorf("Couldn't find directory %q", dir)
		os.Exit(1)
	}

	ui.Infof("Gathering context from %q", dir)
	buf := &bytes.Buffer{}
	if err := gatherContext(dir, buf, "zip"); err != nil {
		return err
	}

	// check config to see if the user has auto-commit enabled
	// if auto-commit is set to ask, ask the user if they want to commit
	// if yes, get git remote url, commit and push

	gitRemote, err := getGitRemoteURL()
	if err != nil {
		return err
	}
	if err = gitAdd(); err != nil {
		return err
	}

	commitID, err := gitCommit()
	if err != nil {
		return err
	}
	if err = gitPush(); err != nil {
		return err
	}
	commitURL := fmt.Sprintf("%s/commit/%s", gitRemote, commitID)
	ui.Infof("Commit URL: %s", commitURL)

	execConfig := types.ExecConfig{
		Image:   "",
		Command: args,
		Keys:    nil,
		Volumes: nil,
		Src: &types.SourceContext{
			MountPath: "/home/ubuntu",
			Context:   io.NopCloser(buf),
		},
	}

	gitConfig := types.GitConfig{
		CommitID: &commitID,
		GitURL:   &gitRemote,
	}

	_, err = sessionCreate(cmd.Context(), execConfig, gitConfig)
	if err != nil {
		return err
	}

	// TODO: subscribe to logs
	return nil
}
