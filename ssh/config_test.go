package ssh

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/franela/goblin"
)

func TestAddHost(t *testing.T) {
	g := Goblin(t)

	g.Describe("AddHost", func() {
		homeDir, _ := os.UserHomeDir()
		sshConfigPath := filepath.Join(homeDir, ".ssh", "test_config")
		unweaveConfigPath := getUnweaveSSHConfigPath()

		g.BeforeEach(func() {
			// Remove any existing config files
			os.Remove(sshConfigPath)
		})

		g.AfterEach(func() {
			// Remove the created config files
			os.Remove(sshConfigPath)
		})

		g.It("should add the Include directive to the .ssh/config file", func() {
			err := AddHost("example", "example.com", "user", 22, sshConfigPath)
			g.Assert(err).Equal(nil)

			configData, err := ioutil.ReadFile(sshConfigPath)
			g.Assert(err).Equal(nil)

			configContent := string(configData)
			expectedInclude := "Include " + unweaveConfigPath

			g.Assert(strings.Contains(configContent, expectedInclude)).Equal(true)
		})
	})
}
