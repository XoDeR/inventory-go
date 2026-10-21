package order

import (
	"context"

	"github.com/google/uuid"

	"inventory-go/order/internal/model"
)

func (s *service) Cancel(ctx context.Context, uuid uuid.UUID) error {
	order, err := s.orderRepository.Get(ctx, uuid.String())
	if err != nil {
		return err
	}

	if order.Status == model.OrderStatusPaid {
		return model.ErrOrderPaid
	}

	order.Status = model.OrderStatusCancelled

	return s.orderRepository.Update(ctx, order)
}
