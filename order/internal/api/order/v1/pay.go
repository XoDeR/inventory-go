package v1

import (
	"context"
	"errors"
	"net/http"

	"inventory-go/order/internal/converter"
	"inventory-go/order/internal/model"
	orderV1 "inventory-go/shared/pkg/openapi/order/v1"
)

// POST /api/v1/orders/{order_uuid}/pay
func (a *api) PayOrder(ctx context.Context, req *orderV1.PayOrderRequest, params orderV1.PayOrderParams) (orderV1.PayOrderRes, error) {
	transactionUUID, err := a.orderService.Pay(ctx, params.OrderUUID, converter.PaymentMethodToModel(req.PaymentMethod))
	if err != nil {
		if errors.Is(err, model.ErrOrderNotFound) {
			return &orderV1.PayOrderNotFound{
				Code:    http.StatusNotFound,
				Message: err.Error(),
			}, nil
		}
		return nil, err
	}

	return &orderV1.PayOrderResponse{
		TransactionUUID: transactionUUID,
	}, nil
}
