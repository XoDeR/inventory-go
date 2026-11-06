package service

import (
	"context"

	"github.com/google/uuid"

	"inventory-go/order/internal/model"
)

type OrderService interface {
	Create(ctx context.Context, info model.CreateOrderInfo) (model.Order, error)
	Get(ctx context.Context, uuid uuid.UUID) (model.Order, error)
	// Pay pays the order and returns the transaction UUID.
	Pay(ctx context.Context, uuid uuid.UUID, method model.PaymentMethod) (uuid.UUID, error)
	Cancel(ctx context.Context, uuid uuid.UUID) error
}

// OrderProducerService publishes order events to Kafka.
type OrderProducerService interface {
	ProduceOrderPaid(ctx context.Context, event model.OrderPaidEvent) error
}

// OrderConsumerService consumes events addressed to the order service from Kafka.
type OrderConsumerService interface {
	// RunConsumer blocks until ctx is cancelled or the consumer fails.
	RunConsumer(ctx context.Context) error
}
