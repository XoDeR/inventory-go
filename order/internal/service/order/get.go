package order

import (
	"context"

	"github.com/google/uuid"

	"inventory-go/order/internal/model"
)

func (s *service) Get(ctx context.Context, uuid uuid.UUID) (model.Order, error) {
	order, err := s.orderRepository.Get(ctx, uuid.String())
	if err != nil {
		return model.Order{}, err
	}

	return order, nil
}
