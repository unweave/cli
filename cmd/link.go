package cmd

import (
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/unweave/cli/config"
	"github.com/unweave/cli/ui"
)

func Link(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true
	ctx := cmd.Context()

	projectURI := args[0]
	parts := strings.Split(projectURI, "/")
	owner := config.Config.Unweave.User.ID

	if len(parts) != 2 {
		ui.Errorf("Invalid project URI: %q. Should be of type '<owner>/<project>", projectURI)
		os.Exit(1)
	}
	owner = parts[0]
	projectName := parts[1]

	uwc := InitUnweaveClient()
	project, err := uwc.Account.ProjectGet(ctx, owner, projectName)
	if err != nil {
		return ui.HandleError(err)
	}

	account, err := uwc.Account.AccountGet(ctx, config.Config.Unweave.User.ID)
	if err != nil {
		return ui.HandleError(err)
	}

	if config.IsProjectLinked() {
		ui.Errorf("Project already linked. Delete the %q directory to unlink.", config.ProjectConfigDirName)
		os.Exit(1)
	}

	if err = config.WriteProjectConfig(projectURI, account.Providers); err != nil {
		ui.Errorf("Failed to write project config: %s", err)
		os.Exit(1)
	}
	if err = config.WriteEnvConfig(); err != nil {
		ui.Errorf("Failed to write environment config: %s", err)
		os.Exit(1)
	}
	if err = config.WriteGitignore(); err != nil {
		ui.Errorf("Failed to write .gitignore: %s", err)
		os.Exit(1)
	}

	ui.Successf("Project linked: %s", project.Name)
	return nil
}
