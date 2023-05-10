package cmd

import (
	"bufio"
	"bytes"
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
	return filepath.Join(config.GetGlobalConfigPath(), "ssh_config")
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

func copySource(execID, rootDir, dstPath string, connectionInfo types.ConnectionInfo, keyPath string) error {
	name := fmt.Sprintf("uw-context-%s.tar.gz", execID)
	tmpFile, err := os.CreateTemp(os.TempDir(), name)
	if err != nil {
		return err
	}

	ui.Infof("🧳 Gathering context from %q", rootDir)

	if err := gatherContext(rootDir, tmpFile, "tar"); err != nil {
		return fmt.Errorf("failed to gather context: %v", err)
	}

	tmpDstPath := filepath.Join("/tmp", name)
	scpCommand := exec.Command(
		"scp",
		"-r",
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		tmpFile.Name(),
		fmt.Sprintf("%s@%s:%s", connectionInfo.User, connectionInfo.Host, tmpDstPath),
	)
	if keyPath != "" {
		scpCommand.Args = append(scpCommand.Args, "-i", keyPath)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	scpCommand.Stdout = stdout
	scpCommand.Stderr = stderr

	ui.Infof("🔄 Copying source to %q", dstPath)

	if err := scpCommand.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			// Exited with non-zero exit code
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				if status.ExitStatus() == 255 {
					ui.Infof("The remote host closed the connection.")
					return fmt.Errorf("failed to copy source: %w", err)
				}
			}
			ui.Infof("Failed to copy source directory to remote host: %s", stderr.String())
			return err
		}
		return fmt.Errorf("scp command failed: %v", err)
	}

	// Unzip on remote host
	sshCommand := exec.Command(
		"ssh",
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-i", keyPath,
		fmt.Sprintf("%s@%s", connectionInfo.User, connectionInfo.Host),
		fmt.Sprintf("tar -xzf %s -C %s && rm -rf %s", tmpDstPath, dstPath, tmpDstPath),
	)

	sshCommand.Stdout = stdout
	sshCommand.Stderr = stderr

	if err := sshCommand.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			// Exited with non-zero exit code
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				if status.ExitStatus() == 255 {
					ui.Infof("The remote host closed the connection.")
					return fmt.Errorf("failed to copy source: %w", err)
				}
			}
			ui.Infof("Failed to extract source directory on remote host: %s", stderr.String())
			return err
		}
		return fmt.Errorf("failed to unzip on remote host: %v", err)
	}

	ui.Infof("✅  Successfully copied source directory to remote host")

	return nil
}

func ssh(ctx context.Context, connectionInfo types.ConnectionInfo, args []string) error {
	overrideUserKnownHostsFile := false
	overrideStrictHostKeyChecking := false

	for _, arg := range args {
		if strings.Contains(arg, "UserKnownHostsFile") {
			overrideUserKnownHostsFile = true
		}
		if strings.Contains(arg, "StrictHostKeyChecking") {
			overrideStrictHostKeyChecking = true
		}
	}

	if !overrideUserKnownHostsFile {
		args = append(args, "-o", "UserKnownHostsFile=/dev/null")
	}
	if !overrideStrictHostKeyChecking {
		args = append(args, "-o", "StrictHostKeyChecking=no")
	}

	sshCommand := exec.Command(
		"ssh",
		append(args, fmt.Sprintf("%s@%s", connectionInfo.User, connectionInfo.Host))...,
	)

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

	errMsg := "❌ Invalid arguments. If you want to pass arguments to the ssh command, " +
		"use the -- flag. See `unweave ssh --help` for  more information."

	execCh := make(chan types.Exec)
	errCh := make(chan error)
	isNew := false

	var err error
	var sshArgs []string
	var execRef string // Can be execID or name

	// If the number of args is great than one, we always expect the first arg to be
	// the separator flag "--". If the number of args is one, we expect it to be the
	// execID or name
	if len(args) > 1 {
		sshArgs = args[1:]
		if sshArgs[0] != "--" {
			ui.Errorf(errMsg)
			os.Exit(1)
		}

		if len(args) == 1 {
			execRef = args[0]
		} else {
			execRef = args[0]
		}
	}

	watchCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	if config.CreateExec {
		// If the flag to create a new exec is passed, any arguments must be forwarded to
		// the ssh command
		if len(args) > 0 && args[0] != "--" {
			ui.Errorf(errMsg)
			os.Exit(1)
		}

		ui.Infof("Initializing node...")

		execCh, errCh, err = execCreateAndWatch(watchCtx, types.ExecConfig{}, types.GitConfig{})
		if err != nil {
			return err
		}
		isNew = true

	} else {

		if execRef == "" {
			var execs []types.Exec

			execRef, execs, err = selectExec(cmd.Context(), "Select a session to connect to")
			if err != nil {
				return err
			}
			if len(execs) == 0 {
				ui.Errorf("❌ No active sessions found and no session name or ID provided. If " +
					"you want to create a new session, use the --new flag.")
				os.Exit(1)
			}
		}

		execCh, errCh, err = execWaitTillReady(watchCtx, execRef)
		if err != nil {
			return err
		}
	}

	for {
		select {
		case e := <-execCh:
			if e.Status == types.StatusRunning {
				if e.Connection == nil {
					ui.Errorf("❌ Something unexpected happened. No connection info found for session %q", e.ID)
					ui.Infof("Run `unweave ls` to see the status of your session and try connecting manually.")
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

				if !config.NoCopySource && isNew {
					dir, err := config.GetActiveProjectPath()
					if err != nil {
						ui.Errorf("Failed to get active project path. Skipping copying source directory")
						return fmt.Errorf("failed to get active project path: %v", err)
					}

					if err := copySource(e.ID, dir, "/home/ubuntu", *e.Connection, ""); err != nil {
						fmt.Println(err)
					}
				} else {
					ui.Infof("Skipping copying source directory")
				}

				if err := ssh(ctx, *e.Connection, args); err != nil {
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

		case err := <-errCh:
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
