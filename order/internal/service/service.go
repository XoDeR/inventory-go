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
