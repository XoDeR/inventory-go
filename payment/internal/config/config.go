package config

import (
	"os"

	"github.com/joho/godotenv"

	"inventory-go/payment/internal/config/env"
)

var appConfig *config

type config struct {
	GRPC PaymentGRPCConfig
	HTTP PaymentHTTPConfig
}

// Load reads the given .env files (a missing file is not an error, the process
// environment is used then) and parses the config from the environment.
func Load(path ...string) error {
	err := godotenv.Load(path...)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	grpcCfg, err := env.NewPaymentGRPCConfig()
	if err != nil {
		return err
	}

	httpCfg, err := env.NewPaymentHTTPConfig()
	if err != nil {
		return err
	}

	appConfig = &config{
		GRPC: grpcCfg,
		HTTP: httpCfg,
	}

	return nil
}

func AppConfig() *config {
	return appConfig
}
