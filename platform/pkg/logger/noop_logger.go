package logger

import "context"

// NoopLogger discards every record.
type NoopLogger struct{}

func (NoopLogger) Debug(context.Context, string, ...any) {}
func (NoopLogger) Info(context.Context, string, ...any)  {}
func (NoopLogger) Warn(context.Context, string, ...any)  {}
func (NoopLogger) Error(context.Context, string, ...any) {}
