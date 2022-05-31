package main

import (
	"github.com/spf13/cobra"
	"github.com/unweave/cli/cmd"
	"github.com/unweave/cli/config"
	"github.com/unweave/cli/constants"
	"os"
)

var rootCmd = &cobra.Command{
	Use:           "unweave [command]",
	Short:         "Zero setup ML infrastructure",
	Long:          "Instant access to the environments and infra you need to do ML, all versioned with Git.",
	RunE:          cmd.RootCmd,
	SilenceUsage:  false,
	SilenceErrors: false,
}

func init() {
	rootCmd.Version = constants.Version
	rootCmd.Flags().BoolP("version", "v", false, "Get the version of current Unweave CLI")
	rootCmd.Flags().BoolVarP(&config.ShowConfig, "config", "c", false, "Show the current config")

	// Accept token to be passed manually - this overrides the token saved from interactive login
	rootCmd.PersistentFlags().StringVarP(&constants.AuthToken, "token", "t", "", "Use a specific token to authenticate - overrides login token")

	// Connect
	rootCmd.AddCommand(&cobra.Command{
		Use:   "connect <project-id> <run-id>",
		Short: "Connect to logs from a active session",
		RunE:  cmd.ConnectCmd,
		Args:  cobra.ExactArgs(2),
	})

	// Init
	rootCmd.AddCommand(&cobra.Command{
		Use:   "init",
		Short: "Init new Unweave project",
		Long:  "Initialized in the current directory or at the path specified as an argument",
		RunE:  cmd.InitCmd,
	})

	// Link
	rootCmd.AddCommand(&cobra.Command{
		Use:   "link <project-id> [<path>]",
		Short: "Link an Unweave project with ID <project-id> to local folder",
		RunE:  cmd.LinkCmd,
		Args:  cobra.RangeArgs(1, 2),
	})

	// List
	rootCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all projects",
		RunE:  cmd.ListCmd,
		Args:  cobra.NoArgs,
	})

	// Login
	login := &cobra.Command{
		Use:   "login",
		Short: "Login through the browser or with an access token (--token)",
		RunE:  cmd.LoginCmd,
	}
	rootCmd.AddCommand(login)
	login.Flags().String("token", "", "--token <access_token>")

	// Logout
	rootCmd.AddCommand(&cobra.Command{
		Use:   "logout",
		Short: "Logout from your Unweave account",
		RunE:  cmd.LogoutCmd,
	})

	// Open
	rootCmd.AddCommand(&cobra.Command{
		Use:   "open <project-id>",
		Short: "Open the dashboard in the browser. Optionally open to a specific project.",
		RunE:  cmd.OpenCmd,
		Args:  cobra.MaximumNArgs(1),
	})

	// Run
	run := &cobra.Command{
		Use:   "run [flags] [<command>]",
		Short: "Run the current project in remotely with Unweave",
		Example: "unweave run python train.py\n" +
			"unweave run --gpu python train.py\n" +
			"unweave run --gpu --path=../rr-project python train.py\n",
		RunE: cmd.RunCmd,
		Args: cobra.RangeArgs(1, 2),
	}
	run.Flags().BoolVarP(&config.IsGpu, "gpu", "g", false, "Use GPU")
	run.Flags().StringVarP(&config.ZeplProjectPath, "path", "p", "", "Path to an Unweave project to run")
	rootCmd.AddCommand(run)

	// Token
	token := &cobra.Command{
		Use:   "token",
		Short: "Configure authentication tokens for the current user",
		Args:  cobra.NoArgs,
	}

	getUserTokens := &cobra.Command{
		Use:   "get-user-tokens",
		Short: "Get all tokens for the current user",
		RunE:  cmd.GetUserTokensCmd,
	}
	token.AddCommand(getUserTokens)

	createUserToken := &cobra.Command{
		Use:   "create-user-token",
		Short: "Create a new token",
		RunE:  cmd.CreateUserTokenCmd,
	}
	token.AddCommand(createUserToken)
	rootCmd.AddCommand(token)

	// TODO: add ability to fetch project tokens
}

func main() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
