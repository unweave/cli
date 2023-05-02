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
	"github.com/unweave/cli/vars"
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

type RunE func(cmd *cobra.Command, args []string) error

func withValidProjectURI(r RunE) RunE {
	return func(cmd *cobra.Command, args []string) error {
		if config.ProjectURI == "" && config.Config.Project.URI == "" {
			ui.Errorf("No project ID set. Either run `unweave link` first or use the `--project` flag to set a project ID.")
			os.Exit(1)
		}
		if config.ProjectURI != "" {
			// Override project ID if set via flag
			config.Config.Project.URI = config.ProjectURI
		}
		return r(cmd, args)
	}
}

func init() {
	rootCmd.Version = config.Version
	rootCmd.Flags().BoolP("version", "v", false, "Get the version of current Unweave CLI")
	rootCmd.AddGroup(&cobra.Group{ID: groupDev, Title: "Dev:"})
	rootCmd.AddGroup(&cobra.Group{ID: groupManagement, Title: "Account Management:"})

	flags := rootCmd.PersistentFlags()
	flags.StringVar(&config.ProjectURI, "project", "", "Use a specific project ID - overrides config")
	flags.StringVarP(&config.AuthToken, "token", "t", "", "Use a specific token to authenticate - overrides login token")
	flags.BoolVar(&vars.Debug, "debug", false, "Enable debug mode")

	rootCmd.AddCommand(&cobra.Command{
		Use:     "build [path]",
		Short:   "Build a project into a container image",
		GroupID: groupDev,
		Args:    cobra.RangeArgs(0, 1),
		RunE:    withValidProjectURI(cmd.Build),
		Hidden:  true,
	})

	boxCmd := &cobra.Command{
		Use:     "box [box-name]",
		Short:   "Create a new session with a persistent filesystem",
		GroupID: groupDev,
		RunE:    withValidProjectURI(cmd.BoxUp),
		Args:    cobra.RangeArgs(0, 1),
		Hidden:  true,
	}
	boxCmd.Flags().StringVar(&config.Provider, "provider", "", "Provider to use")
	boxCmd.Flags().StringVar(&config.NodeTypeID, "type", "", "Node type to use, eg. `gpu_1x_a100`")
	boxCmd.Flags().StringVar(&config.NodeRegion, "region", "", "Region to use, eg. `us_west_2`")
	boxCmd.Flags().StringVarP(&config.SSHKeyName, "key", "k", "", "Name of the SSH key to use for the session")
	boxCmd.Flags().StringVar(&config.SSHPublicKeyPath, "pub", "", "Path to the SSH public key to use")

	rootCmd.AddCommand(boxCmd)

	codeCmd := &cobra.Command{
		Use:     "code",
		Short:   "Create a new session and open it in VS Code",
		GroupID: groupDev,
		Hidden:  true,
		RunE:    withValidProjectURI(cmd.Code),
	}
	codeCmd.Flags().StringVarP(&config.BuildID, "image", "i", "", "Build ID of the container image to use")
	codeCmd.Flags().StringVar(&config.Provider, "provider", "", "Provider to use")
	codeCmd.Flags().StringVar(&config.NodeTypeID, "type", "", "Node type to use, eg. `gpu_1x_a100`")
	codeCmd.Flags().StringVar(&config.NodeRegion, "region", "", "Region to use, eg. `us_west_2`")
	rootCmd.AddCommand(codeCmd)

	rootCmd.AddCommand(&cobra.Command{
		Use:     "config",
		Short:   "Show the current config",
		GroupID: groupDev,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(config.Config.String())
		},
	})
	rootCmd.AddCommand(&cobra.Command{
		Use:     "exec",
		Short:   "Execute a command serverlessly",
		GroupID: groupDev,
		Hidden:  true,
		RunE:    withValidProjectURI(cmd.Exec),
	})
	linkCmd := &cobra.Command{
		Use:     "link [project-id]",
		Aliases: []string{"init"}, // this is temp
		Short:   "Init a new project config in your local directory and link it to an Unweave project",
		GroupID: groupManagement,
		Args:    cobra.ExactArgs(1),
		RunE:    cmd.Link,
	}
	linkCmd.Flags().StringP("path", "p", "", "Path to the project directory")
	rootCmd.AddCommand(linkCmd)

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

	// Provider commands
	lsNodeType := &cobra.Command{
		Use:   "ls-node-types <provider>",
		Short: "List node types available on a provider",
		Args:  cobra.ExactArgs(1),
		RunE:  cmd.ProviderListNodeTypes,
	}
	lsNodeType.Flags().BoolVarP(&config.All, "all", "a", false, "Including out of capacity node types")
	rootCmd.AddCommand(lsNodeType)

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

	// Session commands
	sessionCmd := &cobra.Command{
		Use:     "sessions",
		Aliases: []string{"session", "sess"},
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
		RunE: withValidProjectURI(cmd.SessionCreateCmd),
	}
	createCmd.Flags().StringVarP(&config.BuildID, "image", "i", "", "Build ID of the container image to use")
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
		RunE:  withValidProjectURI(cmd.SessionList),
	}
	lsCmd.Flags().BoolVarP(&config.All, "all", "a", false, "List all sessions")
	sessionCmd.AddCommand(lsCmd)

	sessionCmd.AddCommand(&cobra.Command{
		Use:   "terminate [session-id]",
		Short: "Terminate an Unweave session",
		Args:  cobra.RangeArgs(0, 1),
		RunE:  withValidProjectURI(cmd.SessionTerminate),
	})
	rootCmd.AddCommand(sessionCmd)

	sshCmd := &cobra.Command{
		Use:   "ssh [session-name|id]",
		Short: "SSH into existing session or create a new one",
		Long: "You can specify arguments for the ssh command after a double dash (--). \n" +
			"For example: \n" +
			"	`unweave ssh -- -L 8080:localhost:8080`\n" +
			"   `unweave ssh <session-name|id> -- -L 8080:localhost:8080`\n",
		Hidden:  true,
		GroupID: groupDev,
		RunE:    withValidProjectURI(cmd.SSH),
	}
	sshCmd.Flags().BoolVar(&config.CreateExec, "new", false, "Create a new session if none exists")
	sshCmd.Flags().BoolVar(&config.NoCopySource, "no-copy", false, "Do not copy source code to the session")
	sshCmd.Flags().StringVarP(&config.BuildID, "image", "i", "", "Build ID of the container image to use")
	sshCmd.Flags().StringVar(&config.Provider, "provider", "", "Provider to use")
	sshCmd.Flags().StringVar(&config.NodeTypeID, "type", "", "Node type to use, eg. `gpu_1x_a100`")
	sshCmd.Flags().StringVar(&config.NodeRegion, "region", "", "Region to use, eg. `us_west_2`")
	sshCmd.Flags().StringVar(&config.SSHPrivateKeyPath, "prv", "", "Absolute Path to the SSH private key to use")
	rootCmd.AddCommand(sshCmd)

	// SSH Key commands
	sshKeyCmd := &cobra.Command{
		Use:     "ssh-keys",
		Aliases: []string{"ssh-key", "sshkey"},
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
