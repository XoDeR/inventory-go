package env

import (
	"time"

	"github.com/IBM/sarama"
	"github.com/caarlos0/env/v11"
)

type kafkaEnvConfig struct {
	Brokers []string `env:"KAFKA_BROKERS,required" envSeparator:","`
}

type kafkaConfig struct {
	raw kafkaEnvConfig
}

func NewKafkaConfig() (*kafkaConfig, error) {
	var raw kafkaEnvConfig
	if err := env.Parse(&raw); err != nil {
		return nil, err
	}

	return &kafkaConfig{raw: raw}, nil
}

func (cfg *kafkaConfig) Brokers() []string {
	return cfg.raw.Brokers
}

// ---------------------------------------------------------------------------

type orderPaidConsumerEnvConfig struct {
	TopicName string `env:"ORDER_PAID_TOPIC_NAME,required"`
	GroupID   string `env:"ORDER_PAID_CONSUMER_GROUP_ID,required"`
}

type orderPaidConsumerConfig struct {
	raw orderPaidConsumerEnvConfig
}

func NewOrderPaidConsumerConfig() (*orderPaidConsumerConfig, error) {
	var raw orderPaidConsumerEnvConfig
	if err := env.Parse(&raw); err != nil {
		return nil, err
	}

	return &orderPaidConsumerConfig{raw: raw}, nil
}

func (cfg *orderPaidConsumerConfig) Topic() string {
	return cfg.raw.TopicName
}

func (cfg *orderPaidConsumerConfig) GroupID() string {
	return cfg.raw.GroupID
}

// Config returns the sarama config of a consumer group.
func (cfg *orderPaidConsumerConfig) Config() *sarama.Config {
	config := sarama.NewConfig()
	config.Version = sarama.V3_9_0_0
	config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	return config
}

// ---------------------------------------------------------------------------

type computerAssembledProducerEnvConfig struct {
	TopicName string `env:"COMPUTER_ASSEMBLED_TOPIC_NAME,required"`
}

type computerAssembledProducerConfig struct {
	raw computerAssembledProducerEnvConfig
}

func NewComputerAssembledProducerConfig() (*computerAssembledProducerConfig, error) {
	var raw computerAssembledProducerEnvConfig
	if err := env.Parse(&raw); err != nil {
		return nil, err
	}

	return &computerAssembledProducerConfig{raw: raw}, nil
}

func (cfg *computerAssembledProducerConfig) Topic() string {
	return cfg.raw.TopicName
}

// Config returns the sarama config of a sync producer.
func (cfg *computerAssembledProducerConfig) Config() *sarama.Config {
	config := sarama.NewConfig()
	config.Version = sarama.V3_9_0_0
	config.Producer.Return.Successes = true // required by the SyncProducer
	config.Producer.RequiredAcks = sarama.WaitForAll

	return config
}

// ---------------------------------------------------------------------------

type assemblyEnvConfig struct {
	Duration time.Duration `env:"ASSEMBLY_DURATION" envDefault:"10s"`
}

type assemblyConfig struct {
	raw assemblyEnvConfig
}

func NewAssemblyConfig() (*assemblyConfig, error) {
	var raw assemblyEnvConfig
	if err := env.Parse(&raw); err != nil {
		return nil, err
	}

	return &assemblyConfig{raw: raw}, nil
}

// Duration is how long the assembly of a computer takes.
func (cfg *assemblyConfig) Duration() time.Duration {
	return cfg.raw.Duration
}
