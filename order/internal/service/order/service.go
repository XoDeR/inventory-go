package order

import (
	"context"
	"time"

	grpcClient "inventory-go/order/internal/client/grpc"
	"inventory-go/order/internal/repository"
	def "inventory-go/order/internal/service"
	"inventory-go/platform/pkg/logger"
)

var _ def.OrderService = (*service)(nil)

type service struct {
	orderRepository      repository.OrderRepository
	orderCacheRepository repository.OrderCacheRepository
	cacheTTL             time.Duration

	inventoryClient grpcClient.InventoryClient
	paymentClient   grpcClient.PaymentClient

	orderProducer def.OrderProducerService
}

func NewService(
	orderRepository repository.OrderRepository,
	orderCacheRepository repository.OrderCacheRepository,
	cacheTTL time.Duration,
	inventoryClient grpcClient.InventoryClient,
	paymentClient grpcClient.PaymentClient,
	orderProducer def.OrderProducerService,
) *service {
	return &service{
		orderRepository:      orderRepository,
		orderCacheRepository: orderCacheRepository,
		cacheTTL:             cacheTTL,
		inventoryClient:      inventoryClient,
		paymentClient:        paymentClient,
		orderProducer:        orderProducer,
	}
}

// invalidateCache drops the cached order after it was changed in the database.
// A cache failure is only logged: the entry expires by its TTL anyway.
func (s *service) invalidateCache(ctx context.Context, orderUUID string) {
	if err := s.orderCacheRepository.Delete(ctx, orderUUID); err != nil {
		logger.Error(ctx, "failed to invalidate the cached order", "order_uuid", orderUUID, "error", err)
	}
}
