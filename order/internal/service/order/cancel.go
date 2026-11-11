package order

import (
	"context"

	"github.com/google/uuid"

	"inventory-go/order/internal/model"
)

func (s *service) Cancel(ctx context.Context, orderUUID, userUUID uuid.UUID) error {
	order, err := s.orderRepository.Get(ctx, orderUUID.String())
	if err != nil {
		return err
	}

	if order, err = ownedBy(order, userUUID); err != nil {
		return err
	}

	if order.Status == model.OrderStatusPaid || order.Status == model.OrderStatusCompleted {
		return model.ErrOrderPaid
	}

	order.Status = model.OrderStatusCancelled

	if err = s.orderRepository.Update(ctx, order); err != nil {
		return err
	}

	s.invalidateCache(ctx, orderUUID.String())

	return nil
}
