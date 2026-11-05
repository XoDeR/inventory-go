package consumer

import (
	"context"
	"testing"

	"inventory-go/platform/pkg/kafka"
	"inventory-go/platform/pkg/logger"
)

// The platform logger must satisfy the consumer Logger interface.
var _ Logger = (*logger.Logger)(nil)

func TestNewGroupHandler_MiddlewareOrder(t *testing.T) {
	var calls []string

	mw := func(name string) Middleware {
		return func(next kafka.MessageHandler) kafka.MessageHandler {
			return func(ctx context.Context, msg kafka.Message) error {
				calls = append(calls, name)
				return next(ctx, msg)
			}
		}
	}

	h := NewGroupHandler(func(context.Context, kafka.Message) error {
		calls = append(calls, "handler")
		return nil
	}, logger.Default(), mw("first"), mw("second"))

	if err := h.handler(context.Background(), kafka.Message{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{"first", "second", "handler"}
	if len(calls) != len(want) {
		t.Fatalf("got %v, want %v", calls, want)
	}
	for i := range want {
		if calls[i] != want[i] {
			t.Fatalf("got %v, want %v", calls, want)
		}
	}
}
