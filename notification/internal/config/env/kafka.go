package env

import (
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

func consumerGroupConfig() *sarama.Config {
	config := sarama.NewConfig()
	config.Version = sarama.V3_9_0_0
	config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	return config
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

func (cfg *orderPaidConsumerConfig) Topic() string          { return cfg.raw.TopicName }
func (cfg *orderPaidConsumerConfig) GroupID() string        { return cfg.raw.GroupID }
func (cfg *orderPaidConsumerConfig) Config() *sarama.Config { return consumerGroupConfig() }

// ---------------------------------------------------------------------------

type computerAssembledConsumerEnvConfig struct {
	TopicName string `env:"COMPUTER_ASSEMBLED_TOPIC_NAME,required"`
	GroupID   string `env:"COMPUTER_ASSEMBLED_CONSUMER_GROUP_ID,required"`
}

type computerAssembledConsumerConfig struct {
	raw computerAssembledConsumerEnvConfig
}

func NewComputerAssembledConsumerConfig() (*computerAssembledConsumerConfig, error) {
	var raw computerAssembledConsumerEnvConfig
	if err := env.Parse(&raw); err != nil {
		return nil, err
	}

	return &computerAssembledConsumerConfig{raw: raw}, nil
}

func (cfg *computerAssembledConsumerConfig) Topic() string          { return cfg.raw.TopicName }
func (cfg *computerAssembledConsumerConfig) GroupID() string        { return cfg.raw.GroupID }
func (cfg *computerAssembledConsumerConfig) Config() *sarama.Config { return consumerGroupConfig() }
