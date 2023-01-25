package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/unweave/cli/ui"
	"github.com/unweave/unweave/api/types"
)

func SSHKeyAdd(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	publicKeyPath := args[0]
	name := filepath.Base(publicKeyPath)

	if len(args) == 2 {
		name = args[1]
	}

	publicKey, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return fmt.Errorf("failed reading public key file: %v", err)
	}

	ctx := cmd.Context()
	uwc := InitUnweaveClient()
	params := types.SSHKeyAddParams{
		Name:      &name,
		PublicKey: string(publicKey),
	}

	if err = uwc.SSHKey.Add(ctx, params); err != nil {
		var e *types.HTTPError
		if errors.As(err, &e) {
			uie := &ui.Error{HTTPError: e}
			fmt.Println(uie.Verbose())
			os.Exit(1)
		}
		return err
	}
	return nil
}

func SSHKeyList(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	ctx := cmd.Context()
	uwc := InitUnweaveClient()

	entries, err := uwc.SSHKey.List(ctx)
	if err != nil {
		var e *types.HTTPError
		if errors.As(err, &e) {
			uie := &ui.Error{HTTPError: e}
			fmt.Println(uie.Verbose())
			os.Exit(1)
		}
		return err
	}

	cols := []ui.Column{
		{Title: "Name", Width: 20},
		{Title: "Created", Width: 25},
		{Title: "Public Key", Width: 50},
	}
	rows := make([]ui.Row, len(entries))

	for idx, entry := range entries {
		publicKey := ""
		if entry.PublicKey != nil && len(*entry.PublicKey) > 50 {
			publicKey = *entry.PublicKey
			publicKey = publicKey[len(publicKey)-50:]
		}

		rows[idx] = ui.Row{
			fmt.Sprintf("%s", entry.Name),
			fmt.Sprintf("%s", entry.CreatedAt.Format("2006-01-02 15:04:05")),
			fmt.Sprintf("%s", publicKey),
		}
	}
	ui.Table("SSH Keys", cols, rows)
	return nil
}
