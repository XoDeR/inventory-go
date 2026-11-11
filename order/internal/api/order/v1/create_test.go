package v1

import (
	"errors"
	"net/http"

	"github.com/google/uuid"

	"inventory-go/order/internal/model"
	orderV1 "inventory-go/shared/pkg/openapi/order/v1"
)

func (s *APISuite) TestCreateOrderSuccess() {
	var (
		ctx = s.authCtx()

		partUUID = uuid.New()

		// the body names another user: it must be ignored in favour of the authenticated one
		req = &orderV1.CreateOrderRequest{UserUUID: uuid.New(), PartUuids: []uuid.UUID{partUUID}}

		// The request must be converted to the model before it reaches the service,
		// with the user of the session.
		expectedInfo = model.CreateOrderInfo{UserUuid: s.userUUID, PartUuids: []uuid.UUID{partUUID}}

		order = model.Order{Uuid: uuid.New(), UserUuid: s.userUUID, TotalPrice: 1999.99}
	)

	s.orderService.EXPECT().Create(ctx, expectedInfo).Return(order, nil)

	res, err := s.api.CreateOrder(ctx, req)

	s.Require().NoError(err)
	s.Require().Equal(&orderV1.CreateOrderResponse{UUID: order.Uuid, TotalPrice: order.TotalPrice}, res)
}

func (s *APISuite) TestCreateOrderBadRequest() {
	for name, serviceErr := range map[string]error{
		"empty part uuids": model.ErrEmptyPartUuids,
		"parts not found":  model.ErrPartsNotFound,
	} {
		s.Run(name, func() {
			ctx := s.authCtx()
			req := &orderV1.CreateOrderRequest{UserUUID: uuid.New()}

			s.orderService.EXPECT().Create(ctx, model.CreateOrderInfo{UserUuid: s.userUUID}).
				Return(model.Order{}, serviceErr).Once()

			res, err := s.api.CreateOrder(ctx, req)

			s.Require().NoError(err)
			s.Require().Equal(&orderV1.CreateOrderBadRequest{
				Code:    http.StatusBadRequest,
				Message: serviceErr.Error(),
			}, res)
		})
	}
}

func (s *APISuite) TestCreateOrderServiceError() {
	var (
		ctx        = s.authCtx()
		req        = &orderV1.CreateOrderRequest{UserUUID: uuid.New()}
		serviceErr = errors.New("database is unavailable")
	)

	s.orderService.EXPECT().Create(ctx, model.CreateOrderInfo{UserUuid: s.userUUID}).Return(model.Order{}, serviceErr)

	res, err := s.api.CreateOrder(ctx, req)

	s.Require().Nil(res)
	s.Require().ErrorIs(err, serviceErr)
}
