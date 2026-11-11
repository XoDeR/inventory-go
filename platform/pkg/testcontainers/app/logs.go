package app

import (
	"io"
	"sync"

	"github.com/testcontainers/testcontainers-go"
)

// logConsumer writes the container output to a writer. Stdout and stderr arrive from
// different goroutines, so writes are serialized.
type logConsumer struct {
	mu  sync.Mutex
	out io.Writer
}

func newLogConsumer(out io.Writer) *logConsumer {
	return &logConsumer{out: out}
}

func (c *logConsumer) Accept(l testcontainers.Log) {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, _ = c.out.Write(l.Content)
}
