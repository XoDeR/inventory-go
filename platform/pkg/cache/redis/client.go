package redis

import (
	"context"
	"time"

	redigo "github.com/gomodule/redigo/redis"

	"inventory-go/platform/pkg/cache"
)

var _ cache.RedisClient = (*client)(nil)

type client struct {
	pool              *redigo.Pool
	logger            Logger
	connectionTimeout time.Duration
}

// Logger is the logger the client reports connection problems to.
type Logger interface {
	Info(ctx context.Context, msg string, args ...any)
	Error(ctx context.Context, msg string, args ...any)
}

type redisFn func(ctx context.Context, conn redigo.Conn) error

// NewClient creates a Redis client on top of a connection pool.
// connectionTimeout limits the time spent waiting for a connection from the pool.
func NewClient(pool *redigo.Pool, logger Logger, connectionTimeout time.Duration) *client {
	return &client{
		pool:              pool,
		logger:            logger,
		connectionTimeout: connectionTimeout,
	}
}

// NewPool creates a redigo connection pool for the given address.
// Connections are dialed lazily, so an unavailable Redis does not fail the creation.
func NewPool(address string, maxIdle int, idleTimeout, dialTimeout time.Duration) *redigo.Pool {
	return &redigo.Pool{
		MaxIdle:     maxIdle,
		IdleTimeout: idleTimeout,
		DialContext: func(ctx context.Context) (redigo.Conn, error) {
			return redigo.DialContext(ctx, "tcp", address, redigo.DialConnectTimeout(dialTimeout))
		},
	}
}

func (c *client) withConn(ctx context.Context, fn redisFn) error {
	conn, err := c.getConn(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := conn.Close(); cerr != nil {
			c.logger.Error(ctx, "failed to close the redis connection", "error", cerr)
		}
	}()

	return fn(ctx, conn)
}

func (c *client) getConn(ctx context.Context) (redigo.Conn, error) {
	ctx, cancel := context.WithTimeout(ctx, c.connectionTimeout)
	defer cancel()

	conn, err := c.pool.GetContext(ctx)
	if err != nil {
		c.logger.Error(ctx, "failed to get a redis connection", "error", err)
		return nil, err
	}

	return conn, nil
}

func (c *client) Set(ctx context.Context, key string, value any) error {
	return c.withConn(ctx, func(_ context.Context, conn redigo.Conn) error {
		_, err := conn.Do("SET", key, value)
		return err
	})
}

func (c *client) SetWithTTL(ctx context.Context, key string, value any, ttl time.Duration) error {
	return c.withConn(ctx, func(_ context.Context, conn redigo.Conn) error {
		_, err := conn.Do("SET", key, value, "PX", ttl.Milliseconds())
		return err
	})
}

func (c *client) Get(ctx context.Context, key string) ([]byte, error) {
	var result []byte
	err := c.withConn(ctx, func(_ context.Context, conn redigo.Conn) error {
		val, err := redigo.Bytes(conn.Do("GET", key))
		if err != nil {
			return err
		}
		result = val
		return nil
	})

	return result, err
}

func (c *client) HashSet(ctx context.Context, key string, values any) error {
	return c.withConn(ctx, func(_ context.Context, conn redigo.Conn) error {
		_, err := conn.Do("HSET", redigo.Args{key}.AddFlat(values)...)
		return err
	})
}

func (c *client) HGetAll(ctx context.Context, key string) ([]any, error) {
	var values []any
	err := c.withConn(ctx, func(_ context.Context, conn redigo.Conn) error {
		result, err := redigo.Values(conn.Do("HGETALL", key))
		if err != nil {
			return err
		}
		values = result
		return nil
	})

	return values, err
}

func (c *client) Del(ctx context.Context, key string) error {
	return c.withConn(ctx, func(_ context.Context, conn redigo.Conn) error {
		_, err := conn.Do("DEL", key)
		return err
	})
}

func (c *client) Exists(ctx context.Context, key string) (bool, error) {
	var exists bool
	err := c.withConn(ctx, func(_ context.Context, conn redigo.Conn) error {
		val, err := redigo.Bool(conn.Do("EXISTS", key))
		if err != nil {
			return err
		}
		exists = val
		return nil
	})

	return exists, err
}

func (c *client) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return c.withConn(ctx, func(_ context.Context, conn redigo.Conn) error {
		_, err := conn.Do("PEXPIRE", key, expiration.Milliseconds())
		return err
	})
}

func (c *client) Ping(ctx context.Context) error {
	return c.withConn(ctx, func(_ context.Context, conn redigo.Conn) error {
		_, err := conn.Do("PING")
		return err
	})
}
