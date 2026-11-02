package logger

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
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

	if err := SetLevel("info"); err != nil {
		t.Fatalf("SetLevel(info): %v", err)
	}
	if !slog.Default().Enabled(ctx, slog.LevelInfo) {
		t.Error("info level must be enabled after SetLevel")
	}

	if err := Init("verbose", false); err == nil {
		t.Error("an unknown level must be rejected")
	}
}

func TestContextAttrs(t *testing.T) {
	var buf bytes.Buffer
	h := &contextHandler{Handler: slog.NewTextHandler(&buf, nil)}
	l := New(slog.New(h))

	ctx := WithUserID(WithTraceID(context.Background(), "t-1"), "u-2")
	l.Info(ctx, "hello", "k", "v")

	out := buf.String()
	for _, want := range []string{"msg=hello", "k=v", "trace_id=t-1", "user_id=u-2"} {
		if !strings.Contains(out, want) {
			t.Errorf("output %q misses %q", out, want)
		}
	}
}

func TestNoopLogger(t *testing.T) {
	var l NoopLogger
	l.Info(context.Background(), "ignored")
	l.Error(context.Background(), "ignored", "k", "v")
}
