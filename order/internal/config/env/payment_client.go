package env

import (
	"net"

	"github.com/caarlos0/env/v11"
)

type paymentClientEnvConfig struct {
	Host string `env:"PAYMENT_GRPC_HOST,required"`
	Port string `env:"PAYMENT_GRPC_PORT,required"`
}

type paymentClientConfig struct {
	raw paymentClientEnvConfig
}

func NewPaymentClientConfig() (*paymentClientConfig, error) {
	var raw paymentClientEnvConfig
	if err := env.Parse(&raw); err != nil {
		return nil, err
	}

	return &paymentClientConfig{raw: raw}, nil
}

func (cfg *paymentClientConfig) Address() string {
	return net.JoinHostPort(cfg.raw.Host, cfg.raw.Port)
}
