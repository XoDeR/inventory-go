package service

import (
	"context"

	"inventory-go/inventory/internal/model"
)

type InventoryService interface {
	Get(ctx context.Context, uuid string) (model.Part, error)
	List(ctx context.Context, filter model.PartsFilter) ([]model.Part, error)
}
