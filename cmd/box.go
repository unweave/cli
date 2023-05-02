package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func BoxUp(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true
	//var filesystemID *string
	//
	//if len(args) > 0 {
	//	filesystemID = &args[0]
	//}
	fmt.Println("Box command not implemented yet")
	return nil
	//if _, err := sessionCreate(cmd.Context(), types.ExecConfig{}, types.GitConfig{}); err != nil {
	//	os.Exit(1)
	//	return nil
	//}
	//
	//return nil
}
