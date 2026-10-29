package env

import (
	"net"

	"github.com/caarlos0/env/v11"
)

type inventoryClientEnvConfig struct {
	Host string `env:"INVENTORY_GRPC_HOST,required"`
	Port string `env:"INVENTORY_GRPC_PORT,required"`
}

type inventoryClientConfig struct {
	raw inventoryClientEnvConfig
}

func NewInventoryClientConfig() (*inventoryClientConfig, error) {
	var raw inventoryClientEnvConfig
	if err := env.Parse(&raw); err != nil {
		return nil, err
	}

	return &inventoryClientConfig{raw: raw}, nil
}

func (cfg *inventoryClientConfig) Address() string {
	return net.JoinHostPort(cfg.raw.Host, cfg.raw.Port)
}
