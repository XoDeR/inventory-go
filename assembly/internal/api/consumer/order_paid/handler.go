package order_paid

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"inventory-go/assembly/internal/model"
	"inventory-go/assembly/internal/service"
	"inventory-go/platform/pkg/kafka"
	"inventory-go/platform/pkg/logger"
	"inventory-go/shared/pkg/events"
)

// Consumer listens for OrderPaid events and starts the assembly.
type Consumer struct {
	consumer        kafka.Consumer
	assemblyService service.AssemblyService
}

func NewConsumer(consumer kafka.Consumer, assemblyService service.AssemblyService) *Consumer {
	return &Consumer{consumer: consumer, assemblyService: assemblyService}
}

// Run blocks until ctx is cancelled or the consumer fails.
func (c *Consumer) Run(ctx context.Context) error {
	logger.Info(ctx, "Starting the OrderPaid consumer")

	return c.consumer.Consume(ctx, c.Handle)
}

// Handle decodes a message and passes the event to the assembly service.
func (c *Consumer) Handle(ctx context.Context, msg kafka.Message) error {
	var e events.OrderPaid
	if err := events.Decode(msg.Value, &e); err != nil {
		return err
	}

	event, err := toModel(e)
	if err != nil {
		return err
	}

	return c.assemblyService.Assemble(ctx, event)
}

func toModel(e events.OrderPaid) (model.OrderPaidEvent, error) {
	eventUUID, err := uuid.Parse(e.EventUUID)
	if err != nil {
		return model.OrderPaidEvent{}, fmt.Errorf("invalid event_uuid %q: %w", e.EventUUID, err)
	}

	orderUUID, err := uuid.Parse(e.OrderUUID)
	if err != nil {
		return model.OrderPaidEvent{}, fmt.Errorf("invalid order_uuid %q: %w", e.OrderUUID, err)
	}

	userUUID, err := uuid.Parse(e.UserUUID)
	if err != nil {
		return model.OrderPaidEvent{}, fmt.Errorf("invalid user_uuid %q: %w", e.UserUUID, err)
	}

	transactionUUID, err := uuid.Parse(e.TransactionUUID)
	if err != nil {
		return model.OrderPaidEvent{}, fmt.Errorf("invalid transaction_uuid %q: %w", e.TransactionUUID, err)
	}

	return model.OrderPaidEvent{
		EventUUID:       eventUUID,
		OrderUUID:       orderUUID,
		UserUUID:        userUUID,
		PaymentMethod:   e.PaymentMethod,
		TransactionUUID: transactionUUID,
	}, nil
}
