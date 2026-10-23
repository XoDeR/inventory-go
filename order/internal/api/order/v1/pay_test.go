package v1

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"

	"inventory-go/order/internal/model"
	orderV1 "inventory-go/shared/pkg/openapi/order/v1"
)

func (s *APISuite) TestPayOrderSuccess() {
	var (
		ctx           = context.Background()
		orderUUID     = uuid.New()
		transactionID = uuid.New()

		req    = &orderV1.PayOrderRequest{PaymentMethod: orderV1.PaymentMethodCREDITCARD}
		params = orderV1.PayOrderParams{OrderUUID: orderUUID}
	)

	// The OpenAPI payment method must be converted to the model one before it reaches the service.
	s.orderService.EXPECT().Pay(ctx, orderUUID, model.PaymentMethodCreditCard).Return(transactionID, nil)

	res, err := s.api.PayOrder(ctx, req, params)

	s.Require().NoError(err)
	s.Require().Equal(&orderV1.PayOrderResponse{TransactionUUID: transactionID}, res)
}

func (s *APISuite) TestPayOrderNotFound() {
	var (
		ctx       = context.Background()
		orderUUID = uuid.New()

		req    = &orderV1.PayOrderRequest{PaymentMethod: orderV1.PaymentMethodCARD}
		params = orderV1.PayOrderParams{OrderUUID: orderUUID}
	)

	s.orderService.EXPECT().Pay(ctx, orderUUID, model.PaymentMethodCard).Return(uuid.Nil, model.ErrOrderNotFound)

	res, err := s.api.PayOrder(ctx, req, params)

	s.Require().NoError(err)
	s.Require().Equal(&orderV1.PayOrderNotFound{
		Code:    http.StatusNotFound,
		Message: model.ErrOrderNotFound.Error(),
	}, res)
}

func (s *APISuite) TestPayOrderServiceError() {
	var (
		ctx        = context.Background()
		orderUUID  = uuid.New()
		serviceErr = errors.New("payment is unavailable")

		req    = &orderV1.PayOrderRequest{PaymentMethod: orderV1.PaymentMethodCARD}
		params = orderV1.PayOrderParams{OrderUUID: orderUUID}
	)

	s.orderService.EXPECT().Pay(ctx, orderUUID, model.PaymentMethodCard).Return(uuid.Nil, serviceErr)

	res, err := s.api.PayOrder(ctx, req, params)

	s.Require().Nil(res)
	s.Require().ErrorIs(err, serviceErr)
}
