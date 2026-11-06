package order

import (
	grpcClient "inventory-go/order/internal/client/grpc"
	"inventory-go/order/internal/repository"
	def "inventory-go/order/internal/service"
)

var _ def.OrderService = (*service)(nil)

type service struct {
	orderRepository repository.OrderRepository

	inventoryClient grpcClient.InventoryClient
	paymentClient   grpcClient.PaymentClient

	orderProducer def.OrderProducerService
}

func NewService(
	orderRepository repository.OrderRepository,
	inventoryClient grpcClient.InventoryClient,
	paymentClient grpcClient.PaymentClient,
	orderProducer def.OrderProducerService,
) *service {
	return &service{
		orderRepository: orderRepository,
		inventoryClient: inventoryClient,
		paymentClient:   paymentClient,
		orderProducer:   orderProducer,
	}
}
