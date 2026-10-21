package part

import (
	"context"

	"inventory-go/inventory/internal/model"
)

func (s *service) Get(ctx context.Context, uuid string) (model.Part, error) {
	part, err := s.inventoryRepository.Get(ctx, uuid)
	if err != nil {
		return model.Part{}, err
	}

	return part, nil
}
