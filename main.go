package main

import (
	"fmt"
	"os"

	"github.com/muesli/reflow/wordwrap"
	"github.com/skratchdot/open-golang/open"
	"github.com/spf13/cobra"
	"github.com/unweave/cli/cmd"
	"github.com/unweave/cli/config"
	"github.com/unweave/cli/ui"
)

var (
	groupDev        = "dev"
	groupManagement = "management"

	rootCmd = &cobra.Command{
		Use:           "unweave <command>",
		Short:         "Create serverless sessions to train your ML models",
		Args:          cobra.MinimumNArgs(0),
		SilenceUsage:  false,
		SilenceErrors: false,
	}
)

func init() {
	rootCmd.Version = ""
	rootCmd.Flags().BoolP("version", "v", false, "Get the version of current Unweave CLI")
	rootCmd.AddGroup(&cobra.Group{ID: groupDev, Title: "Dev:"})
	rootCmd.AddGroup(&cobra.Group{ID: groupManagement, Title: "Account Management:"})

	flags := rootCmd.PersistentFlags()
	flags.StringVarP(&config.AuthToken, "token", "t", "", "Use a specific token to authenticate - overrides login token")
	flags.StringVarP(&config.ProjectPath, "path", "p", "", "ProjectPath to an Unweave project to run")

	rootCmd.AddCommand(&cobra.Command{
		Use:     "config",
		Short:   "Show the current config",
		GroupID: groupDev,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(config.Config.String())
		},
	})

	initCmd := &cobra.Command{
		Use:     "init [project-id]",
		Short:   "Init a new project config in your local directory and link it to an Unweave project",
		GroupID: groupManagement,
		Args:    cobra.ExactArgs(1),
		RunE:    cmd.Init,
	}
	initCmd.Flags().StringP("path", "p", "", "Path to the project directory")
	rootCmd.AddCommand(initCmd)

	// Auth
	loginCmd := &cobra.Command{
		Use:     "login",
		Short:   "Login to Unweave",
		GroupID: groupManagement,
		RunE:    cmd.Login,
	}
	rootCmd.AddCommand(loginCmd)

	rootCmd.AddCommand(&cobra.Command{
		Use:    "logout",
		Short:  "Logout of Unweave",
		RunE:   cmd.Logout,
		Hidden: true,
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "open",
		Short: "Open the Unweave dashboard in your browser",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := open.Run(config.Config.Unweave.AppURL); err != nil {
				ui.Errorf("Failed to open browser: %s", err)
				os.Exit(1)
			}
			return nil
		},
		Args: cobra.NoArgs,
	})

	// Provider commands
	lsNodeType := &cobra.Command{
		Use:   "ls-node-types <provider>",
		Short: "List node types available on a provider",
		Args:  cobra.ExactArgs(1),
		RunE:  cmd.ProviderListNodeTypes,
	}
	lsNodeType.Flags().BoolVarP(&config.All, "all", "a", false, "Including out of capacity node types")
	rootCmd.AddCommand(lsNodeType)

	// Session commands
	sessionCmd := &cobra.Command{
		Use:     "session",
		Short:   "Manage Unweave sessions: create | ls | terminate",
		GroupID: groupDev,
		Args:    cobra.NoArgs,
	}

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new Unweave session.",
		Long: wordwrap.String("Create a new Unweave session. If no region is provided,"+
			"the first available one will be selected.", ui.MaxOutputLineLength),
		Args: cobra.NoArgs,
		RunE: cmd.SessionCreateCmd,
	}
	createCmd.Flags().StringVar(&config.Provider, "provider", "", "Provider to use")
	createCmd.Flags().StringVar(&config.NodeTypeID, "type", "", "Node type to use, eg. `gpu_1x_a100`")
	createCmd.Flags().StringVar(&config.NodeRegion, "region", "", "Region to use, eg. `us_west_2`")
	createCmd.Flags().StringVarP(&config.SSHKeyName, "key", "k", "", "Name of the SSH key to use for the session")
	createCmd.Flags().StringVar(&config.SSHPublicKeyPath, "pub", "", "Path to the SSH public key to use")
	sessionCmd.AddCommand(createCmd)

	lsCmd := &cobra.Command{
		Use:   "ls",
		Short: "List active Unweave sessions",
		Long:  "List active Unweave sessions. To list all sessions, use the --all flag.",
		Args:  cobra.NoArgs,
		RunE:  cmd.SessionList,
	}
	lsCmd.Flags().BoolVarP(&config.All, "all", "a", false, "List all sessions")
	sessionCmd.AddCommand(lsCmd)

	sessionCmd.AddCommand(&cobra.Command{
		Use:   "terminate <session-id>",
		Short: "Terminate an Unweave session",
		Args:  cobra.ExactArgs(1),
		RunE:  cmd.SessionTerminate,
	})
	rootCmd.AddCommand(sessionCmd)

	// SSH Key commands
	sshKeyCmd := &cobra.Command{
		Use:     "ssh-keys",
		Short:   "Manage Unweave SSH keys: add | generate | ls",
		GroupID: groupDev,
		Args:    cobra.NoArgs,
	}
	sshKeyCmd.AddCommand(&cobra.Command{
		Use:   "add <public-key-path> [name]",
		Short: "Add a new SSH key to Unweave",
		Args:  cobra.RangeArgs(1, 2),
		RunE:  cmd.SSHKeyAdd,
	})
	sshKeyCmd.AddCommand(&cobra.Command{
		Use:   "generate [name]",
		Short: "Generate a new SSH key and add it to Unweave",
		Args:  cobra.RangeArgs(0, 1),
		RunE:  cmd.SSHKeyGenerate,
	})
	sshKeyCmd.AddCommand(&cobra.Command{
		Use:   "ls",
		Short: "List Unweave SSH keys",
		Args:  cobra.NoArgs,
		RunE:  cmd.SSHKeyList,
	})
	rootCmd.AddCommand(sshKeyCmd)
}

func main() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	defer func() {
		if r := recover(); r != nil {
			// TODO: Send telemetry
			ui.Errorf("Aw snap 😣 Something went wrong! %v", r)
			os.Exit(1)
		}
	}()

	config.Init()

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
