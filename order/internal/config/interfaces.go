package config

type OrderHTTPConfig interface {
	Address() string
}

type InventoryClientConfig interface {
	Address() string
}

type PaymentClientConfig interface {
	Address() string
}

type PostgresConfig interface {
	URI() string
}
