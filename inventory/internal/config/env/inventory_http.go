package env

import (
	"net"

	"github.com/caarlos0/env/v11"
)

type inventoryHTTPEnvConfig struct {
	Host string `env:"HTTP_HOST,required"`
	Port string `env:"HTTP_PORT,required"`
}

type inventoryHTTPConfig struct {
	raw inventoryHTTPEnvConfig
}

func NewInventoryHTTPConfig() (*inventoryHTTPConfig, error) {
	var raw inventoryHTTPEnvConfig
	if err := env.Parse(&raw); err != nil {
		return nil, err
	}

	return &inventoryHTTPConfig{raw: raw}, nil
}

func (cfg *inventoryHTTPConfig) Address() string {
	return net.JoinHostPort(cfg.raw.Host, cfg.raw.Port)
}
