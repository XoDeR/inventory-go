package kafka

import (
	"context"

	"inventory-go/platform/pkg/kafka"
	"inventory-go/platform/pkg/kafka/consumer"
)

// Logger is the logger used by the middleware. It is satisfied by platform/pkg/logger.Logger.
type Logger interface {
	Info(ctx context.Context, msg string, args ...any)
}

// Logging returns a consumer middleware that logs every received message.
func Logging(logger Logger) consumer.Middleware {
	return func(next kafka.MessageHandler) kafka.MessageHandler {
		return func(ctx context.Context, msg kafka.Message) error {
			logger.Info(ctx, "Kafka msg received", "topic", msg.Topic)
			return next(ctx, msg)
		}
	}
}
