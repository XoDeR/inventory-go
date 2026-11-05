package consumer

import (
	"context"
	"errors"

	"github.com/IBM/sarama"

	"inventory-go/platform/pkg/kafka"
)

// Logger is the logger used by the consumer. It is satisfied by platform/pkg/logger.Logger.
type Logger interface {
	Info(ctx context.Context, msg string, args ...any)
	Error(ctx context.Context, msg string, args ...any)
}

type consumer struct {
	group       sarama.ConsumerGroup
	topics      []string
	logger      Logger
	middlewares []Middleware
}

// NewConsumer creates a new consumer.
func NewConsumer(group sarama.ConsumerGroup, topics []string, logger Logger, middlewares ...Middleware) *consumer {
	return &consumer{
		group:       group,
		topics:      topics,
		logger:      logger,
		middlewares: middlewares,
	}
}

// Consume starts the consumer for the list of topics. It blocks until the context
// is cancelled, the consumer group is closed or an unrecoverable error occurs.
func (c *consumer) Consume(ctx context.Context, handler kafka.MessageHandler) error {
	newGroupHandler := NewGroupHandler(handler, c.logger, c.middlewares...)

	for {
		if err := c.group.Consume(ctx, c.topics, newGroupHandler); err != nil {
			if errors.Is(err, sarama.ErrClosedConsumerGroup) {
				return nil
			}

			c.logger.Error(ctx, "Kafka consume error", "error", err)
			return err
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}

		c.logger.Info(ctx, "Kafka consumer group rebalancing...")
	}
}
