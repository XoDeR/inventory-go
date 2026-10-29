package payment

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"inventory-go/payment/internal/model"
)

// PayOrder simulates a payment: it validates the request, logs it and
// returns a new transaction UUID. No state is stored.
func (s *service) PayOrder(_ context.Context, info model.PayOrderInfo) (string, error) {
	if info.OrderUuid == "" {
		return "", fmt.Errorf("%w: order_uuid is required", model.ErrInvalidArgument)
	}
	if info.UserUuid == "" {
		return "", fmt.Errorf("%w: user_uuid is required", model.ErrInvalidArgument)
	}
	if info.PaymentMethod == model.PaymentMethodUnspecified {
		return "", fmt.Errorf("%w: payment_method is required", model.ErrInvalidArgument)
	}

	transactionUUID := uuid.NewString()
	slog.Info("payment successful",
		"order_uuid", info.OrderUuid,
		"user_uuid", info.UserUuid,
		"method", info.PaymentMethod,
		"transaction_uuid", transactionUUID,
	)

	return transactionUUID, nil
}
