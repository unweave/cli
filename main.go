package main

import (
	"github.com/spf13/cobra"
	"github.com/unweave/cli/cmd"
)

var rootCmd = &cobra.Command{
	Use:           "unweave [<command>]",
	Short:         "Zero setup ML infrastructure",
	Long:          "Instant access to the environments and infra you need to do ML, all versioned with Git.",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	// List
	rootCmd.AddCommand(&cobra.Command{
		Use:   "init",
		Short: "Init new Unweave project",
		Long:  "Initialized in the current directory or at the path specified as an argument",
		RunE:  cmd.InitCmd,
	})

	// List
	rootCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all projects",
		RunE:  cmd.ListCmd,
	})

	// Login
	login := &cobra.Command{
		Use:   "login",
		Short: "Login through the browser of with an access token (--headless)",
		RunE:  cmd.LoginCmd,
	}
	rootCmd.AddCommand(login)
	login.Flags().Bool("headless", false, "--headless")

	// Logout
	rootCmd.AddCommand(&cobra.Command{
		Use:   "logout",
		Short: "Logout from your Unweave account",
		RunE:  cmd.LogoutCmd,
	})
}

func main() {
	rootCmd.Execute()
}
