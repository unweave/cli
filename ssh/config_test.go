package ssh

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/franela/goblin"
)

func TestConfig(t *testing.T) {
	g := Goblin(t)

	g.Describe("AddHost", func() {
		testCfgPath := filepath.Join(homeDirPath, ".ssh", "test_config")
		unweaveConfigPath := getUnweaveSSHConfigPath()

		g.BeforeEach(func() {
			if _, err := os.Stat(testCfgPath); os.IsNotExist(err) {
				initialConfig := "# Test SSH Config\n"
				err = ioutil.WriteFile(testCfgPath, []byte(initialConfig), 0600)
				g.Assert(err).Equal(nil)
			}

			sshConfigPath = testCfgPath
		})

		g.AfterEach(func() {
			// Only perform create and destroy operations on the test config file
			if testCfgPath != sshConfigPath {
				return
			}

			err := os.Remove(testCfgPath)
			g.Assert(err).Equal(nil)
		})

		g.It("should add the Include directive to the .ssh/config file", func() {
			err := AddHost("example", "example.com", "user", 22, "")
			g.Assert(err).Equal(nil)

			configData, err := ioutil.ReadFile(sshConfigPath)
			g.Assert(err).Equal(nil)

			configContent := string(configData)
			expectedInclude := "Include " + unweaveConfigPath

			g.Assert(strings.Contains(configContent, expectedInclude)).Equal(true)
		})
	})
}
