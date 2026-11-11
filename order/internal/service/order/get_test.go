package order

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"inventory-go/order/internal/model"
)

func testOrder(orderUUID uuid.UUID) model.Order {
	return model.Order{
		Uuid:       orderUUID,
		UserUuid:   testUserUUID,
		PartUuids:  []uuid.UUID{uuid.New(), uuid.New()},
		TotalPrice: 2499.5,
		Status:     model.OrderStatusPendingPayment,
	}
}

func (s *ServiceSuite) TestGetCacheHit() {
	var (
		ctx           = context.Background()
		orderUUID     = uuid.New()
		expectedOrder = testOrder(orderUUID)
	)

	// the repository mock has no expectations: calling it would fail the test
	s.orderCacheRepository.EXPECT().Get(ctx, orderUUID.String()).Return(expectedOrder, nil)

	order, err := s.service.Get(ctx, orderUUID, testUserUUID)

	s.Require().NoError(err)
	s.Require().Equal(expectedOrder, order)
}

func (s *ServiceSuite) TestGetCacheMissFillsCache() {
	var (
		ctx           = context.Background()
		orderUUID     = uuid.New()
		expectedOrder = testOrder(orderUUID)
	)

	s.orderCacheRepository.EXPECT().Get(ctx, orderUUID.String()).Return(model.Order{}, model.ErrOrderNotFound)
	s.orderRepository.EXPECT().Get(ctx, orderUUID.String()).Return(expectedOrder, nil)
	s.orderCacheRepository.EXPECT().Set(ctx, expectedOrder, testCacheTTL).Return(nil)

	order, err := s.service.Get(ctx, orderUUID, testUserUUID)

	s.Require().NoError(err)
	s.Require().Equal(expectedOrder, order)
}

func (s *ServiceSuite) TestGetCacheUnavailableFallsBackToDatabase() {
	var (
		ctx           = context.Background()
		orderUUID     = uuid.New()
		expectedOrder = testOrder(orderUUID)
		cacheErr      = errors.New("redis is down")
	)

	s.orderCacheRepository.EXPECT().Get(ctx, orderUUID.String()).Return(model.Order{}, cacheErr)
	s.orderRepository.EXPECT().Get(ctx, orderUUID.String()).Return(expectedOrder, nil)
	s.orderCacheRepository.EXPECT().Set(ctx, expectedOrder, testCacheTTL).Return(cacheErr)

	order, err := s.service.Get(ctx, orderUUID, testUserUUID)

	s.Require().NoError(err)
	s.Require().Equal(expectedOrder, order)
}

func (s *ServiceSuite) TestGetNotFound() {
	var (
		ctx       = context.Background()
		orderUUID = uuid.New()
	)

	s.orderCacheRepository.EXPECT().Get(ctx, orderUUID.String()).Return(model.Order{}, model.ErrOrderNotFound)
	s.orderRepository.EXPECT().Get(ctx, orderUUID.String()).Return(model.Order{}, model.ErrOrderNotFound)

	order, err := s.service.Get(ctx, orderUUID, testUserUUID)

	s.Require().ErrorIs(err, model.ErrOrderNotFound)
	s.Require().Equal(model.Order{}, order)
}

func (s *ServiceSuite) TestGetRepositoryError() {
	var (
		ctx       = context.Background()
		orderUUID = uuid.New()
		repoErr   = errors.New("database is unavailable")

		// The repository returns a non-empty order together with the error,
		// the service must not leak it to the caller (nor cache it).
		orderWithError = model.Order{Uuid: orderUUID, UserUuid: testUserUUID, TotalPrice: 1}
	)

	s.orderCacheRepository.EXPECT().Get(ctx, orderUUID.String()).Return(model.Order{}, model.ErrOrderNotFound)
	s.orderRepository.EXPECT().Get(ctx, orderUUID.String()).Return(orderWithError, repoErr)

	order, err := s.service.Get(ctx, orderUUID, testUserUUID)

	s.Require().ErrorIs(err, repoErr)
	s.Require().Equal(model.Order{}, order)
}
