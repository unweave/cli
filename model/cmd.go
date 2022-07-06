package model

import "github.com/spf13/cobra"

type Command struct {
	Cmd  *cobra.Command
	Args []string
}
