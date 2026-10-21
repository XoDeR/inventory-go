package order

import (
	"context"

	"inventory-go/order/internal/model"
	repoConverter "inventory-go/order/internal/repository/converter"
)

func (r *repository) Create(_ context.Context, order model.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.data[order.Uuid.String()] = repoConverter.OrderToRepoModel(order)

	return nil
}
