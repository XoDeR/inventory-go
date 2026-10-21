package v1

import (
	"context"
	"errors"
	"net/http"

	"inventory-go/order/internal/model"
	orderV1 "inventory-go/shared/pkg/openapi/order/v1"
)

// POST /api/v1/orders/{order_uuid}/cancel
func (a *api) CancelOrder(ctx context.Context, params orderV1.CancelOrderParams) (orderV1.CancelOrderRes, error) {
	err := a.orderService.Cancel(ctx, params.OrderUUID)
	if err != nil {
		switch {
		case errors.Is(err, model.ErrOrderNotFound):
			return &orderV1.CancelOrderNotFound{
				Code:    http.StatusNotFound,
				Message: err.Error(),
			}, nil
		case errors.Is(err, model.ErrOrderPaid):
			return &orderV1.CancelOrderConflict{
				Code:    http.StatusConflict,
				Message: err.Error(),
			}, nil
		}
		return nil, err
	}

	return &orderV1.CancelOrderNoContent{}, nil
}
