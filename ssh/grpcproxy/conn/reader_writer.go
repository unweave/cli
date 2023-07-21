package conn

import (
	"errors"
	"fmt"
	"io"
	"sync"
)

// MaxChunkSize is the maximum number of bytes to send in a single data message.
const MaxChunkSize int = 1024 * 16

// ReadWriter wraps a grpc source with an [io.ReadWriter] interface.
// All reads are consumed from [Source.Recv] and all writes and sent
// via [Source.Send].
type ReadWriter struct {
	source Source

	wLock  sync.Mutex
	rLock  sync.Mutex
	rBytes []byte
}

type Source interface {
	Send([]byte) error
	Recv() ([]byte, error)
}

// NewReadWriter creates a new ReadWriter that leverages the provided
// source to retrieve data from and write data to.
func NewReadWriter(source Source) *ReadWriter {
	return &ReadWriter{source: source}
}

// Read returns data received from the stream source. Any
// data received from the stream that is not consumed will
// be buffered and returned on subsequent reads until there
// is none left. Only then will data be sourced from the stream
// again.
func (c *ReadWriter) Read(b []byte) (n int, err error) {
	c.rLock.Lock()
	defer c.rLock.Unlock()

	if len(c.rBytes) == 0 {
		data, err := c.source.Recv()
		if errors.Is(err, io.EOF) {
			return 0, io.EOF
		}
		if err != nil {
			return 0, fmt.Errorf("failed to receive from source: %w", err)
		}

		if data == nil {
			return 0, errors.New("received invalid data from source")
		}

		c.rBytes = data
	}

	n = copy(b, c.rBytes)
	c.rBytes = c.rBytes[n:]

	// Stop holding onto buffer immediately
	if len(c.rBytes) == 0 {
		c.rBytes = nil
	}

	return n, nil
}

// Write consumes all data provided and sends it on
// the grpc stream. To prevent exhausting the stream all
// sends on the stream are limited to be at most MaxChunkSize.
// If the data exceeds the MaxChunkSize it will be sent in
// batches.
func (c *ReadWriter) Write(b []byte) (int, error) {
	c.wLock.Lock()
	defer c.wLock.Unlock()

	var sent int
	for len(b) > 0 {
		chunk := b
		if len(chunk) > MaxChunkSize {
			chunk = chunk[:MaxChunkSize]
		}

		if err := c.source.Send(chunk); err != nil {
			return sent, fmt.Errorf("failed to send on source: %w", err)
		}

		sent += len(chunk)
		b = b[len(chunk):]
	}

	return sent, nil
}

// Close cleans up resources used by the stream.
func (c *ReadWriter) Close() error {
	if cs, ok := c.source.(io.Closer); ok {
		c.wLock.Lock()
		defer c.wLock.Unlock()
		return cs.Close()
	}

	return nil
}
