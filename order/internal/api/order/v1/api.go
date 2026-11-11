package v1

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"

	"inventory-go/order/internal/model"
	"inventory-go/order/internal/service"
	grpcAuth "inventory-go/platform/pkg/middleware/grpc"
	orderV1 "inventory-go/shared/pkg/openapi/order/v1"
)

var _ orderV1.Handler = (*api)(nil)

type api struct {
	orderService service.OrderService
}

func NewAPI(orderService service.OrderService) *api {
	return &api{
		orderService: orderService,
	}
}

// currentUserUUID returns the user the auth middleware put into the context. Every order
// operation belongs to a user, so a request without one is refused instead of being served.
func currentUserUUID(ctx context.Context) (uuid.UUID, error) {
	user, ok := grpcAuth.GetUserFromContext(ctx)
	if !ok {
		return uuid.Nil, model.ErrUnauthenticated
	}

	userUUID, err := uuid.Parse(user.UUID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid user uuid in session: %w", err)
	}

	return userUUID, nil
}

// NewError is used for the common default response.
func (a *api) NewError(_ context.Context, err error) *orderV1.UnexpectedErrorStatusCode {
	statusCode := http.StatusInternalServerError
	if errors.Is(err, model.ErrUnauthenticated) {
		statusCode = http.StatusUnauthorized
	}

	return &orderV1.UnexpectedErrorStatusCode{
		StatusCode: statusCode,
		Response: orderV1.Error{
			Code:    int32(statusCode),
			Message: err.Error(),
		},
	}
}
