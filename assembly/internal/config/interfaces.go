package config

import (
	"time"

	"github.com/IBM/sarama"
)

type LoggerConfig interface {
	Level() string
	AsJSON() bool
}

type KafkaConfig interface {
	Brokers() []string
}

type OrderPaidConsumerConfig interface {
	Topic() string
	GroupID() string
	Config() *sarama.Config
}

type ComputerAssembledProducerConfig interface {
	Topic() string
	Config() *sarama.Config
}

type AssemblyConfig interface {
	Duration() time.Duration
}
