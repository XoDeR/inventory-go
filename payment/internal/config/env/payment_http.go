package env

import (
	"net"

	"github.com/caarlos0/env/v11"
)

type paymentHTTPEnvConfig struct {
	Host string `env:"HTTP_HOST,required"`
	Port string `env:"HTTP_PORT,required"`
}

type paymentHTTPConfig struct {
	raw paymentHTTPEnvConfig
}

func NewPaymentHTTPConfig() (*paymentHTTPConfig, error) {
	var raw paymentHTTPEnvConfig
	if err := env.Parse(&raw); err != nil {
		return nil, err
	}

	return &paymentHTTPConfig{raw: raw}, nil
}

func (cfg *paymentHTTPConfig) Address() string {
	return net.JoinHostPort(cfg.raw.Host, cfg.raw.Port)
}
