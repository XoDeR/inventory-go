package v1

import (
	"context"

	clientConverter "inventory-go/order/internal/client/converter"
	"inventory-go/order/internal/model"
)

func (c *client) PayOrder(ctx context.Context, info model.PayOrderInfo) (string, error) {
	resp, err := c.generatedClient.PayOrder(ctx, clientConverter.PayOrderInfoToProto(info))
	if err != nil {
		return "", err
	}

	return resp.GetTransactionUuid(), nil
}
