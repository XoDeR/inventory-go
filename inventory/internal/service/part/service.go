package part

import (
	"time"

	"inventory-go/inventory/internal/repository"
	def "inventory-go/inventory/internal/service"
)

var _ def.InventoryService = (*service)(nil)

type service struct {
	inventoryRepository      repository.InventoryRepository
	inventoryCacheRepository repository.InventoryCacheRepository
	cacheTTL                 time.Duration
}

func NewService(
	inventoryRepository repository.InventoryRepository,
	inventoryCacheRepository repository.InventoryCacheRepository,
	cacheTTL time.Duration,
) *service {
	return &service{
		inventoryRepository:      inventoryRepository,
		inventoryCacheRepository: inventoryCacheRepository,
		cacheTTL:                 cacheTTL,
	}
}
