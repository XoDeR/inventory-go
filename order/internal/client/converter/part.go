package converter

import (
	"inventory-go/order/internal/model"
	inventoryV1 "inventory-go/shared/pkg/proto/inventory/v1"
)

func PartsFilterToProto(filter model.PartsFilter) *inventoryV1.PartsFilter {
	return &inventoryV1.PartsFilter{
		Uuids: filter.Uuids,
	}
}

func PartsToModel(parts []*inventoryV1.Part) []model.Part {
	result := make([]model.Part, 0, len(parts))
	for _, part := range parts {
		result = append(result, PartToModel(part))
	}

	return result
}

func PartToModel(part *inventoryV1.Part) model.Part {
	return model.Part{
		Uuid:  part.GetUuid(),
		Price: part.GetPrice(),
	}
}
