package client

import (
	"net"
	"time"

	"github.com/unweave/cli/ssh/grpcproxy/conn"
)

type clientConn struct {
	rw *conn.ReadWriter

	closer func()

	src net.Addr
	dst net.Addr
}

// Read reads data from the connection.
// Read can be made to time out and return an error after a fixed
// time limit; see SetDeadline and SetReadDeadline.
func (c *clientConn) Read(b []byte) (n int, err error) {
	return c.rw.Read(b)
}

// Write writes data to the connection.
// Write can be made to time out and return an error after a fixed
// time limit; see SetDeadline and SetWriteDeadline.
func (c *clientConn) Write(b []byte) (n int, err error) {
	return c.rw.Write(b)
}

// Close closes the connection.
// Any blocked Read or Write operations will be unblocked and return errors.
func (c *clientConn) Close() error {
	if c.closer != nil {
		defer c.closer()
	}

	return c.rw.Close()
}

// LocalAddr returns the local network address, if known.
func (c *clientConn) LocalAddr() net.Addr {
	return c.src
}

// RemoteAddr returns the remote network address, if known.
func (c *clientConn) RemoteAddr() net.Addr {
	return c.dst
}
func (c *clientConn) SetDeadline(t time.Time) error {
	return nil
}

func (c *clientConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *clientConn) SetWriteDeadline(t time.Time) error {
	return nil
}
