package cmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/unweave/cli/ui"
	"github.com/unweave/unweave/api/types"
	"github.com/unweave/unweave/tools"
)

func sshKeyAdd(ctx context.Context, publicKeyPath string, name string) error {
	publicKey, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return fmt.Errorf("failed reading public key file: %v", err)
	}

	uwc := InitUnweaveClient()
	params := types.SSHKeyAddParams{
		Name:      &name,
		PublicKey: string(publicKey),
	}

	if err = uwc.SSHKey.Add(ctx, params); err != nil {
		var e *types.Error
		if errors.As(err, &e) {
			uie := &ui.Error{Error: e}
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

func sshKeyGenerate(ctx context.Context, name *string) (keyName string, pub []byte, err error) {
	uwc := InitUnweaveClient()
	params := types.SSHKeyGenerateParams{Name: name}
	res, err := uwc.SSHKey.Generate(ctx, params)
	if err != nil {
		return "", nil, ui.HandleError(err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", nil, err
	}

	prv := []byte(res.PrivateKey)
	pub = []byte(res.PublicKey)
	dotSSHPath := filepath.Join(home, ".ssh")
	publicKeyPath := filepath.Join(dotSSHPath, res.Name+".pub")
	privateKeyPath := filepath.Join(dotSSHPath, res.Name+".pem")

	if _, err := os.Stat(dotSSHPath); os.IsNotExist(err) {
		if err := os.MkdirAll(dotSSHPath, 0700); err != nil {
			ui.Errorf(".ssh directory not found and attempt to create it failed: %s", err)
			os.Exit(1)
			return "", nil, nil
		}
	}

	if err = os.WriteFile(privateKeyPath, prv, 0600); err != nil {
		ui.Errorf("Failed to write private key to %s: %v", privateKeyPath, err)
		os.Exit(1)
		return "", nil, nil
	}
	if err = os.WriteFile(publicKeyPath, pub, 0600); err != nil {
		ui.Errorf("Failed to write public key to %s: %v", publicKeyPath, err)
		os.Exit(1)
		return "", nil, nil
	}
	ui.Attentionf("Created new SSH key pair:\n"+
		"  Name: %s\n"+
		"  Path: %s\n",
		res.Name, publicKeyPath)

	return res.Name, pub, nil
}

func sshKeyGenerateFromRSA(ctx context.Context, name string, path string) (keyName string, pub []byte, err error) {
	filename := filepath.Base(path)
	if filename != "id_rsa" {
		ui.Errorf("Invalid RSA private key filename: %s. Only ida_rsa is supported", filename)
		return "", nil, fmt.Errorf("invalid RSA private key filename: %s", filename)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		ui.Errorf("Private key not found: %s", path)
		return "", nil, fmt.Errorf("private key not found: %s", path)
	}

	command := []string{
		"ssh-keygen",
		"-y",
		"-f",
		path,
	}

	cmd := exec.Command(command[0], command[1:]...)
	cmd.Stderr = os.Stderr

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return "", nil, err
	}
	if err = cmd.Start(); err != nil {
		return "", nil, err
	}
	var stdout bytes.Buffer
	_, err = io.Copy(&stdout, stdoutPipe)
	if err = cmd.Wait(); err != nil {
		return "", nil, err
	}
	output := stdout.String()
	pub = []byte(output)

	if err = os.WriteFile(path+".pub", pub, 0600); err != nil {
		ui.Attentionf("Could not write public key to file.")
	}

	ui.Infof("Generated public key from private key at path: %s", path)

	if err = sshKeyAdd(ctx, path+".pub", name); err != nil {
		return "", nil, err
	}
	return name, pub, nil
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
