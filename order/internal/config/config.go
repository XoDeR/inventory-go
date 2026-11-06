package config

import (
	"os"

	"github.com/joho/godotenv"

	"inventory-go/order/internal/config/env"
)

var appConfig *config

type config struct {
	Logger                 LoggerConfig
	HTTP                   OrderHTTPConfig
	Inventory              InventoryClientConfig
	Payment                PaymentClientConfig
	Postgres               PostgresConfig
	Kafka                  KafkaConfig
	OrderPaidProducer      OrderPaidProducerConfig
	OrderAssembledConsumer OrderAssembledConsumerConfig
}

// Load reads the given .env files (a missing file is not an error, the process
// environment is used then) and parses the config from the environment.
func Load(path ...string) error {
	err := godotenv.Load(path...)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	httpCfg, err := env.NewOrderHTTPConfig()
	if err != nil {
		return err
	}

	inventoryCfg, err := env.NewInventoryClientConfig()
	if err != nil {
		return err
	}

	paymentCfg, err := env.NewPaymentClientConfig()
	if err != nil {
		return err
	}

	postgresCfg, err := env.NewPostgresConfig()
	if err != nil {
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

	orderPaidProducerCfg, err := env.NewOrderPaidProducerConfig()
	if err != nil {
		return err
	}

	orderAssembledConsumerCfg, err := env.NewOrderAssembledConsumerConfig()
	if err != nil {
		return err
	}

	appConfig = &config{
		Logger:                 loggerCfg,
		HTTP:                   httpCfg,
		Inventory:              inventoryCfg,
		Payment:                paymentCfg,
		Postgres:               postgresCfg,
		Kafka:                  kafkaCfg,
		OrderPaidProducer:      orderPaidProducerCfg,
		OrderAssembledConsumer: orderAssembledConsumerCfg,
	}

	return nil
}

func AppConfig() *config {
	return appConfig
}
