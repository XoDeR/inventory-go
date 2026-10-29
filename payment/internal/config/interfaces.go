package config

type LoggerConfig interface {
	Level() string
	AsJSON() bool
}

type PaymentGRPCConfig interface {
	Address() string
}

type PaymentHTTPConfig interface {
	Address() string
}
