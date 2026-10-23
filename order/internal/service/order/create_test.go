package order

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"inventory-go/order/internal/model"
)

func (s *ServiceSuite) TestCreateSuccess() {
	var (
		ctx = context.Background()

		userUUID = uuid.New()
		partA    = uuid.New()
		partB    = uuid.New()

		info = model.CreateOrderInfo{UserUuid: userUUID, PartUuids: []uuid.UUID{partA, partB}}

		filter = model.PartsFilter{Uuids: []string{partA.String(), partB.String()}}
		parts  = []model.Part{
			{Uuid: partA.String(), Price: 1999.99},
			{Uuid: partB.String(), Price: 119.5},
		}
	)

	s.inventoryClient.EXPECT().ListParts(ctx, filter).Return(parts, nil)

	// The order UUID is generated inside the service, so it is only checked to be set
	// and the rest of the order is matched field by field.
	s.orderRepository.EXPECT().Create(ctx, mock.MatchedBy(func(order model.Order) bool {
		return order.Uuid != uuid.Nil &&
			order.UserUuid == userUUID &&
			len(order.PartUuids) == 2 &&
			order.Status == model.OrderStatusPendingPayment &&
			order.TransactionUuid == nil &&
			order.PaymentMethod == nil
	})).Return(nil)

	order, err := s.service.Create(ctx, info)

	s.Require().NoError(err)
	s.Require().NotEqual(uuid.Nil, order.Uuid)
	s.Require().Equal(userUUID, order.UserUuid)
	s.Require().Equal(info.PartUuids, order.PartUuids)
	s.Require().InDelta(2119.49, order.TotalPrice, 0.0001)
	s.Require().Equal(model.OrderStatusPendingPayment, order.Status)
}

func (s *ServiceSuite) TestCreateSamePartTwiceIsCountedTwice() {
	var (
		ctx = context.Background()

		part = uuid.New()
		info = model.CreateOrderInfo{UserUuid: uuid.New(), PartUuids: []uuid.UUID{part, part}}

		filter = model.PartsFilter{Uuids: []string{part.String(), part.String()}}
		parts  = []model.Part{{Uuid: part.String(), Price: 10}}
	)

	s.inventoryClient.EXPECT().ListParts(ctx, filter).Return(parts, nil)
	s.orderRepository.EXPECT().Create(ctx, mock.Anything).Return(nil)

	order, err := s.service.Create(ctx, info)

	s.Require().NoError(err)
	s.Require().InDelta(20.0, order.TotalPrice, 0.0001)
}

func (s *ServiceSuite) TestCreateEmptyPartUuids() {
	ctx := context.Background()

	// Neither the inventory nor the repository is called: no expectations are set.
	order, err := s.service.Create(ctx, model.CreateOrderInfo{UserUuid: uuid.New()})

	s.Require().ErrorIs(err, model.ErrEmptyPartUuids)
	s.Require().Equal(model.Order{}, order)
}

func (s *ServiceSuite) TestCreatePartsNotFound() {
	var (
		ctx = context.Background()

		partA = uuid.New()
		partB = uuid.New()
		info  = model.CreateOrderInfo{UserUuid: uuid.New(), PartUuids: []uuid.UUID{partA, partB}}

		filter = model.PartsFilter{Uuids: []string{partA.String(), partB.String()}}
		// The inventory knows only one of the two requested parts.
		parts = []model.Part{{Uuid: partA.String(), Price: 10}}
	)

	s.inventoryClient.EXPECT().ListParts(ctx, filter).Return(parts, nil)

	order, err := s.service.Create(ctx, info)

	s.Require().ErrorIs(err, model.ErrPartsNotFound)
	s.Require().Equal(model.Order{}, order)
}

func (s *ServiceSuite) TestCreateInventoryError() {
	var (
		ctx          = context.Background()
		part         = uuid.New()
		info         = model.CreateOrderInfo{UserUuid: uuid.New(), PartUuids: []uuid.UUID{part}}
		filter       = model.PartsFilter{Uuids: []string{part.String()}}
		inventoryErr = errors.New("inventory is unavailable")
	)

	s.inventoryClient.EXPECT().ListParts(ctx, filter).Return(nil, inventoryErr)

	order, err := s.service.Create(ctx, info)

	s.Require().ErrorIs(err, inventoryErr)
	s.Require().Equal(model.Order{}, order)
}

func (s *ServiceSuite) TestCreateRepositoryError() {
	var (
		ctx     = context.Background()
		part    = uuid.New()
		info    = model.CreateOrderInfo{UserUuid: uuid.New(), PartUuids: []uuid.UUID{part}}
		filter  = model.PartsFilter{Uuids: []string{part.String()}}
		parts   = []model.Part{{Uuid: part.String(), Price: 10}}
		repoErr = errors.New("database is unavailable")
	)

	s.inventoryClient.EXPECT().ListParts(ctx, filter).Return(parts, nil)
	s.orderRepository.EXPECT().Create(ctx, mock.Anything).Return(repoErr)

	order, err := s.service.Create(ctx, info)

	s.Require().ErrorIs(err, repoErr)
	s.Require().Equal(model.Order{}, order)
}
