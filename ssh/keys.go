package ssh

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/unweave/cli/config"
	"github.com/unweave/cli/ui"
	"github.com/unweave/unweave/api/types"
)

func Add(ctx context.Context, publicKeyPath, owner string, name *string) (string, error) {
	publicKey, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return "", fmt.Errorf("failed reading public key file: %v", err)
	}

	user := strings.Split(config.Config.Unweave.User.Email, "@")[0]

	filename := filepath.Base(publicKeyPath)
	if filename == "id_rsa" || filename == "id_rsa.pub" {
		filename = fmt.Sprintf("%s_id_rsa.pub", user)
	}

	// If the name is id_rsa or id_rsa.pub, we'll use the user's email address to avoid conflicts
	if name != nil {
		if *name == "id_rsa" || *name == "id_rsa.pub" {
			n := fmt.Sprintf("%s_id_rsa.pub", user)
			name = &n
		}
	} else {
		name = &filename
	}

	uwc := config.InitUnweaveClient()
	params := types.SSHKeyAddParams{
		Name:      name,
		PublicKey: string(publicKey),
	}

	keyname, err := uwc.SSHKey.Add(ctx, owner, params)
	if err != nil {
		var e *types.Error
		if errors.As(err, &e) {
			uie := &ui.Error{Error: e}
			fmt.Println(uie.Verbose())
			os.Exit(1)
		}
		return "", err
	}
	return keyname, nil
}

func Generate(ctx context.Context, owner string, name *string) (keyname, keypath string, pub []byte, err error) {
	uwc := config.InitUnweaveClient()
	params := types.SSHKeyGenerateParams{Name: name}

	res, err := uwc.SSHKey.Generate(ctx, owner, params)
	if err != nil {
		return "", "", nil, ui.HandleError(err)
	}

	prv := []byte(res.PrivateKey)
	pub = []byte(res.PublicKey)

	sshDir := config.GetUnweaveSSHKeysFolder()
	pubPath := filepath.Join(sshDir, res.Name+".pub")
	prvPath := filepath.Join(sshDir, res.Name)

	if err = os.WriteFile(prvPath, prv, 0600); err != nil {
		ui.Errorf("Failed to write private key to %s: %v", prvPath, err)
		os.Exit(1)
		return "", "", nil, nil
	}

	if err = os.WriteFile(pubPath, pub, 0600); err != nil {
		ui.Errorf("Failed to write public key to %s: %v", pubPath, err)
		os.Exit(1)
		return "", "", nil, nil
	}

	return res.Name, pubPath, pub, nil
}

func GenerateFromPrivateKey(ctx context.Context, path string, name *string) (keyName string, pub []byte, err error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		ui.Errorf("Private key not found: %s", path)
		return "", nil, fmt.Errorf("private key not found: %s", path)
	}

	if name == nil {
		n := filepath.Base(path)
		name = &n
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

	keyname, err := Add(ctx, path+".pub", config.Config.Unweave.User.ID, name)
	if err != nil {
		return "", nil, err
	}
	return keyname, pub, nil
}
