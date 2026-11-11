package order

import (
	"context"

	"github.com/google/uuid"

	"inventory-go/order/internal/model"
)

// An order of another user looks exactly like a missing one, and nothing else is touched:
// the mocks have no further expectations, so a payment, update or cache write would fail the test.

func (s *ServiceSuite) foreignOrder(orderUUID uuid.UUID) model.Order {
	return model.Order{
		Uuid:       orderUUID,
		UserUuid:   uuid.New(), // not testUserUUID
		TotalPrice: 100,
		Status:     model.OrderStatusPendingPayment,
	}
}

func (s *ServiceSuite) TestGetForeignOrderFromDatabase() {
	ctx := context.Background()
	orderUUID := uuid.New()
	order := s.foreignOrder(orderUUID)

	s.orderCacheRepository.EXPECT().Get(ctx, orderUUID.String()).Return(model.Order{}, model.ErrOrderNotFound)
	s.orderRepository.EXPECT().Get(ctx, orderUUID.String()).Return(order, nil)
	s.orderCacheRepository.EXPECT().Set(ctx, order, testCacheTTL).Return(nil)

	got, err := s.service.Get(ctx, orderUUID, testUserUUID)

	s.Require().ErrorIs(err, model.ErrOrderNotFound)
	s.Require().Equal(model.Order{}, got)
}

func (s *ServiceSuite) TestGetForeignOrderFromCache() {
	ctx := context.Background()
	orderUUID := uuid.New()

	s.orderCacheRepository.EXPECT().Get(ctx, orderUUID.String()).Return(s.foreignOrder(orderUUID), nil)

	got, err := s.service.Get(ctx, orderUUID, testUserUUID)

	s.Require().ErrorIs(err, model.ErrOrderNotFound)
	s.Require().Equal(model.Order{}, got)
}

func (s *ServiceSuite) TestPayForeignOrder() {
	ctx := context.Background()
	orderUUID := uuid.New()

	s.orderRepository.EXPECT().Get(ctx, orderUUID.String()).Return(s.foreignOrder(orderUUID), nil)

	transactionUUID, err := s.service.Pay(ctx, orderUUID, testUserUUID, model.PaymentMethodCard)

	s.Require().ErrorIs(err, model.ErrOrderNotFound)
	s.Require().Equal(uuid.Nil, transactionUUID)
}

func (s *ServiceSuite) TestCancelForeignOrder() {
	ctx := context.Background()
	orderUUID := uuid.New()

	s.orderRepository.EXPECT().Get(ctx, orderUUID.String()).Return(s.foreignOrder(orderUUID), nil)

	err := s.service.Cancel(ctx, orderUUID, testUserUUID)

	s.Require().ErrorIs(err, model.ErrOrderNotFound)
}
