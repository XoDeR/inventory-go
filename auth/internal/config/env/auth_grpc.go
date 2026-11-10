package env

import (
	"net"

	"github.com/caarlos0/env/v11"
)

type authGRPCEnvConfig struct {
	Host string `env:"GRPC_HOST,required"`
	Port string `env:"GRPC_PORT,required"`
}

type authGRPCConfig struct {
	raw authGRPCEnvConfig
}

func NewAuthGRPCConfig() (*authGRPCConfig, error) {
	var raw authGRPCEnvConfig
	if err := env.Parse(&raw); err != nil {
		return nil, err
	}

	return &authGRPCConfig{raw: raw}, nil
}

func (cfg *authGRPCConfig) Address() string {
	return net.JoinHostPort(cfg.raw.Host, cfg.raw.Port)
}
