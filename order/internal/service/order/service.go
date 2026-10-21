package order

import (
	"inventory-go/order/internal/repository"
	def "inventory-go/order/internal/service"
)

var _ def.OrderService = (*service)(nil)

type service struct {
	orderRepository repository.OrderRepository
}

func NewService(orderRepository repository.OrderRepository) *service {
	return &service{
		orderRepository: orderRepository,
	}
}
