package v1

import (
	"context"

	clientConverter "inventory-go/order/internal/client/converter"
	"inventory-go/order/internal/model"
	inventoryV1 "inventory-go/shared/pkg/proto/inventory/v1"
)

func (c *client) ListParts(ctx context.Context, filter model.PartsFilter) ([]model.Part, error) {
	resp, err := c.generatedClient.ListParts(ctx, &inventoryV1.ListPartsRequest{
		Filter: clientConverter.PartsFilterToProto(filter),
	})
	if err != nil {
		return nil, err
	}

	return clientConverter.PartsToModel(resp.GetParts()), nil
}
