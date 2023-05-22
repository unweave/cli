package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	. "github.com/franela/goblin"
)

func TestUseUnweaveGlobalSSHKeys(t *testing.T) {
	g := Goblin(t)
	ctx := context.Background()

	g.Describe("useUnweaveGlobalSSHKeys", func() {
		var tmpDir string

		g.BeforeEach(func() {
			// Create a temporary directory for testing
			dir, err := ioutil.TempDir("", "ssh-keys")
			if err != nil {
				g.Fail(err)
			}
			tmpDir = dir
		})

		g.AfterEach(func() {
			os.RemoveAll(tmpDir)
		})

		g.It("should retrieve the public key when it exists", func() {
			// test public key file
			pubKeyFile := filepath.Join(tmpDir, "test_key.pub")
			err := ioutil.WriteFile(pubKeyFile, []byte("test public key"), 0644)
			if err != nil {
				g.Fail(err)
			}

			keyname, pubKeyBytes, err := getFirstPublicKeyInPath(ctx, tmpDir)
			g.Assert(err).Equal(nil)

			expectedKeyname := "test_key"
			expectedPubKey := []byte("test public key")
			g.Assert(keyname).Equal(expectedKeyname)
			g.Assert(pubKeyBytes).Equal(expectedPubKey)
		})

		g.It("should return an error when no public key is found", func() {
			emptyDir := filepath.Join(tmpDir, "empty")
			err := os.Mkdir(emptyDir, 0755)
			if err != nil {
				g.Fail(err)
			}

			_, _, err = getFirstPublicKeyInPath(ctx, emptyDir)
			g.Assert(err.Error()).Equal(fmt.Sprintf("no public SSH key found in %s", emptyDir))
		})

		g.It("should return an error when the directory cannot be read", func() {
			invalidDir := filepath.Join(tmpDir, "nonexistent")
			_, _, err := getFirstPublicKeyInPath(ctx, invalidDir)
			g.Assert(err.Error()).Equal(fmt.Sprintf("directory %s cannot be read", invalidDir))
		})
	})
}
