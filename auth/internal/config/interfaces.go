package config

import "time"

type LoggerConfig interface {
	Level() string
	AsJSON() bool
}

type AuthGRPCConfig interface {
	Address() string
}

type PostgresConfig interface {
	URI() string
}

type RedisConfig interface {
	Address() string
	ConnectionTimeout() time.Duration
	MaxIdle() int
	IdleTimeout() time.Duration
}

type SessionConfig interface {
	// TTL is the lifetime of a user session.
	TTL() time.Duration
}
