package service

import (
	"context"

	"inventory-go/payment/internal/model"
)

type PaymentService interface {
	// PayOrder processes a payment and returns the transaction UUID.
	PayOrder(ctx context.Context, info model.PayOrderInfo) (string, error)
}
