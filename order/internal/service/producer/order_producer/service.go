package order_producer

import (
	"context"

	"inventory-go/order/internal/model"
	def "inventory-go/order/internal/service"
	"inventory-go/platform/pkg/kafka"
	"inventory-go/shared/pkg/events"
)

var _ def.OrderProducerService = (*service)(nil)

type service struct {
	orderPaidProducer kafka.Producer
}

func NewService(orderPaidProducer kafka.Producer) *service {
	return &service{orderPaidProducer: orderPaidProducer}
}

// ProduceOrderPaid sends the OrderPaid event, keyed by the order UUID.
func (s *service) ProduceOrderPaid(ctx context.Context, event model.OrderPaidEvent) error {
	payload, err := events.Encode(events.OrderPaid{
		EventUUID:       event.EventUUID.String(),
		OrderUUID:       event.OrderUUID.String(),
		UserUUID:        event.UserUUID.String(),
		PaymentMethod:   string(event.PaymentMethod),
		TransactionUUID: event.TransactionUUID.String(),
	})
	if err != nil {
		return err
	}

	return s.orderPaidProducer.Send(ctx, []byte(event.OrderUUID.String()), payload)
}
