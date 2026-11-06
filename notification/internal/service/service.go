package service

import (
	"context"

	"inventory-go/notification/internal/model"
)

// NotificationService turns domain events into notifications.
type NotificationService interface {
	NotifyOrderPaid(ctx context.Context, event model.OrderPaidEvent) error
	NotifyComputerAssembled(ctx context.Context, event model.ComputerAssembledEvent) error
}
