// Package logger configures the process-wide slog logger and offers a thin
// slog-shaped wrapper that takes a context as the first argument.
package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
)

type ctxKey string

const (
	traceIDKey ctxKey = "trace_id"
	userIDKey  ctxKey = "user_id"
)

// levelVar allows changing the level of the default logger at runtime.
var levelVar slog.LevelVar

// Init builds a slog logger that writes to stdout and installs it as the default one.
// level is one of debug, info, warn, error (case-insensitive); asJSON picks the JSON
// handler over the human-readable text one. Records logged with a context carry the
// trace_id and user_id attached via WithTraceID and WithUserID.
func Init(level string, asJSON bool) error {
	var lvl slog.Level
	if err := lvl.UnmarshalText([]byte(level)); err != nil {
		return fmt.Errorf("invalid log level %q: %w", level, err)
	}

	levelVar.Set(lvl)

	opts := &slog.HandlerOptions{Level: &levelVar}

	var handler slog.Handler
	if asJSON {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	slog.SetDefault(slog.New(&contextHandler{Handler: handler}))

	return nil
}

// SetLevel changes the level of the logger installed by Init.
func SetLevel(level string) error {
	var lvl slog.Level
	if err := lvl.UnmarshalText([]byte(level)); err != nil {
		return fmt.Errorf("invalid log level %q: %w", level, err)
	}

	levelVar.Set(lvl)

	return nil
}

// WithTraceID returns a context whose log records carry the given trace_id.
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

// WithUserID returns a context whose log records carry the given user_id.
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// contextHandler adds the context-bound attributes to every record.
type contextHandler struct {
	slog.Handler
}

func (h *contextHandler) Handle(ctx context.Context, r slog.Record) error {
	if traceID, ok := ctx.Value(traceIDKey).(string); ok && traceID != "" {
		r.AddAttrs(slog.String(string(traceIDKey), traceID))
	}

	if userID, ok := ctx.Value(userIDKey).(string); ok && userID != "" {
		r.AddAttrs(slog.String(string(userIDKey), userID))
	}

	return h.Handler.Handle(ctx, r)
}

func (h *contextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &contextHandler{Handler: h.Handler.WithAttrs(attrs)}
}

func (h *contextHandler) WithGroup(name string) slog.Handler {
	return &contextHandler{Handler: h.Handler.WithGroup(name)}
}

// Logger wraps *slog.Logger with context-first methods. Arguments follow the
// slog convention: alternating keys and values, or slog.Attr values.
type Logger struct {
	l *slog.Logger
}

// New wraps the given slog logger.
func New(l *slog.Logger) *Logger {
	return &Logger{l: l}
}

// Default wraps the current slog default logger.
func Default() *Logger {
	return New(slog.Default())
}

// With returns a logger that adds the given arguments to every record.
func (l *Logger) With(args ...any) *Logger {
	return &Logger{l: l.l.With(args...)}
}

// WithGroup returns a logger that nests the following attributes under name.
func (l *Logger) WithGroup(name string) *Logger {
	return &Logger{l: l.l.WithGroup(name)}
}

// Slog returns the underlying slog logger.
func (l *Logger) Slog() *slog.Logger {
	return l.l
}

func (l *Logger) Debug(ctx context.Context, msg string, args ...any) {
	l.l.Log(ctx, slog.LevelDebug, msg, args...)
}

func (l *Logger) Info(ctx context.Context, msg string, args ...any) {
	l.l.Log(ctx, slog.LevelInfo, msg, args...)
}

func (l *Logger) Warn(ctx context.Context, msg string, args ...any) {
	l.l.Log(ctx, slog.LevelWarn, msg, args...)
}

func (l *Logger) Error(ctx context.Context, msg string, args ...any) {
	l.l.Log(ctx, slog.LevelError, msg, args...)
}

// Debug logs with the default logger.
func Debug(ctx context.Context, msg string, args ...any) { Default().Debug(ctx, msg, args...) }

// Info logs with the default logger.
func Info(ctx context.Context, msg string, args ...any) { Default().Info(ctx, msg, args...) }

// Warn logs with the default logger.
func Warn(ctx context.Context, msg string, args ...any) { Default().Warn(ctx, msg, args...) }

// Error logs with the default logger.
func Error(ctx context.Context, msg string, args ...any) { Default().Error(ctx, msg, args...) }
