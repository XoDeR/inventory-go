package kafka

import (
	"context"
)

// MessageHandler handles a single Kafka message.
type MessageHandler func(ctx context.Context, msg Message) error

// Consumer reads messages from Kafka and passes them to a handler.
type Consumer interface {
	Consume(ctx context.Context, handler MessageHandler) error
}

// Producer sends messages to Kafka.
type Producer interface {
	Send(ctx context.Context, key, value []byte) error
}
