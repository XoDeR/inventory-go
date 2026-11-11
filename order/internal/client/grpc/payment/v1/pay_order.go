package v1

import (
	"context"

	clientConverter "inventory-go/order/internal/client/converter"
	"inventory-go/order/internal/model"
	grpcAuth "inventory-go/platform/pkg/middleware/grpc"
)

func (c *client) PayOrder(ctx context.Context, info model.PayOrderInfo) (string, error) {
	resp, err := c.generatedClient.PayOrder(grpcAuth.ForwardSessionUUIDToGRPC(ctx), clientConverter.PayOrderInfoToProto(info))
	if err != nil {
		return "", err
	}

	return resp.GetTransactionUuid(), nil
}
