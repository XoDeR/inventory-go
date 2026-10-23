package order

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"inventory-go/order/internal/model"
)

func (s *ServiceSuite) TestGetSuccess() {
	var (
		ctx       = context.Background()
		orderUUID = uuid.New()

		expectedOrder = model.Order{
			Uuid:       orderUUID,
			UserUuid:   uuid.New(),
			PartUuids:  []uuid.UUID{uuid.New(), uuid.New()},
			TotalPrice: 2499.5,
			Status:     model.OrderStatusPendingPayment,
		}
	)

	s.orderRepository.EXPECT().Get(ctx, orderUUID.String()).Return(expectedOrder, nil)

	order, err := s.service.Get(ctx, orderUUID)

	s.Require().NoError(err)
	s.Require().Equal(expectedOrder, order)
}

func (s *ServiceSuite) TestGetNotFound() {
	var (
		ctx       = context.Background()
		orderUUID = uuid.New()
	)

	s.orderRepository.EXPECT().Get(ctx, orderUUID.String()).Return(model.Order{}, model.ErrOrderNotFound)

	order, err := s.service.Get(ctx, orderUUID)

	s.Require().ErrorIs(err, model.ErrOrderNotFound)
	s.Require().Equal(model.Order{}, order)
}

func (s *ServiceSuite) TestGetRepositoryError() {
	var (
		ctx       = context.Background()
		orderUUID = uuid.New()
		repoErr   = errors.New("database is unavailable")

		// The repository returns a non-empty order together with the error,
		// the service must not leak it to the caller.
		orderWithError = model.Order{Uuid: orderUUID, TotalPrice: 1}
	)

	s.orderRepository.EXPECT().Get(ctx, orderUUID.String()).Return(orderWithError, repoErr)

	order, err := s.service.Get(ctx, orderUUID)

	s.Require().ErrorIs(err, repoErr)
	s.Require().Equal(model.Order{}, order)
}
