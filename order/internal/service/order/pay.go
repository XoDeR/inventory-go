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

	transactionID, err := s.paymentClient.PayOrder(ctx, model.PayOrderInfo{
		OrderUuid:     order.Uuid,
		UserUuid:      order.UserUuid,
		PaymentMethod: method,
	})
	if err != nil {
		return uuid.Nil, err
	}

	transactionUUID, err := uuid.Parse(transactionID)
	if err != nil {
		return uuid.Nil, err
	}

	order.Status = model.OrderStatusPaid
	order.TransactionUuid = &transactionUUID
	order.PaymentMethod = &method

	if err := s.orderRepository.Update(ctx, order); err != nil {
		return uuid.Nil, err
	}

	return transactionUUID, nil
}
