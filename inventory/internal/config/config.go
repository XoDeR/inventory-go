package config

import (
	"os"

	"github.com/joho/godotenv"

	"inventory-go/inventory/internal/config/env"
)

var appConfig *config

type config struct {
	GRPC  InventoryGRPCConfig
	HTTP  InventoryHTTPConfig
	Mongo MongoConfig
}

// Load reads the given .env files (a missing file is not an error, the process
// environment is used then) and parses the config from the environment.
func Load(path ...string) error {
	err := godotenv.Load(path...)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	grpcCfg, err := env.NewInventoryGRPCConfig()
	if err != nil {
		return err
	}

	httpCfg, err := env.NewInventoryHTTPConfig()
	if err != nil {
		return err
	}

	mongoCfg, err := env.NewMongoConfig()
	if err != nil {
		return err
	}

	appConfig = &config{
		GRPC:  grpcCfg,
		HTTP:  httpCfg,
		Mongo: mongoCfg,
	}

	return nil
}

func AppConfig() *config {
	return appConfig
}
