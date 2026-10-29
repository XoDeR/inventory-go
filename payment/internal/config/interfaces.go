package config

type PaymentGRPCConfig interface {
	Address() string
}

type PaymentHTTPConfig interface {
	Address() string
}
