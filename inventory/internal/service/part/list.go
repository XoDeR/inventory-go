package part

import (
	"context"

	"inventory-go/inventory/internal/model"
)

func (s *service) List(ctx context.Context, filter model.PartsFilter) ([]model.Part, error) {
	return s.inventoryRepository.List(ctx, filter)
}
