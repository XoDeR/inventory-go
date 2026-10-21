package v1

import (
	"context"
	"net/http"

	"inventory-go/order/internal/service"
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

// NewError is used for the common default response.
func (a *api) NewError(_ context.Context, err error) *orderV1.UnexpectedErrorStatusCode {
	return &orderV1.UnexpectedErrorStatusCode{
		StatusCode: http.StatusInternalServerError,
		Response: orderV1.Error{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		},
	}
}
