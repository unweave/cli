package ssh

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/unweave/cli/ssh/grpcproxy/client"
	"github.com/unweave/cli/ui"
)

type Proxied struct {
	user        string
	privKeyPath string
	sshAddr     string

	serverAddr string
	listenHost string
	listenPort string
	proxyC     net.Conn
	lis        net.Listener

	wg   *sync.WaitGroup
	done chan struct{}
}

func NewProxied(user, listenAddr, serverAddr, sshAddr, privKeyPath string) (*Proxied, error) {
	host, port, err := net.SplitHostPort(listenAddr)
	if err != nil {
		return nil, fmt.Errorf("parse listen addr: %w")
	}

	p := &Proxied{
		serverAddr:  serverAddr,
		listenHost:  host,
		listenPort:  port,
		user:        user,
		privKeyPath: privKeyPath,
		sshAddr:     sshAddr,
		wg:          &sync.WaitGroup{},
	}

	return p, nil
}

func (p *Proxied) setup() error {
	p.done = make(chan struct{})
	p.wg.Add(2)

	ui.Debugf("[proxyssh] setting up grpc proxy connection for: %s:%s => %s => %s", p.listenHost, p.listenPort, p.serverAddr, p.sshAddr)

	proxyC, err := client.GrpcProxyConn(p.serverAddr, p.sshAddr)
	if err != nil {
		return fmt.Errorf("setup grpc conn: %w", err)
	}

	p.proxyC = proxyC

	lis, err := net.Listen("tcp", fmt.Sprint(p.listenHost, ":", p.listenPort))
	if err != nil {
		return fmt.Errorf("listening: %w", err)
	}

	p.lis = lis

	go func() {
		defer p.wg.Done()
		<-p.done
		lis.Close()
		proxyC.Close()
	}()

	go func() {
		defer p.wg.Done()

		for {
			select {
			case <-p.done:
				return
			default:
			}

			conn, err := lis.Accept()
			if err != nil {
				ui.Debugf("[proxiedssh] Error on listener: %w", err)
				return
			}

			defer conn.Close()

			go io.Copy(proxyC, conn)
			go io.Copy(conn, proxyC)
		}
	}()

	return nil
}

func (p *Proxied) teardown() {
	defer ui.Debugf("Closed proxied ssh connection")
	close(p.done)
	p.wg.Wait()
}

func (p *Proxied) RunCommand(options []string, command []string) error {
	if err := p.setup(); err != nil {
		return fmt.Errorf("conn setup: %w", err)
	}
	defer p.teardown()

	sshArgs := append(options, "-p", p.listenPort, fmt.Sprintf("%s@%s", p.user, p.listenHost))

	if len(command) != 0 {
		sshArgs = append(sshArgs, command...)
	}

	sshCommand := exec.Command(
		"ssh",
		sshArgs...,
	)

	ui.Debugf("[proxiedssh] Running command: %s", strings.Join(sshCommand.Args, " "))

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

func (p *Proxied) DownloadSCP(remoteDir, localDir string) error {
	if err := p.setup(); err != nil {
		return fmt.Errorf("conn setup: %w", err)
	}
	defer p.teardown()

	remoteTarget := fmt.Sprintf("%s@%s:%s", p.user, p.listenHost, remoteDir)

	return p.copySourceSCP(remoteTarget, localDir)
}

func (p *Proxied) UploadSCP(localDir, remoteDir string) error {
	if err := p.setup(); err != nil {
		return fmt.Errorf("conn setup: %w", err)
	}
	defer p.teardown()

	remoteTarget := fmt.Sprintf("%s@%s:%s", p.user, p.listenHost, remoteDir)

	return p.copySourceSCP(localDir, remoteTarget)
}

func (p *Proxied) copySourceSCP(from, to string) error {
	scpCommandArgs := []string{
		"-r",
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-P", p.listenPort,
	}
	if p.privKeyPath != "" {
		scpCommandArgs = append(scpCommandArgs, "-i", p.privKeyPath)
	}

	scpCommandArgs = append(scpCommandArgs, []string{from, to}...)

	scpCommand := exec.Command("scp", scpCommandArgs...)

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	scpCommand.Stdout = stdout
	scpCommand.Stderr = stderr

	ui.Debugf("Running command: %+v", scpCommand)

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

	return nil
}

func (p *Proxied) CopySourceUnTar(srcPath, dstPath string) error {
	if err := p.setup(); err != nil {
		return fmt.Errorf("conn setup: %w", err)
	}
	defer p.teardown()

	sshCommand := exec.Command(
		"ssh",
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-i", p.privKeyPath,
		"-p", p.listenPort,
		fmt.Sprintf("%s@%s", p.user, p.listenHost),
		// ensure dstPath exist and root logs into that path
		fmt.Sprintf("mkdir -p %s && echo 'cd %s' > /root/.bashrc &&", dstPath, dstPath),
		fmt.Sprintf("tar -xzf %s -C %s && rm -rf %s", srcPath, dstPath, srcPath),
	)

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	sshCommand.Stdout = stdout
	sshCommand.Stderr = stderr

	ui.Debugf("[proxiedssh] Running command: %+v", sshCommand)

	if err := sshCommand.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			// Exited with non-zero exit code
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				if status.ExitStatus() == 255 {
					ui.Debugf("stdout: %s", stdout.String())
					ui.Debugf("stderr: %s", stderr.String())
					ui.Infof("The remote host closed the connection. %s", err)
					return fmt.Errorf("failed to copy source: %w", err)
				}
			}
			ui.Infof("Failed to extract source directory on remote host: %s", stderr.String())
			return err
		}
		return fmt.Errorf("failed to unzip on remote host: %v", err)
	}

	return nil
}

// tarRemoteDirectory takes an ssh target, and zips up the contents of that target to a returned in the remote /tmp
func (p *Proxied) TarRemoteDirectory(dir string) (remoteArchiveLoc string, err error) {
	if err := p.setup(); err != nil {
		return "", fmt.Errorf("conn setup: %w", err)
	}
	defer p.teardown()

	timestamp := time.Now().Unix()
	remoteArchiveLoc = fmt.Sprintf("/tmp/uw-context-%d.tar.gz", timestamp)
	tarCmd := fmt.Sprintf("tar -czf %s -C %s .", remoteArchiveLoc, dir)

	sshCommand := exec.Command(
		"ssh",
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-i", p.privKeyPath,
		fmt.Sprintf("%s@%s -p %s", p.user, p.listenHost, p.listenPort),
		tarCmd,
	)

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	sshCommand.Stdout = stdout
	sshCommand.Stderr = stderr

	ui.Debugf("[proxiedssh] Running command: %+v", sshCommand)

	if err := sshCommand.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			// Exited with non-zero exit code
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				if status.ExitStatus() == 255 {
					ui.Infof("The remote host closed the connection.")
					return "", fmt.Errorf("failed to copy source: %w", err)
				}
			}
			ui.Infof("Failed to extract source directory on remote host: %s", stderr.String())
			return "", err
		}
		return "", fmt.Errorf("failed to unzip on remote host: %v", err)
	}

	return remoteArchiveLoc, nil
}
