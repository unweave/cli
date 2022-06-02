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

	// Accept token to be passed manually - this overrides the token saved from interactive loginCmd
	rootCmd.PersistentFlags().StringVarP(&constants.AuthToken, "token", "t", "", "Use a specific token to authenticate - overrides loginCmd token")

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
	loginCmd := &cobra.Command{
		Use:   "login",
		Short: "Login through the browser or with an access token (--token)",
		RunE:  cmd.LoginCmd,
	}
	rootCmd.AddCommand(loginCmd)
	loginCmd.Flags().String("token", "", "--token <access_token>")

	// Logout
	rootCmd.AddCommand(&cobra.Command{
		Use:   "logout",
		Short: "Logout from your Unweave account",
		RunE:  cmd.LogoutCmd,
	})

	// Logs
	rootCmd.AddCommand(&cobra.Command{
		Use:   "logs <zepl-id>",
		Short: "Tail logs from a zepl run",
		RunE:  cmd.LogsCmd,
		Args:  cobra.ExactArgs(1),
	})

	// Open
	rootCmd.AddCommand(&cobra.Command{
		Use:   "open <project-id>",
		Short: "Open the dashboard in the browser. Optionally open to a specific project.",
		RunE:  cmd.OpenCmd,
		Args:  cobra.MaximumNArgs(1),
	})

	// Run
	runCmd := &cobra.Command{
		Use:   "run [flags] [<command>]",
		Short: "Run the current project in remotely with Unweave",
		Example: "unweave run python train.py\n" +
			"unweave run --gpu python train.py\n" +
			"unweave run --gpu --path=../rr-project python train.py\n",
		RunE: cmd.RunCmd,
		Args: cobra.RangeArgs(1, 2),
	}
	runCmd.Flags().BoolVarP(&config.IsGpu, "gpu", "g", false, "Use GPU")
	runCmd.Flags().StringVarP(&config.ZeplProjectPath, "path", "p", "", "Path to an Unweave project to run")
	rootCmd.AddCommand(runCmd)

	// Token
	tokenCmd := &cobra.Command{
		Use:   "token",
		Short: "Configure authentication tokens for the current user",
		Args:  cobra.NoArgs,
	}

	tokenCmd.AddCommand(&cobra.Command{
		Use:   "get-user-tokens",
		Short: "Get all tokens for the current user",
		RunE:  cmd.GetUserTokensCmd,
	})

	tokenCmd.AddCommand(&cobra.Command{
		Use:   "create-user-token",
		Short: "Create a new token",
		RunE:  cmd.CreateUserTokenCmd,
	})
	rootCmd.AddCommand(tokenCmd)

	// TODO: add ability to fetch project tokens
}

func main() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
