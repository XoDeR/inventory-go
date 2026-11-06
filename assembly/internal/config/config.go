package config

import (
	"os"

	"github.com/joho/godotenv"

	"inventory-go/assembly/internal/config/env"
)

var appConfig *config

type config struct {
	Logger                    LoggerConfig
	Kafka                     KafkaConfig
	OrderPaidConsumer         OrderPaidConsumerConfig
	ComputerAssembledProducer ComputerAssembledProducerConfig
	Assembly                  AssemblyConfig
}

// Load reads the given .env files (a missing file is not an error, the process
// environment is used then) and parses the config from the environment.
func Load(path ...string) error {
	err := godotenv.Load(path...)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	loggerCfg, err := env.NewLoggerConfig()
	if err != nil {
		return err
	}

	kafkaCfg, err := env.NewKafkaConfig()
	if err != nil {
		return err
	}

	orderPaidConsumerCfg, err := env.NewOrderPaidConsumerConfig()
	if err != nil {
		return err
	}

	computerAssembledProducerCfg, err := env.NewComputerAssembledProducerConfig()
	if err != nil {
		return err
	}

	assemblyCfg, err := env.NewAssemblyConfig()
	if err != nil {
		return err
	}

	appConfig = &config{
		Logger:                    loggerCfg,
		Kafka:                     kafkaCfg,
		OrderPaidConsumer:         orderPaidConsumerCfg,
		ComputerAssembledProducer: computerAssembledProducerCfg,
		Assembly:                  assemblyCfg,
	}

	return nil
}

func AppConfig() *config {
	return appConfig
}
