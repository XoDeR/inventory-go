package config

import "time"

type LoggerConfig interface {
	Level() string
	AsJSON() bool
}

type InventoryGRPCConfig interface {
	Address() string
}

type InventoryHTTPConfig interface {
	Address() string
}

type MongoConfig interface {
	URI() string
	DatabaseName() string
}

type RedisConfig interface {
	Address() string
	ConnectionTimeout() time.Duration
	MaxIdle() int
	IdleTimeout() time.Duration
	CacheTTL() time.Duration
}
