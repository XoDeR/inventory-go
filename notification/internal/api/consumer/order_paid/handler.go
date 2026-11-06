package order_paid

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"inventory-go/notification/internal/model"
	"inventory-go/notification/internal/service"
	"inventory-go/platform/pkg/kafka"
	"inventory-go/platform/pkg/logger"
	"inventory-go/shared/pkg/events"
)

// Consumer listens for OrderPaid events and sends notifications.
type Consumer struct {
	consumer            kafka.Consumer
	notificationService service.NotificationService
}

func NewConsumer(consumer kafka.Consumer, notificationService service.NotificationService) *Consumer {
	return &Consumer{consumer: consumer, notificationService: notificationService}
}

// Run blocks until ctx is cancelled or the consumer fails.
func (c *Consumer) Run(ctx context.Context) error {
	logger.Info(ctx, "Starting the OrderPaid consumer")

	return c.consumer.Consume(ctx, c.Handle)
}

// Handle decodes a message and passes the event to the notification service.
func (c *Consumer) Handle(ctx context.Context, msg kafka.Message) error {
	var e events.OrderPaid
	if err := events.Decode(msg.Value, &e); err != nil {
		return err
	}

	eventUUID, err := uuid.Parse(e.EventUUID)
	if err != nil {
		return fmt.Errorf("invalid event_uuid %q: %w", e.EventUUID, err)
	}

	orderUUID, err := uuid.Parse(e.OrderUUID)
	if err != nil {
		return fmt.Errorf("invalid order_uuid %q: %w", e.OrderUUID, err)
	}

	userUUID, err := uuid.Parse(e.UserUUID)
	if err != nil {
		return fmt.Errorf("invalid user_uuid %q: %w", e.UserUUID, err)
	}

	transactionUUID, err := uuid.Parse(e.TransactionUUID)
	if err != nil {
		return fmt.Errorf("invalid transaction_uuid %q: %w", e.TransactionUUID, err)
	}

	return c.notificationService.NotifyOrderPaid(ctx, model.OrderPaidEvent{
		EventUUID:       eventUUID,
		OrderUUID:       orderUUID,
		UserUUID:        userUUID,
		PaymentMethod:   e.PaymentMethod,
		TransactionUUID: transactionUUID,
	})
}
