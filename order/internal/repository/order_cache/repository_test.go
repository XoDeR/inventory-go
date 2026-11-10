package order_cache

import (
	"context"
	"time"

	"github.com/google/uuid"

	"inventory-go/order/internal/model"
)

func (s *RepositorySuite) TestSetGetRoundTrip() {
	ctx := context.Background()
	transactionUUID := uuid.New()
	method := model.PaymentMethodCard

	order := model.Order{
		Uuid:            uuid.New(),
		UserUuid:        uuid.New(),
		PartUuids:       []uuid.UUID{uuid.New(), uuid.New()},
		TotalPrice:      2499.5,
		TransactionUuid: &transactionUUID,
		PaymentMethod:   &method,
		Status:          model.OrderStatusPaid,
	}

	s.Require().NoError(s.repository.Set(ctx, order, time.Minute))
	s.T().Cleanup(func() { _ = s.repository.Delete(ctx, order.Uuid.String()) })

	got, err := s.repository.Get(ctx, order.Uuid.String())

	s.Require().NoError(err)
	s.Require().Equal(order, got)
}

func (s *RepositorySuite) TestGetMiss() {
	_, err := s.repository.Get(context.Background(), uuid.NewString())

	s.Require().ErrorIs(err, model.ErrOrderNotFound)
}

func (s *RepositorySuite) TestDelete() {
	ctx := context.Background()
	order := model.Order{Uuid: uuid.New(), Status: model.OrderStatusPendingPayment}

	s.Require().NoError(s.repository.Set(ctx, order, time.Minute))
	s.Require().NoError(s.repository.Delete(ctx, order.Uuid.String()))

	_, err := s.repository.Get(ctx, order.Uuid.String())
	s.Require().ErrorIs(err, model.ErrOrderNotFound)
}

func (s *RepositorySuite) TestExpiresAfterTTL() {
	ctx := context.Background()
	order := model.Order{Uuid: uuid.New(), Status: model.OrderStatusPendingPayment}

	s.Require().NoError(s.repository.Set(ctx, order, 200*time.Millisecond))

	s.Require().Eventually(func() bool {
		_, err := s.repository.Get(ctx, order.Uuid.String())
		return err == model.ErrOrderNotFound
	}, 3*time.Second, 100*time.Millisecond)
}
