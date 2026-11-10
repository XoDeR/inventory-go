package repository

import (
	"time"

	"context"

	"inventory-go/inventory/internal/model"
)

type InventoryRepository interface {
	Get(ctx context.Context, uuid string) (model.Part, error)
	List(ctx context.Context, filter model.PartsFilter) ([]model.Part, error)
}

// InventoryCacheRepository is a cache of the parts. It knows nothing about the database:
// the cache-aside logic lives in the service layer.
type InventoryCacheRepository interface {
	// Get returns model.ErrPartNotFound when the part is not cached.
	Get(ctx context.Context, uuid string) (model.Part, error)
	Set(ctx context.Context, part model.Part, ttl time.Duration) error
	Delete(ctx context.Context, uuid string) error
}
