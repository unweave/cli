package client

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"

	"github.com/unweave/cli/ssh/grpcproxy/conn"
	"github.com/unweave/cli/ui"
	"github.com/unweave/unweave/proto/proxyssh"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func GrpcProxyConn(serverAddr, sshAddr string) (net.Conn, error) {
	creds := insecure.NewCredentials()

	gconn, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("grpc dial: %w", err)
	}

	client := proxyssh.NewProxyConnectionServiceClient(gconn)

	mdCtx := metadata.AppendToOutgoingContext(context.Background(), "content-type", "application/grpc")

	stream, err := client.ProxySSH(mdCtx)
	if err != nil {
		return nil, fmt.Errorf("create stream: %w", err)
	}

	if err = stream.Send(&proxyssh.ProxySSHRequest{
		DialTarget: &proxyssh.TargetHost{
			HostPort: sshAddr,
		},
	}); err != nil {
		return nil, fmt.Errorf("initial dial: %w", err)
	}

	resp, err := stream.Recv()
	if err != nil {
		return nil, fmt.Errorf("recv connect success: %w", err)
	}

	if !resp.Connect {
		return nil, errors.New("Expected response connect")
	}

	host, port, err := net.SplitHostPort(serverAddr)
	if err != nil {
		return nil, fmt.Errorf("parse serverhost port: %w", err)
	}

	if host == "localhost" {
		host = "127.0.0.1"
	}

	portNum, err := strconv.Atoi(port)
	if err != nil {
		return nil, fmt.Errorf("parse port into number: %w", err)
	}

	rw := conn.NewReadWriter(&clientSource{stream: stream})
	cconn := &clientConn{
		closer: func() {
			gconn.Close()
		},
		rw:  rw,
		src: nil,
		dst: &net.TCPAddr{
			IP:   net.ParseIP(host),
			Port: portNum,
		},
	}

	ui.Debugf("[grpc] created proxy connection for %q, %q", host, port)

	return cconn, nil
}

type clientSource struct {
	stream proxyssh.ProxyConnectionService_ProxySSHClient
}

func (c *clientSource) Send(b []byte) error {
	return c.stream.Send(&proxyssh.ProxySSHRequest{
		Ssh: &proxyssh.Frame{
			Payload: b,
		},
	})
}

func (c *clientSource) Recv() ([]byte, error) {
	req, err := c.stream.Recv()
	if err != nil {
		return nil, fmt.Errorf("client stream source: %w", err)
	}

	return req.Ssh.Payload, nil
}
