package v1

import (
	"context"
	"errors"
	"net/http"

	"inventory-go/order/internal/converter"
	"inventory-go/order/internal/model"
	orderV1 "inventory-go/shared/pkg/openapi/order/v1"
)

// POST /api/v1/orders
func (a *api) CreateOrder(ctx context.Context, req *orderV1.CreateOrderRequest) (orderV1.CreateOrderRes, error) {
	order, err := a.orderService.Create(ctx, converter.CreateOrderInfoToModel(req))
	if err != nil {
		if errors.Is(err, model.ErrEmptyPartUuids) || errors.Is(err, model.ErrPartsNotFound) {
			return &orderV1.CreateOrderBadRequest{
				Code:    http.StatusBadRequest,
				Message: err.Error(),
			}, nil
		}
		return nil, err
	}

	return &orderV1.CreateOrderResponse{
		UUID:       order.Uuid,
		TotalPrice: order.TotalPrice,
	}, nil
}
