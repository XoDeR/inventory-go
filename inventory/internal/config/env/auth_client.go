package env

import (
	"net"

	"github.com/caarlos0/env/v11"
)

type authClientEnvConfig struct {
	Host string `env:"AUTH_GRPC_HOST,required"`
	Port string `env:"AUTH_GRPC_PORT,required"`
}

type authClientConfig struct {
	raw authClientEnvConfig
}

func NewAuthClientConfig() (*authClientConfig, error) {
	var raw authClientEnvConfig
	if err := env.Parse(&raw); err != nil {
		return nil, err
	}

	return &authClientConfig{raw: raw}, nil
}

func (cfg *authClientConfig) Address() string {
	return net.JoinHostPort(cfg.raw.Host, cfg.raw.Port)
}
