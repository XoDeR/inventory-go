package closer

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"

	"inventory-go/platform/pkg/logger"
)

// shutdownTimeout is the default time given to the shutdown on a system signal.
const shutdownTimeout = 5 * time.Second

// Logger is the logging contract of Closer. It matches the context-first,
// slog-style methods of logger.Logger.
type Logger interface {
	Info(ctx context.Context, msg string, args ...any)
	Error(ctx context.Context, msg string, args ...any)
}

// Closer drives the graceful shutdown of the application.
type Closer struct {
	mu     sync.Mutex                    // guards funcs and logger
	once   sync.Once                     // makes CloseAll run only once
	done   chan struct{}                 // closed when CloseAll has finished
	funcs  []func(context.Context) error // registered close functions
	logger Logger
}

// globalCloser is the instance shared by the whole application.
var globalCloser = NewWithLogger(logger.NoopLogger{})

// AddNamed registers a named close function on the global closer.
func AddNamed(name string, f func(context.Context) error) {
	globalCloser.AddNamed(name, f)
}

// Add registers close functions on the global closer.
func Add(f ...func(context.Context) error) {
	globalCloser.Add(f...)
}

// CloseAll runs all close functions registered on the global closer.
func CloseAll(ctx context.Context) error {
	return globalCloser.CloseAll(ctx)
}

// SetLogger sets the logger of the global closer.
func SetLogger(l Logger) {
	globalCloser.SetLogger(l)
}

// Configure makes the global closer react to the given system signals.
func Configure(signals ...os.Signal) {
	go globalCloser.handleSignals(signals...)
}

// New creates a Closer that logs with the default slog logger.
func New(signals ...os.Signal) *Closer {
	return NewWithLogger(logger.Default(), signals...)
}

// NewWithLogger creates a Closer with the given logger.
// If signals are passed, the Closer listens to them and calls CloseAll on receipt.
func NewWithLogger(l Logger, signals ...os.Signal) *Closer {
	c := &Closer{
		done:   make(chan struct{}),
		logger: l,
	}

	if len(signals) > 0 {
		go c.handleSignals(signals...)
	}

	return c
}

// SetLogger sets the logger of the Closer.
func (c *Closer) SetLogger(l Logger) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.logger = l
}

func (c *Closer) log() Logger {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.logger
}

// handleSignals waits for a signal and calls CloseAll with a fresh shutdown context.
func (c *Closer) handleSignals(signals ...os.Signal) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, signals...)
	defer signal.Stop(ch)

	select {
	case <-ch:
		c.log().Info(context.Background(), "received a system signal, starting graceful shutdown")

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer shutdownCancel()

		if err := c.CloseAll(shutdownCtx); err != nil {
			c.log().Error(context.Background(), "failed to close resources", "error", err)
		}

	case <-c.done:
		// CloseAll was already called manually, nothing to do
	}
}

// AddNamed registers a close function and logs its name, outcome and duration.
func (c *Closer) AddNamed(name string, f func(context.Context) error) {
	c.Add(func(ctx context.Context) error {
		start := time.Now()
		c.log().Info(ctx, fmt.Sprintf("closing %s", name))

		err := f(ctx)

		duration := time.Since(start)
		if err != nil {
			c.log().Error(ctx, fmt.Sprintf("failed to close %s", name), "error", err, "duration", duration)
		} else {
			c.log().Info(ctx, fmt.Sprintf("%s closed", name), "duration", duration)
		}

		return err
	})
}

// Add registers one or more close functions.
func (c *Closer) Add(f ...func(context.Context) error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.funcs = append(c.funcs, f...)
}

// CloseAll calls all registered close functions concurrently and returns the first error, if any.
func (c *Closer) CloseAll(ctx context.Context) error {
	var result error

	c.once.Do(func() {
		defer close(c.done)

		c.mu.Lock()
		funcs := c.funcs
		c.funcs = nil // release the memory
		c.mu.Unlock()

		if len(funcs) == 0 {
			c.log().Info(ctx, "no close functions registered")
			return
		}

		c.log().Info(ctx, "starting graceful shutdown")

		errCh := make(chan error, len(funcs))
		var wg sync.WaitGroup

		// Start in the reverse order of registration
		for i := len(funcs) - 1; i >= 0; i-- {
			f := funcs[i]
			wg.Add(1)
			go func(f func(context.Context) error) {
				defer wg.Done()

				// Protect against panics
				defer func() {
					if r := recover(); r != nil {
						errCh <- errors.New("panic recovered in closer")
						c.log().Error(ctx, "panic in close function", "panic", r)
					}
				}()

				if err := f(ctx); err != nil {
					errCh <- err
				}
			}(f)
		}

		// Close the error channel once all functions have finished
		go func() {
			wg.Wait()
			close(errCh)
		}()

		// Collect errors or wait for the context cancellation
		for {
			select {
			case <-ctx.Done():
				c.log().Info(ctx, "context cancelled during shutdown", "error", ctx.Err())
				if result == nil {
					result = ctx.Err()
				}
				return
			case err, ok := <-errCh:
				if !ok {
					c.log().Info(ctx, "all resources closed")
					return
				}
				c.log().Error(ctx, "failed to close", "error", err)
				if result == nil {
					result = err
				}
			}
		}
	})

	return result
}
