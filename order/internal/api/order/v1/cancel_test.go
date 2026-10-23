package v1

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"

	"inventory-go/order/internal/model"
	orderV1 "inventory-go/shared/pkg/openapi/order/v1"
)

func (s *APISuite) TestCancelOrderSuccess() {
	var (
		ctx       = context.Background()
		orderUUID = uuid.New()
	)

	s.orderService.EXPECT().Cancel(ctx, orderUUID).Return(nil)

	res, err := s.api.CancelOrder(ctx, orderV1.CancelOrderParams{OrderUUID: orderUUID})

	s.Require().NoError(err)
	s.Require().Equal(&orderV1.CancelOrderNoContent{}, res)
}

func (s *APISuite) TestCancelOrderNotFound() {
	var (
		ctx       = context.Background()
		orderUUID = uuid.New()
	)

	s.orderService.EXPECT().Cancel(ctx, orderUUID).Return(model.ErrOrderNotFound)

	res, err := s.api.CancelOrder(ctx, orderV1.CancelOrderParams{OrderUUID: orderUUID})

	s.Require().NoError(err)
	s.Require().Equal(&orderV1.CancelOrderNotFound{
		Code:    http.StatusNotFound,
		Message: model.ErrOrderNotFound.Error(),
	}, res)
}

func (s *APISuite) TestCancelOrderAlreadyPaid() {
	var (
		ctx       = context.Background()
		orderUUID = uuid.New()
	)

	s.orderService.EXPECT().Cancel(ctx, orderUUID).Return(model.ErrOrderPaid)

	res, err := s.api.CancelOrder(ctx, orderV1.CancelOrderParams{OrderUUID: orderUUID})

	s.Require().NoError(err)
	s.Require().Equal(&orderV1.CancelOrderConflict{
		Code:    http.StatusConflict,
		Message: model.ErrOrderPaid.Error(),
	}, res)
}

func (s *APISuite) TestCancelOrderServiceError() {
	var (
		ctx        = context.Background()
		orderUUID  = uuid.New()
		serviceErr = errors.New("database is unavailable")
	)

	s.orderService.EXPECT().Cancel(ctx, orderUUID).Return(serviceErr)

	res, err := s.api.CancelOrder(ctx, orderV1.CancelOrderParams{OrderUUID: orderUUID})

	s.Require().Nil(res)
	s.Require().ErrorIs(err, serviceErr)
}
