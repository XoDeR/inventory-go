package repository

import (
	"time"

	"context"

	"inventory-go/order/internal/model"
)

type OrderRepository interface {
	Create(ctx context.Context, order model.Order) error
	Get(ctx context.Context, uuid string) (model.Order, error)
	Update(ctx context.Context, order model.Order) error
}

// OrderCacheRepository is a cache of the orders. It knows nothing about the database:
// the cache-aside logic lives in the service layer.
type OrderCacheRepository interface {
	// Get returns model.ErrOrderNotFound when the order is not cached.
	Get(ctx context.Context, uuid string) (model.Order, error)
	Set(ctx context.Context, order model.Order, ttl time.Duration) error
	Delete(ctx context.Context, uuid string) error
}
