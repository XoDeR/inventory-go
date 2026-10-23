package v1

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"

	"inventory-go/order/internal/converter"
	"inventory-go/order/internal/model"
	orderV1 "inventory-go/shared/pkg/openapi/order/v1"
)

func (s *APISuite) TestGetOrderByUuidSuccess() {
	var (
		ctx           = context.Background()
		orderUUID     = uuid.New()
		transactionID = uuid.New()
		method        = model.PaymentMethodCard

		order = model.Order{
			Uuid:            orderUUID,
			UserUuid:        uuid.New(),
			PartUuids:       []uuid.UUID{uuid.New()},
			TotalPrice:      100,
			TransactionUuid: &transactionID,
			PaymentMethod:   &method,
			Status:          model.OrderStatusPaid,
		}
	)

	s.orderService.EXPECT().Get(ctx, orderUUID).Return(order, nil)

	res, err := s.api.GetOrderByUuid(ctx, orderV1.GetOrderByUuidParams{OrderUUID: orderUUID})

	s.Require().NoError(err)
	s.Require().Equal(converter.OrderToOpenAPI(order), res)

	got, ok := res.(*orderV1.Order)
	s.Require().True(ok)
	s.Require().Equal(orderV1.OrderStatus(model.OrderStatusPaid), got.Status)
	s.Require().Equal(transactionID, got.TransactionUUID.Value)
	s.Require().Equal(orderV1.PaymentMethodCARD, got.PaymentMethod.Value)
}

func (s *APISuite) TestGetOrderByUuidPendingHasNoPaymentInfo() {
	var (
		ctx       = context.Background()
		orderUUID = uuid.New()

		order = model.Order{Uuid: orderUUID, Status: model.OrderStatusPendingPayment}
	)

	s.orderService.EXPECT().Get(ctx, orderUUID).Return(order, nil)

	res, err := s.api.GetOrderByUuid(ctx, orderV1.GetOrderByUuidParams{OrderUUID: orderUUID})

	s.Require().NoError(err)

	got, ok := res.(*orderV1.Order)
	s.Require().True(ok)
	s.Require().False(got.TransactionUUID.IsSet())
	s.Require().False(got.PaymentMethod.IsSet())
}

func (s *APISuite) TestGetOrderByUuidNotFound() {
	var (
		ctx       = context.Background()
		orderUUID = uuid.New()
	)

	s.orderService.EXPECT().Get(ctx, orderUUID).Return(model.Order{}, model.ErrOrderNotFound)

	res, err := s.api.GetOrderByUuid(ctx, orderV1.GetOrderByUuidParams{OrderUUID: orderUUID})

	s.Require().NoError(err)
	s.Require().Equal(&orderV1.GetOrderByUuidNotFound{
		Code:    http.StatusNotFound,
		Message: model.ErrOrderNotFound.Error(),
	}, res)
}

func (s *APISuite) TestGetOrderByUuidServiceError() {
	var (
		ctx        = context.Background()
		orderUUID  = uuid.New()
		serviceErr = errors.New("database is unavailable")
	)

	s.orderService.EXPECT().Get(ctx, orderUUID).Return(model.Order{}, serviceErr)

	res, err := s.api.GetOrderByUuid(ctx, orderV1.GetOrderByUuidParams{OrderUUID: orderUUID})

	s.Require().Nil(res)
	s.Require().ErrorIs(err, serviceErr)
}
