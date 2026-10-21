package order

import (
	"context"

	"inventory-go/order/internal/model"
	repoConverter "inventory-go/order/internal/repository/converter"
)

func (r *repository) Update(_ context.Context, order model.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := order.Uuid.String()
	if _, ok := r.data[key]; !ok {
		return model.ErrOrderNotFound
	}

	r.data[key] = repoConverter.OrderToRepoModel(order)

	return nil
}
