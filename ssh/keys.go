package ssh

import (
	"context"
	"errors"
	"fmt"
	"os"
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

func Generate(ctx context.Context, owner string, name *string) (keyName string, pub []byte, err error) {
	uwc := config.InitUnweaveClient()
	params := types.SSHKeyGenerateParams{Name: name}
	res, err := uwc.SSHKey.Generate(ctx, owner, params)
	if err != nil {
		return "", nil, ui.HandleError(err)
	}

	prv := []byte(res.PrivateKey)
	pub = []byte(res.PublicKey)
	dotSSHPath := config.GetSSHKeysFolder()
	publicKeyPath := filepath.Join(dotSSHPath, res.Name+".pub")
	privateKeyPath := filepath.Join(dotSSHPath, res.Name)

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
	ui.Successf("Created new SSH key pair:\n"+
		"  Name: %s\n"+
		"  Path: %s\n",
		res.Name, publicKeyPath)

	return res.Name, pub, nil
}
