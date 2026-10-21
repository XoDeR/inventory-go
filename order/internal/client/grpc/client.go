package grpc

import (
	"context"

	"inventory-go/order/internal/model"
)

type InventoryClient interface {
	ListParts(ctx context.Context, filter model.PartsFilter) ([]model.Part, error)
}

type PaymentClient interface {
	// PayOrder processes a payment and returns the transaction UUID.
	PayOrder(ctx context.Context, info model.PayOrderInfo) (string, error)
}
