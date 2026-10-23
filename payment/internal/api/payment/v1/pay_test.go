package v1

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"inventory-go/payment/internal/model"
	paymentV1 "inventory-go/shared/pkg/proto/payment/v1"
)

func (s *APISuite) TestPayOrderSuccess() {
	var (
		ctx           = context.Background()
		transactionID = "5d6c1e3a-7b1f-4a2e-8c3d-9e0f1a2b3c4d"

		req = &paymentV1.PayOrderRequest{
			OrderUuid:     "order-uuid",
			UserUuid:      "user-uuid",
			PaymentMethod: paymentV1.PaymentMethod_PAYMENT_METHOD_CREDIT_CARD,
		}

		// The proto request must be converted to the model before it reaches the service.
		expectedInfo = model.PayOrderInfo{
			OrderUuid:     "order-uuid",
			UserUuid:      "user-uuid",
			PaymentMethod: model.PaymentMethodCreditCard,
		}
	)

	s.paymentService.EXPECT().PayOrder(ctx, expectedInfo).Return(transactionID, nil)

	resp, err := s.api.PayOrder(ctx, req)

	s.Require().NoError(err)
	s.Require().Equal(transactionID, resp.GetTransactionUuid())
}

func (s *APISuite) TestPayOrderInvalidArgument() {
	var (
		ctx = context.Background()
		req = &paymentV1.PayOrderRequest{UserUuid: "user-uuid"}

		expectedInfo = model.PayOrderInfo{UserUuid: "user-uuid"}
		serviceErr   = fmt.Errorf("%w: order_uuid is required", model.ErrInvalidArgument)
	)

	s.paymentService.EXPECT().PayOrder(ctx, expectedInfo).Return("", serviceErr)

	resp, err := s.api.PayOrder(ctx, req)

	s.Require().Nil(resp)
	s.Require().Equal(codes.InvalidArgument, status.Code(err))
	s.Require().Equal(serviceErr.Error(), status.Convert(err).Message())
}

func (s *APISuite) TestPayOrderServiceError() {
	var (
		ctx        = context.Background()
		req        = &paymentV1.PayOrderRequest{OrderUuid: "order-uuid", UserUuid: "user-uuid"}
		serviceErr = errors.New("payment provider is unavailable")

		expectedInfo = model.PayOrderInfo{OrderUuid: "order-uuid", UserUuid: "user-uuid"}
	)

	s.paymentService.EXPECT().PayOrder(ctx, expectedInfo).Return("", serviceErr)

	resp, err := s.api.PayOrder(ctx, req)

	s.Require().Nil(resp)
	s.Require().ErrorIs(err, serviceErr)
}
