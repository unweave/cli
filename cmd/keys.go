package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/unweave/cli/ui"
	"github.com/unweave/unweave/api/types"
	"github.com/unweave/unweave/tools"
)

func sshKeyAdd(ctx context.Context, publicKeyPath, name string) error {
	publicKey, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return fmt.Errorf("failed reading public key file: %v", err)
	}

	uwc := InitUnweaveClient()
	params := types.SSHKeyAddRequestParams{
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

func SSHKeyAdd(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true
	publicKeyPath := args[0]
	name := filepath.Base(publicKeyPath)
	if len(args) == 2 {
		name = args[1]
	}
	return sshKeyAdd(cmd.Context(), publicKeyPath, name)
}

func sshKeyGenerate(ctx context.Context, name *string) (string, string, error) {
	uwc := InitUnweaveClient()
	params := types.SSHKeyGenerateRequestParams{Name: name}
	res, err := uwc.SSHKey.Generate(ctx, params)
	if err != nil {
		return "", "", ui.HandleError(err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", "", err
	}

	dotSSHPath := filepath.Join(home, ".ssh")
	publicKeyPath := filepath.Join(dotSSHPath, res.Name+".pub")
	privateKeyPath := filepath.Join(dotSSHPath, res.Name+".pem")

	if err = os.WriteFile(privateKeyPath, []byte(res.PrivateKey), 0600); err != nil {
		ui.Errorf("Failed to write private key to %s: %v", privateKeyPath, err)
		os.Exit(1)
		return "", "", nil
	}
	if err = os.WriteFile(publicKeyPath, []byte(res.PublicKey), 0600); err != nil {
		ui.Errorf("Failed to write public key to %s: %v", publicKeyPath, err)
		os.Exit(1)
		return "", "", nil
	}
	ui.Attentionf("Created new SSH key pair:\n"+
		"  Name: %s\n"+
		"  Path: %s\n",
		res.Name, publicKeyPath)

	return res.Name, publicKeyPath, nil
}

func SSHKeyGenerate(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true
	var name *string
	if len(args) != 0 {
		name = tools.Stringy(args[0])
	}
	_, _, err := sshKeyGenerate(cmd.Context(), name)
	return err
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
