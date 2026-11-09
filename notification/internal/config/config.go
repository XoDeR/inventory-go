package config

import (
	"os"

	"github.com/joho/godotenv"

	"inventory-go/notification/internal/config/env"
)

var appConfig *config

type config struct {
	Logger                    LoggerConfig
	Kafka                     KafkaConfig
	OrderPaidConsumer         OrderPaidConsumerConfig
	ComputerAssembledConsumer ComputerAssembledConsumerConfig
	SlackBot                  SlackBotConfig
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

	computerAssembledConsumerCfg, err := env.NewComputerAssembledConsumerConfig()
	if err != nil {
		return err
	}

	slackBotCfg, err := env.NewSlackBotConfig()
	if err != nil {
		return err
	}

	appConfig = &config{
		Logger:                    loggerCfg,
		Kafka:                     kafkaCfg,
		OrderPaidConsumer:         orderPaidConsumerCfg,
		ComputerAssembledConsumer: computerAssembledConsumerCfg,
		SlackBot:                  slackBotCfg,
	}

	return nil
}

func AppConfig() *config {
	return appConfig
}
