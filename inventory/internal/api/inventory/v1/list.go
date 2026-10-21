package v1

import (
	"context"

	"inventory-go/inventory/internal/converter"
	inventoryV1 "inventory-go/shared/pkg/proto/inventory/v1"
)

func (a *api) ListParts(ctx context.Context, req *inventoryV1.ListPartsRequest) (*inventoryV1.ListPartsResponse, error) {
	parts, err := a.inventoryService.List(ctx, converter.PartsFilterToModel(req.GetFilter()))
	if err != nil {
		return nil, err
	}

	return &inventoryV1.ListPartsResponse{
		Parts: converter.PartsToProto(parts),
	}, nil
}
