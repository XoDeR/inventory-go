package session

import (
	"context"
	"errors"
	"time"

	redigo "github.com/gomodule/redigo/redis"

	"inventory-go/auth/internal/model"
	def "inventory-go/auth/internal/repository"
	"inventory-go/platform/pkg/cache"
)

const keyPrefix = "auth:session:"

var _ def.SessionRepository = (*repository)(nil)

type repository struct {
	cache cache.RedisClient
}

func NewRepository(cache cache.RedisClient) *repository {
	return &repository{cache: cache}
}

func key(sessionUUID string) string {
	return keyPrefix + sessionUUID
}

func (r *repository) Create(ctx context.Context, sessionUUID, userUUID string, ttl time.Duration) error {
	return r.cache.SetWithTTL(ctx, key(sessionUUID), userUUID, ttl)
}

func (r *repository) Get(ctx context.Context, sessionUUID string) (string, error) {
	data, err := r.cache.Get(ctx, key(sessionUUID))
	if err != nil {
		if errors.Is(err, redigo.ErrNil) {
			return "", model.ErrSessionNotFound
		}

		return "", err
	}

	return string(data), nil
}
