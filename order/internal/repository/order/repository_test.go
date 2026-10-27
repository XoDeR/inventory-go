package order

import (
	"context"

	"github.com/google/uuid"

	"inventory-go/order/internal/migrator"
	"inventory-go/order/internal/model"
	"inventory-go/order/migrations"
)

func newOrder() model.Order {
	return model.Order{
		Uuid:       uuid.New(),
		UserUuid:   uuid.New(),
		PartUuids:  []uuid.UUID{uuid.New(), uuid.New()},
		TotalPrice: 123.45,
		Status:     model.OrderStatusPendingPayment,
	}
}

func (s *RepositorySuite) TestCreateAndGet() {
	ctx := context.Background()
	order := newOrder()

	s.Require().NoError(s.repository.Create(ctx, order))

	got, err := s.repository.Get(ctx, order.Uuid.String())

	s.Require().NoError(err)
	s.Require().Equal(order, got)
	s.Require().Nil(got.TransactionUuid)
	s.Require().Nil(got.PaymentMethod)
}

func (s *RepositorySuite) TestCreateDuplicate() {
	ctx := context.Background()
	order := newOrder()

	s.Require().NoError(s.repository.Create(ctx, order))

	s.Require().Error(s.repository.Create(ctx, order))
}

func (s *RepositorySuite) TestGetNotFound() {
	got, err := s.repository.Get(context.Background(), uuid.NewString())

	s.Require().ErrorIs(err, model.ErrOrderNotFound)
	s.Require().Equal(model.Order{}, got)
}

func (s *RepositorySuite) TestUpdate() {
	ctx := context.Background()
	order := newOrder()
	s.Require().NoError(s.repository.Create(ctx, order))

	transactionUuid := uuid.New()
	method := model.PaymentMethodCard
	order.Status = model.OrderStatusPaid
	order.TransactionUuid = &transactionUuid
	order.PaymentMethod = &method

	s.Require().NoError(s.repository.Update(ctx, order))

	got, err := s.repository.Get(ctx, order.Uuid.String())

	s.Require().NoError(err)
	s.Require().Equal(order, got)
}

func (s *RepositorySuite) TestUpdateStatusToCancelled() {
	ctx := context.Background()
	order := newOrder()
	s.Require().NoError(s.repository.Create(ctx, order))

	order.Status = model.OrderStatusCancelled

	s.Require().NoError(s.repository.Update(ctx, order))

	got, err := s.repository.Get(ctx, order.Uuid.String())

	s.Require().NoError(err)
	s.Require().Equal(model.OrderStatusCancelled, got.Status)
}

func (s *RepositorySuite) TestUpdateNotFound() {
	err := s.repository.Update(context.Background(), newOrder())

	s.Require().ErrorIs(err, model.ErrOrderNotFound)
}

func (s *RepositorySuite) TestMigrationsAreIdempotent() {
	ctx := context.Background()
	order := newOrder()
	s.Require().NoError(s.repository.Create(ctx, order))

	// SetupTest already applied the migrations, applying them again must be a no-op that keeps the data.
	db := stdlibDB(s)
	s.Require().NoError(migrator.NewMigrator(db).Up(ctx, migrations.FS))
	s.Require().NoError(db.Close())

	_, err := s.repository.Get(ctx, order.Uuid.String())
	s.Require().NoError(err)
}
