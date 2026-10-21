package order

import (
	"context"

	"github.com/google/uuid"

	"inventory-go/order/internal/model"
)

func (s *service) Pay(ctx context.Context, orderUUID uuid.UUID, method model.PaymentMethod) (uuid.UUID, error) {
	order, err := s.orderRepository.Get(ctx, orderUUID.String())
	if err != nil {
		return uuid.Nil, err
	}

	// TODO: call PaymentService.PayOrder with user_uuid, order_uuid and payment_method.
	// Until the payment client exists, the transaction uuid is generated locally.
	transactionUUID := uuid.New()

	order.Status = model.OrderStatusPaid
	order.TransactionUuid = &transactionUUID
	order.PaymentMethod = &method

	if err := s.orderRepository.Update(ctx, order); err != nil {
		return uuid.Nil, err
	}

	return transactionUUID, nil
}
