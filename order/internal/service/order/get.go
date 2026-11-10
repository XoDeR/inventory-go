package order

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"inventory-go/order/internal/model"
	"inventory-go/platform/pkg/logger"
)

// Get implements cache-aside: the cache is tried first, the database is the fallback
// and fills the cache. Cache failures are logged and never fail the request.
func (s *service) Get(ctx context.Context, uuid uuid.UUID) (model.Order, error) {
	cached, err := s.orderCacheRepository.Get(ctx, uuid.String())
	if err == nil {
		return cached, nil
	}

	if !errors.Is(err, model.ErrOrderNotFound) {
		logger.Error(ctx, "failed to read the order from the cache", "order_uuid", uuid, "error", err)
	}

	order, err := s.orderRepository.Get(ctx, uuid.String())
	if err != nil {
		return model.Order{}, err
	}

	if err = s.orderCacheRepository.Set(ctx, order, s.cacheTTL); err != nil {
		logger.Error(ctx, "failed to cache the order", "order_uuid", uuid, "error", err)
	}

	return order, nil
}
