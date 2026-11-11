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
// The order is returned only to its owner.
func (s *service) Get(ctx context.Context, orderUUID, userUUID uuid.UUID) (model.Order, error) {
	cached, err := s.orderCacheRepository.Get(ctx, orderUUID.String())
	if err == nil {
		return ownedBy(cached, userUUID)
	}

	if !errors.Is(err, model.ErrOrderNotFound) {
		logger.Error(ctx, "failed to read the order from the cache", "order_uuid", orderUUID, "error", err)
	}

	order, err := s.orderRepository.Get(ctx, orderUUID.String())
	if err != nil {
		return model.Order{}, err
	}

	if err = s.orderCacheRepository.Set(ctx, order, s.cacheTTL); err != nil {
		logger.Error(ctx, "failed to cache the order", "order_uuid", orderUUID, "error", err)
	}

	return ownedBy(order, userUUID)
}

// ownedBy returns the order if it belongs to the user, and model.ErrOrderNotFound otherwise.
func ownedBy(order model.Order, userUUID uuid.UUID) (model.Order, error) {
	if order.UserUuid != userUUID {
		return model.Order{}, model.ErrOrderNotFound
	}

	return order, nil
}
