package service

import (
	"context"

	"github.com/google/uuid"

	"inventory-go/order/internal/model"
)

type OrderService interface {
	Create(ctx context.Context, info model.CreateOrderInfo) (model.Order, error)
	// Get, Pay and Cancel work only on the orders of userUUID: for any other user the order
	// is reported as model.ErrOrderNotFound, so the existence of foreign orders is not leaked.
	Get(ctx context.Context, orderUUID, userUUID uuid.UUID) (model.Order, error)
	// Pay pays the order and returns the transaction UUID.
	Pay(ctx context.Context, orderUUID, userUUID uuid.UUID, method model.PaymentMethod) (uuid.UUID, error)
	Cancel(ctx context.Context, orderUUID, userUUID uuid.UUID) error
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
