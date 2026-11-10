package order_cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	redigo "github.com/gomodule/redigo/redis"

	"inventory-go/order/internal/model"
	def "inventory-go/order/internal/repository"
	"inventory-go/platform/pkg/cache"
)

const cacheKeyPrefix = "order:"

var _ def.OrderCacheRepository = (*repository)(nil)

type repository struct {
	cache cache.RedisClient
}

func NewRepository(cache cache.RedisClient) *repository {
	return &repository{cache: cache}
}

func cacheKey(uuid string) string {
	return cacheKeyPrefix + uuid
}

func (r *repository) Get(ctx context.Context, uuid string) (model.Order, error) {
	data, err := r.cache.Get(ctx, cacheKey(uuid))
	if err != nil {
		if errors.Is(err, redigo.ErrNil) {
			return model.Order{}, model.ErrOrderNotFound
		}

		return model.Order{}, err
	}

	var order model.Order
	if err = json.Unmarshal(data, &order); err != nil {
		return model.Order{}, fmt.Errorf("failed to decode the cached order: %w", err)
	}

	return order, nil
}

func (r *repository) Set(ctx context.Context, order model.Order, ttl time.Duration) error {
	data, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to encode the order: %w", err)
	}

	return r.cache.SetWithTTL(ctx, cacheKey(order.Uuid.String()), data, ttl)
}

func (r *repository) Delete(ctx context.Context, uuid string) error {
	return r.cache.Del(ctx, cacheKey(uuid))
}
