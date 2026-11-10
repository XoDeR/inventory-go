package part

import (
	"context"
	"errors"

	"inventory-go/inventory/internal/model"
	"inventory-go/platform/pkg/logger"
)

// Get implements cache-aside: the cache is tried first, the database is the fallback
// and fills the cache. Cache failures are logged and never fail the request.
func (s *service) Get(ctx context.Context, uuid string) (model.Part, error) {
	cached, err := s.inventoryCacheRepository.Get(ctx, uuid)
	if err == nil {
		return cached, nil
	}

	if !errors.Is(err, model.ErrPartNotFound) {
		logger.Error(ctx, "failed to read the part from the cache", "uuid", uuid, "error", err)
	}

	part, err := s.inventoryRepository.Get(ctx, uuid)
	if err != nil {
		return model.Part{}, err
	}

	if err = s.inventoryCacheRepository.Set(ctx, part, s.cacheTTL); err != nil {
		logger.Error(ctx, "failed to cache the part", "uuid", uuid, "error", err)
	}

	return part, nil
}
