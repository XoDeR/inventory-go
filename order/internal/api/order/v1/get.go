package v1

import (
	"context"
	"errors"
	"net/http"

	"inventory-go/order/internal/converter"
	"inventory-go/order/internal/model"
	orderV1 "inventory-go/shared/pkg/openapi/order/v1"
)

// GET /api/v1/orders/{order_uuid}
func (a *api) GetOrderByUuid(ctx context.Context, params orderV1.GetOrderByUuidParams) (orderV1.GetOrderByUuidRes, error) {
	userUUID, err := currentUserUUID(ctx)
	if err != nil {
		return nil, err
	}

	order, err := a.orderService.Get(ctx, params.OrderUUID, userUUID)
	if err != nil {
		if errors.Is(err, model.ErrOrderNotFound) {
			return &orderV1.GetOrderByUuidNotFound{
				Code:    http.StatusNotFound,
				Message: err.Error(),
			}, nil
		}
		return nil, err
	}

	return converter.OrderToOpenAPI(order), nil
}
