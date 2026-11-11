package v1

import (
	"context"

	"github.com/google/uuid"

	"inventory-go/order/internal/model"
	grpcAuth "inventory-go/platform/pkg/middleware/grpc"
	orderV1 "inventory-go/shared/pkg/openapi/order/v1"
)

// Without an authenticated user no operation reaches the service: the service mock has no
// expectations, so any call to it would fail the test.
func (s *APISuite) TestOperationsRequireAuthenticatedUser() {
	ctx := context.Background()
	orderUUID := uuid.New()

	_, err := s.api.CreateOrder(ctx, &orderV1.CreateOrderRequest{PartUuids: []uuid.UUID{uuid.New()}})
	s.Require().ErrorIs(err, model.ErrUnauthenticated)

	_, err = s.api.GetOrderByUuid(ctx, orderV1.GetOrderByUuidParams{OrderUUID: orderUUID})
	s.Require().ErrorIs(err, model.ErrUnauthenticated)

	_, err = s.api.PayOrder(ctx, &orderV1.PayOrderRequest{PaymentMethod: orderV1.PaymentMethodCARD}, orderV1.PayOrderParams{OrderUUID: orderUUID})
	s.Require().ErrorIs(err, model.ErrUnauthenticated)

	_, err = s.api.CancelOrder(ctx, orderV1.CancelOrderParams{OrderUUID: orderUUID})
	s.Require().ErrorIs(err, model.ErrUnauthenticated)
}

func (s *APISuite) TestMalformedSessionUserIsRejected() {
	ctx := grpcAuth.WithSession(context.Background(), &grpcAuth.User{UUID: "not-a-uuid"}, uuid.NewString())

	_, err := s.api.GetOrderByUuid(ctx, orderV1.GetOrderByUuidParams{OrderUUID: uuid.New()})

	s.Require().Error(err)
	s.Require().NotErrorIs(err, model.ErrUnauthenticated)
}

func (s *APISuite) TestNewErrorMapsUnauthenticatedTo401() {
	res := s.api.NewError(context.Background(), model.ErrUnauthenticated)

	s.Require().Equal(401, res.StatusCode)
	s.Require().EqualValues(401, res.Response.Code)
}
