package logger

import (
	"context"
	"log/slog"
	"testing"
)

func TestInit(t *testing.T) {
	prev := slog.Default()
	t.Cleanup(func() { slog.SetDefault(prev) })

	ctx := context.Background()

	if err := Init("debug", true); err != nil {
		t.Fatalf("Init(debug, json): %v", err)
	}
	if !slog.Default().Enabled(ctx, slog.LevelDebug) {
		t.Error("debug level must be enabled")
	}

	if err := Init("WARN", false); err != nil {
		t.Fatalf("Init(WARN, text): %v", err)
	}
	if slog.Default().Enabled(ctx, slog.LevelInfo) {
		t.Error("info level must be disabled at warn")
	}
	if !slog.Default().Enabled(ctx, slog.LevelWarn) {
		t.Error("warn level must be enabled")
	}

	if err := Init("verbose", false); err == nil {
		t.Error("an unknown level must be rejected")
	}
}
