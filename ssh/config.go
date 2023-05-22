package ssh

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/unweave/cli/config"
	"github.com/unweave/cli/ui"
)

func getUnweaveSSHConfigPath() string {
	return filepath.Join(config.GetGlobalConfigPath(), "ssh_config")
}

var homeDirPath = func() string {
	path, err := os.UserHomeDir()
	if err != nil {
		ui.Errorf("Failed to get user home directory: %v", err)
		os.Exit(1)
	}
	return path
}()

var sshConfigPath = filepath.Join(homeDirPath, ".ssh", "config")

func AddHost(alias, host, user string, port int, identityFile string) error {
	configEntry := fmt.Sprintf(`Host %s
    HostName %s
    User %s
    Port %d
    StrictHostKeyChecking no
    UserKnownHostsFile /dev/null
    RequestTTY yes
    ForwardAgent yes
    IdentityFile %s
`, host, host, user, port, identityFile)

	file, err := os.OpenFile(getUnweaveSSHConfigPath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	// Make sure the host block doesn't already exist
	if e := RemoveHost(alias); e != nil {
		ui.Debugf("Failed to remove existing host block: %v", e)
	}

	if _, err = file.WriteString(configEntry); err != nil {
		return err
	}

	// Add an Include directive to the user's ssh config to unweave_global SSH configs - used for vscode-remote:
	err = os.MkdirAll(filepath.Join(homeDirPath, ".ssh"), 0700)
	if err != nil {
		fmt.Println("Failed to create .ssh folder:", err)
	}
	if _, err := os.Stat(sshConfigPath); os.IsNotExist(err) {
		if _, err = os.Create(sshConfigPath); err != nil {
			return err
		}
	}
	lines, err := readLines(sshConfigPath)
	if err != nil {
		return err
	}

	// Add to the top of the file if it doesn't already exist
	includeEntry := "Include " + getUnweaveSSHConfigPath()
	for _, line := range lines {
		if strings.HasPrefix(line, includeEntry) {
			return nil
		}
	}
	lines = append([]string{includeEntry}, lines...)

	return writeLines(sshConfigPath, lines)
}

func RemoveHost(alias string) error {
	lines, err := readLines(getUnweaveSSHConfigPath())
	if err != nil {
		return err
	}

	startIndex := -1
	for i, line := range lines {
		if strings.HasPrefix(line, "Host "+alias) {
			startIndex = i
			break
		}
	}

	if startIndex == -1 {
		ui.Debugf("Host block not found: %s", alias)
		return nil
	}
	ui.Debugf("Removing host block: %s", alias)

	endIndex := startIndex + 1
	for i := startIndex + 1; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "Host ") {
			break
		}
		endIndex = i
	}

	// If the host block is not the last block, append the rest of the lines
	if endIndex != len(lines)-1 {
		lines = append(lines[:startIndex], lines[endIndex:]...)
	} else {
		lines = lines[:startIndex]
	}

	return writeLines(getUnweaveSSHConfigPath(), lines)
}

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func writeLines(path string, lines []string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, line := range lines {
		fmt.Fprintln(writer, line)
	}
	return writer.Flush()
}

func RemoveKnownHostsEntry(hostname string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %v", err)
	}
	knownHostsFile := fmt.Sprintf("%s/.ssh/known_hosts", home)
	removeHostKeyCmd := exec.Command("ssh-keygen", "-R", hostname, "-f", knownHostsFile)

	ui.Debugf("Removing host key from known_hosts: %s", strings.Join(removeHostKeyCmd.Args, " "))

	if err = removeHostKeyCmd.Run(); err != nil {
		return fmt.Errorf("failed to remove host key from known_hosts: %v", err)
	}
	return nil
}
