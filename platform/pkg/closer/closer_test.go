package closer

import (
	"context"
	"errors"
	"testing"

	"inventory-go/platform/pkg/logger"
)

func TestCloseAllFirstErrorAndOnce(t *testing.T) {
	c := NewWithLogger(logger.NoopLogger{})

	wantErr := errors.New("boom")
	calls := 0

	c.Add(func(context.Context) error { calls++; return nil })
	c.AddNamed("failing", func(context.Context) error { return wantErr })

	if err := c.CloseAll(context.Background()); !errors.Is(err, wantErr) {
		t.Fatalf("CloseAll() = %v, want %v", err, wantErr)
	}
	if calls != 1 {
		t.Errorf("close function ran %d times, want 1", calls)
	}

	// the second call is a no-op
	if err := c.CloseAll(context.Background()); err != nil {
		t.Errorf("second CloseAll() = %v, want nil", err)
	}
}

func TestCloseAllRecoversPanic(t *testing.T) {
	c := NewWithLogger(logger.NoopLogger{})
	c.Add(func(context.Context) error { panic("oops") })

	if err := c.CloseAll(context.Background()); err == nil {
		t.Fatal("CloseAll() = nil, want an error for the panic")
	}
}
