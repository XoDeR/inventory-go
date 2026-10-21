package order

import (
	"context"

	"inventory-go/order/internal/model"
	repoConverter "inventory-go/order/internal/repository/converter"
)

func (r *repository) Get(_ context.Context, uuid string) (model.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	repoOrder, ok := r.data[uuid]
	if !ok {
		return model.Order{}, model.ErrOrderNotFound
	}

	return repoConverter.OrderToModel(repoOrder), nil
}
