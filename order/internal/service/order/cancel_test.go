package order

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"inventory-go/order/internal/model"
)

func (s *ServiceSuite) TestCancelSuccess() {
	var (
		ctx       = context.Background()
		orderUUID = uuid.New()

		storedOrder = model.Order{
			Uuid:       orderUUID,
			UserUuid:   uuid.New(),
			TotalPrice: 100,
			Status:     model.OrderStatusPendingPayment,
		}

		cancelledOrder = storedOrder
	)
	cancelledOrder.Status = model.OrderStatusCancelled

	s.orderRepository.EXPECT().Get(ctx, orderUUID.String()).Return(storedOrder, nil)
	s.orderRepository.EXPECT().Update(ctx, cancelledOrder).Return(nil)

	err := s.service.Cancel(ctx, orderUUID)

	s.Require().NoError(err)
}

func (s *ServiceSuite) TestCancelAlreadyPaid() {
	var (
		ctx       = context.Background()
		orderUUID = uuid.New()

		paidOrder = model.Order{Uuid: orderUUID, Status: model.OrderStatusPaid}
	)

	// A paid order must not be updated: no Update expectation is set, so any call fails the test.
	s.orderRepository.EXPECT().Get(ctx, orderUUID.String()).Return(paidOrder, nil)

	err := s.service.Cancel(ctx, orderUUID)

	s.Require().ErrorIs(err, model.ErrOrderPaid)
}

func (s *ServiceSuite) TestCancelNotFound() {
	var (
		ctx       = context.Background()
		orderUUID = uuid.New()
	)

	s.orderRepository.EXPECT().Get(ctx, orderUUID.String()).Return(model.Order{}, model.ErrOrderNotFound)

	err := s.service.Cancel(ctx, orderUUID)

	s.Require().ErrorIs(err, model.ErrOrderNotFound)
}

func (s *ServiceSuite) TestCancelUpdateError() {
	var (
		ctx       = context.Background()
		orderUUID = uuid.New()
		repoErr   = errors.New("database is unavailable")

		storedOrder    = model.Order{Uuid: orderUUID, Status: model.OrderStatusPendingPayment}
		cancelledOrder = model.Order{Uuid: orderUUID, Status: model.OrderStatusCancelled}
	)

	s.orderRepository.EXPECT().Get(ctx, orderUUID.String()).Return(storedOrder, nil)
	s.orderRepository.EXPECT().Update(ctx, cancelledOrder).Return(repoErr)

	err := s.service.Cancel(ctx, orderUUID)

	s.Require().ErrorIs(err, repoErr)
}
