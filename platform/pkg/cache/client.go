package cache

import (
	"context"
	"time"
)

// RedisClient is the set of Redis operations the services use.
type RedisClient interface {
	Set(ctx context.Context, key string, value any) error
	SetWithTTL(ctx context.Context, key string, value any, ttl time.Duration) error
	// Get returns the value of the key. A missing key is reported as redigo.ErrNil.
	Get(ctx context.Context, key string) ([]byte, error)
	HashSet(ctx context.Context, key string, values any) error
	HGetAll(ctx context.Context, key string) ([]any, error)
	Del(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	Expire(ctx context.Context, key string, expiration time.Duration) error
	Ping(ctx context.Context) error
	SetOperator
}

// SetOperator is the set of Redis operations on the SET data type.
type SetOperator interface {
	SAdd(ctx context.Context, key, value string) error
	SRem(ctx context.Context, key, value string) error
	SIsMember(ctx context.Context, key, value string) (bool, error)
	SMembers(ctx context.Context, key string) ([]string, error)
}
