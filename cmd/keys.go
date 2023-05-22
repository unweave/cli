package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/unweave/cli/config"
	"github.com/unweave/cli/ssh"
	"github.com/unweave/cli/ui"
	"github.com/unweave/unweave/api/types"
	"github.com/unweave/unweave/tools"
)

func SSHKeyAdd(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true
	publicKeyPath := args[0]
	name := filepath.Base(publicKeyPath)
	if len(args) == 2 {
		name = args[1]
	}
	keyname, err := ssh.Add(cmd.Context(), publicKeyPath, config.Config.Unweave.User.ID, &name)
	if err != nil {
		return err
	}

	ui.Successf("SSH key added as %q", keyname)
	return nil
}

func sshKeyAddIDRSA(ctx context.Context, path string, name *string) (keyName string, pub []byte, err error) {
	filename := filepath.Base(path)
	if filename != "id_rsa" {
		ui.Errorf("Invalid RSA private key filename: %s. Only ida_rsa is supported", filename)
		return "", nil, fmt.Errorf("invalid RSA private key filename: %s", filename)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		ui.Errorf("Private key not found: %s", path)
		return "", nil, fmt.Errorf("private key not found: %s", path)
	}

	if _, err := os.Stat(path + ".pub"); os.IsNotExist(err) {
		ui.Errorf("Public key not found: %s", path+".pub")
		return "", nil, fmt.Errorf("public key not found: %s", path+".pub")
	}

	keyname, err := ssh.Add(ctx, path+".pub", config.Config.Unweave.User.ID, name)
	if err != nil {
		return "", nil, err
	}

	return keyname, pub, nil
}

func SSHKeyGenerate(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true
	var name *string
	if len(args) != 0 {
		name = tools.Stringy(args[0])
	}
	keyname, keypath, _, err := ssh.Generate(cmd.Context(), config.Config.Unweave.User.ID, name)
	if err != nil {
		return err
	}

	ui.Successf("Created new SSH key pair:\n"+
		"  Name: %s\n"+
		"  Path: %s\n",
		keyname, keypath)

	return err
}

func SSHKeyList(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	ctx := cmd.Context()
	uwc := config.InitUnweaveClient()

	entries, err := uwc.SSHKey.List(ctx, config.Config.Unweave.User.ID)
	if err != nil {
		var e *types.Error
		if errors.As(err, &e) {
			uie := &ui.Error{Error: e}
			fmt.Println(uie.Verbose())
			os.Exit(1)
		}
		return err
	}

	cols := []ui.Column{
		{Title: "Name", Width: -1},
		{Title: "Created", Width: 25},
		{Title: "Public Key", Width: 50},
	}
	rows := make([]ui.Row, len(entries))

	for idx, entry := range entries {
		publicKey := ""
		if entry.PublicKey != nil && len(*entry.PublicKey) > 50 {
			publicKey = *entry.PublicKey
			publicKey = publicKey[:50]
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

func getFirstPublicKeyInPath(ctx context.Context, dirPath string) (name string, pubKey []byte, err error) {
	_, err = os.Stat(dirPath)
	if err != nil {
		return "", nil, fmt.Errorf("directory %s cannot be read", dirPath)
	}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return "", nil, err
	}

	keys := filterPublicKeys(entries)

	if len(keys) == 0 {
		return "", nil, fmt.Errorf("no public SSH key found in %s", dirPath)
	}

	filename := keys[0].Name()
	pubKeyPath := filepath.Join(dirPath, filename)
	pubKeyBytes, err := os.ReadFile(pubKeyPath)
	if err != nil {
		return "", nil, err
	}

	keyname := strings.TrimSuffix(filename, ".pub")
	return keyname, pubKeyBytes, nil
}

func filterPublicKeys(entries []os.DirEntry) []os.DirEntry {
	var keys []os.DirEntry

	for _, entry := range entries {
		if !entry.IsDir() && isPublicKey(entry.Name()) {
			keys = append(keys, entry)
		}
	}

	return keys
}

func isPublicKey(filename string) bool {
	return strings.HasSuffix(filename, ".pub")
}
