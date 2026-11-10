package part_cache

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"time"

	redigo "github.com/gomodule/redigo/redis"

	"inventory-go/inventory/internal/model"
	def "inventory-go/inventory/internal/repository"
	"inventory-go/platform/pkg/cache"
)

const cacheKeyPrefix = "inventory:part:"

var _ def.InventoryCacheRepository = (*repository)(nil)

type repository struct {
	cache cache.RedisClient
}

func NewRepository(cache cache.RedisClient) *repository {
	return &repository{cache: cache}
}

func cacheKey(uuid string) string {
	return cacheKeyPrefix + uuid
}

func (r *repository) Get(ctx context.Context, uuid string) (model.Part, error) {
	data, err := r.cache.Get(ctx, cacheKey(uuid))
	if err != nil {
		if errors.Is(err, redigo.ErrNil) {
			return model.Part{}, model.ErrPartNotFound
		}

		return model.Part{}, err
	}

	// gob keeps the exact types of the Metadata values (int64 stays int64, unlike JSON).
	var part model.Part
	if err = gob.NewDecoder(bytes.NewReader(data)).Decode(&part); err != nil {
		return model.Part{}, fmt.Errorf("failed to decode the cached part: %w", err)
	}

	return part, nil
}

func (r *repository) Set(ctx context.Context, part model.Part, ttl time.Duration) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(part); err != nil {
		return fmt.Errorf("failed to encode the part: %w", err)
	}

	return r.cache.SetWithTTL(ctx, cacheKey(part.Uuid), buf.Bytes(), ttl)
}

func (r *repository) Delete(ctx context.Context, uuid string) error {
	return r.cache.Del(ctx, cacheKey(uuid))
}
