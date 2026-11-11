package v1

import (
	"errors"
	"net/http"

	"github.com/google/uuid"

	"inventory-go/order/internal/model"
	orderV1 "inventory-go/shared/pkg/openapi/order/v1"
)

func (s *APISuite) TestCancelOrderSuccess() {
	var (
		ctx       = s.authCtx()
		orderUUID = uuid.New()
	)

	s.orderService.EXPECT().Cancel(ctx, orderUUID, s.userUUID).Return(nil)

	res, err := s.api.CancelOrder(ctx, orderV1.CancelOrderParams{OrderUUID: orderUUID})

	s.Require().NoError(err)
	s.Require().Equal(&orderV1.CancelOrderNoContent{}, res)
}

func (s *APISuite) TestCancelOrderNotFound() {
	var (
		ctx       = s.authCtx()
		orderUUID = uuid.New()
	)

	s.orderService.EXPECT().Cancel(ctx, orderUUID, s.userUUID).Return(model.ErrOrderNotFound)

	res, err := s.api.CancelOrder(ctx, orderV1.CancelOrderParams{OrderUUID: orderUUID})

	s.Require().NoError(err)
	s.Require().Equal(&orderV1.CancelOrderNotFound{
		Code:    http.StatusNotFound,
		Message: model.ErrOrderNotFound.Error(),
	}, res)
}

func (s *APISuite) TestCancelOrderAlreadyPaid() {
	var (
		ctx       = s.authCtx()
		orderUUID = uuid.New()
	)

	s.orderService.EXPECT().Cancel(ctx, orderUUID, s.userUUID).Return(model.ErrOrderPaid)

	res, err := s.api.CancelOrder(ctx, orderV1.CancelOrderParams{OrderUUID: orderUUID})

	s.Require().NoError(err)
	s.Require().Equal(&orderV1.CancelOrderConflict{
		Code:    http.StatusConflict,
		Message: model.ErrOrderPaid.Error(),
	}, res)
}

func (s *APISuite) TestCancelOrderServiceError() {
	var (
		ctx        = s.authCtx()
		orderUUID  = uuid.New()
		serviceErr = errors.New("database is unavailable")
	)

	s.orderService.EXPECT().Cancel(ctx, orderUUID, s.userUUID).Return(serviceErr)

	res, err := s.api.CancelOrder(ctx, orderV1.CancelOrderParams{OrderUUID: orderUUID})

	s.Require().Nil(res)
	s.Require().ErrorIs(err, serviceErr)
}
