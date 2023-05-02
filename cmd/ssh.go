package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/unweave/cli/config"
	"github.com/unweave/cli/ui"
	"github.com/unweave/unweave/api/types"
)

func getUnweaveSSHConfigPath() string {
	return filepath.Join(config.GetDotUnweavePath(), "ssh_config")
}

func addHost(alias, host, user string, port int) error {
	configEntry := fmt.Sprintf(`Host %s
    HostName %s
    User %s
    Port %d
    StrictHostKeyChecking no
    UserKnownHostsFile /dev/null
    RequestTTY yes
	ForwardAgent yes
`, alias, host, user, port)

	file, err := os.OpenFile(getUnweaveSSHConfigPath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	// Make sure the host block doesn't already exist
	if e := removeHost(alias); e != nil {
		ui.Debugf("Failed to remove existing host block: %v", e)
	}

	if _, err = file.WriteString(configEntry); err != nil {
		return err
	}

	home, err := os.UserHomeDir()
	if err != nil {
		ui.Errorf("Failed to get user home directory: %v", err)
		os.Exit(1)
	}
	sshConfigPath := filepath.Join(home, ".ssh", "config")

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

func removeHost(alias string) error {
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

func removeKnownHostsEntry(hostname string) error {
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

func ssh(ctx context.Context, connectionInfo types.ConnectionInfo, prvKeyPath *string) error {
	sshCommand := exec.Command(
		"ssh",
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		fmt.Sprintf("%s@%s", connectionInfo.User, connectionInfo.Host),
	)
	if prvKeyPath != nil {
		sshCommand.Args = append(sshCommand.Args, "-i", *prvKeyPath)
	}

	ui.Debugf("Running SSH command: %s", strings.Join(sshCommand.Args, " "))

	sshCommand.Stdin = os.Stdin
	sshCommand.Stdout = os.Stdout
	sshCommand.Stderr = os.Stderr

	if err := sshCommand.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			// Exited with non-zero exit code
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				if status.ExitStatus() == 255 {
					ui.Infof("The remote host closed the connection.")
					return nil
				}
			}
			return err
		}
		return fmt.Errorf("SSH command failed: %v", err)
	}
	return nil
}

func SSH(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true
	ctx := cmd.Context()

	// TODO: parse args and forward to ssh command

	ui.Infof("Initializing node...")

	execch, errch, err := sessionCreateAndWatch(ctx, types.ExecCtx{})
	if err != nil {
		return err
	}

	for {
		select {
		case e := <-execch:
			if e.Status == types.StatusRunning {
				if e.Connection == nil {
					ui.Errorf("❌ Something unexpected happened. No connection info found for session %q", e.ID)
					ui.Infof("Run `unweave session ls` to see the status of your session and try connecting manually.")
					os.Exit(1)
				}
				ui.Infof("🚀 Session %q up and running", e.ID)

				if err := removeKnownHostsEntry(e.Connection.Host); err != nil {
					// Log and continue anyway. Most likely the entry is not there.
					ui.Debugf("Failed to remove known_hosts entry: %v", err)
				}

				if err := addHost("uw:"+e.ID, e.Connection.Host, e.Connection.User, e.Connection.Port); err != nil {
					ui.Debugf("Failed to add host to ssh config: %v", err)
				}

				defer func() {
					if e := removeHost("uw:" + e.ID); e != nil {
						ui.Debugf("Failed to remove host from ssh config: %v", e)
					}
				}()

				if err := ssh(ctx, *e.Connection, nil); err != nil {
					ui.Errorf("%s", err)
					os.Exit(1)
				}

				if terminate := ui.Confirm("SSH session terminated. Do you want to terminate the session?", "n"); terminate {
					if err := sessionTerminate(ctx, e.ID); err != nil {
						return err
					}
					ui.Infof("Session %q terminated.", e.ID)
				}
				return nil
			}

		case err := <-errch:
			var e *types.Error
			if errors.As(err, &e) {
				uie := &ui.Error{Error: e}
				fmt.Println(uie.Verbose())
				os.Exit(1)
			}
			return err

		case <-ctx.Done():
			return nil
		}
	}
}
