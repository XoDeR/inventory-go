package order

import (
	"context"

	"github.com/google/uuid"

	"inventory-go/order/internal/model"
	"inventory-go/platform/pkg/logger"
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

	s.invalidateCache(ctx, orderUUID.String())

	// The payment is already done and saved, so a Kafka failure must not fail the request:
	// it is logged and the order stays PAID (there is no outbox yet to retry the publication).
	err = s.orderProducer.ProduceOrderPaid(ctx, model.OrderPaidEvent{
		EventUUID:       uuid.New(),
		OrderUUID:       order.Uuid,
		UserUUID:        order.UserUuid,
		PaymentMethod:   method,
		TransactionUUID: transactionUUID,
	})
	if err != nil {
		logger.Error(ctx, "failed to publish the OrderPaid event", "order_uuid", order.Uuid, "error", err)
	}

	return transactionUUID, nil
}
