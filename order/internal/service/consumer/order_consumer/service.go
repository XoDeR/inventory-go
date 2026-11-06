package order_consumer

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"inventory-go/order/internal/model"
	"inventory-go/order/internal/repository"
	def "inventory-go/order/internal/service"
	"inventory-go/platform/pkg/kafka"
	"inventory-go/platform/pkg/logger"
	"inventory-go/shared/pkg/events"
)

var _ def.OrderConsumerService = (*service)(nil)

type service struct {
	computerAssembledConsumer kafka.Consumer
	orderRepository           repository.OrderRepository
}

func NewService(computerAssembledConsumer kafka.Consumer, orderRepository repository.OrderRepository) *service {
	return &service{
		computerAssembledConsumer: computerAssembledConsumer,
		orderRepository:           orderRepository,
	}
}

// RunConsumer listens for ComputerAssembled events until ctx is cancelled.
func (s *service) RunConsumer(ctx context.Context) error {
	logger.Info(ctx, "Starting the ComputerAssembled consumer")

	return s.computerAssembledConsumer.Consume(ctx, s.ComputerAssembledHandler)
}

// ComputerAssembledHandler moves the order to COMPLETED.
func (s *service) ComputerAssembledHandler(ctx context.Context, msg kafka.Message) error {
	var event events.ComputerAssembled
	if err := events.Decode(msg.Value, &event); err != nil {
		return err
	}

	orderUUID, err := uuid.Parse(event.OrderUUID)
	if err != nil {
		return fmt.Errorf("invalid order_uuid %q: %w", event.OrderUUID, err)
	}

	order, err := s.orderRepository.Get(ctx, orderUUID.String())
	if err != nil {
		return err
	}

	switch order.Status {
	case model.OrderStatusCompleted:
		// a redelivered event: nothing to do
		return nil
	case model.OrderStatusPaid:
	default:
		return fmt.Errorf("order %s is in status %s, expected PAID", orderUUID, order.Status)
	}

	order.Status = model.OrderStatusCompleted

	if err = s.orderRepository.Update(ctx, order); err != nil {
		return err
	}

	logger.Info(ctx, "Order completed", "order_uuid", orderUUID, "event_uuid", event.EventUUID)

	return nil
}
