package producer

import (
	"context"

	"github.com/IBM/sarama"
)

// Logger is the logger used by the producer. It is satisfied by platform/pkg/logger.Logger.
type Logger interface {
	Info(ctx context.Context, msg string, args ...any)
	Error(ctx context.Context, msg string, args ...any)
}

type producer struct {
	syncProducer sarama.SyncProducer
	topic        string
	logger       Logger
}

// NewProducer creates a new producer that sends messages to the given topic.
func NewProducer(syncProducer sarama.SyncProducer, topic string, logger Logger) *producer {
	return &producer{
		syncProducer: syncProducer,
		topic:        topic,
		logger:       logger,
	}
}

// Send sends a message with the given key and value to the producer topic.
func (p *producer) Send(ctx context.Context, key, value []byte) error {
	partition, offset, err := p.syncProducer.SendMessage(&sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.ByteEncoder(key),
		Value: sarama.ByteEncoder(value),
	})
	if err != nil {
		p.logger.Error(ctx, "Failed to send message", "error", err)
		return err
	}

	p.logger.Info(ctx, "Message sent",
		"topic", p.topic,
		"partition", partition,
		"offset", offset,
		"key", string(key),
		"value", string(value),
	)

	return nil
}
