package converter

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	"inventory-go/inventory/internal/model"
	inventoryV1 "inventory-go/shared/pkg/proto/inventory/v1"
)

func PartsFilterToModel(filter *inventoryV1.PartsFilter) model.PartsFilter {
	if filter == nil {
		return model.PartsFilter{}
	}

	categories := make([]model.Category, 0, len(filter.GetCategories()))
	for _, category := range filter.GetCategories() {
		categories = append(categories, model.Category(category))
	}

	return model.PartsFilter{
		Uuids:                 filter.GetUuids(),
		Names:                 filter.GetNames(),
		Categories:            categories,
		ManufacturerCountries: filter.GetManufacturerCountries(),
		Tags:                  filter.GetTags(),
	}
}

func PartsToProto(parts []model.Part) []*inventoryV1.Part {
	result := make([]*inventoryV1.Part, 0, len(parts))
	for _, part := range parts {
		result = append(result, PartToProto(part))
	}

	return result
}

func PartToProto(part model.Part) *inventoryV1.Part {
	var dimensions *inventoryV1.Dimensions
	if part.Dimensions != nil {
		dimensions = &inventoryV1.Dimensions{
			Length: part.Dimensions.Length,
			Width:  part.Dimensions.Width,
			Height: part.Dimensions.Height,
			Weight: part.Dimensions.Weight,
		}
	}

	var manufacturer *inventoryV1.Manufacturer
	if part.Manufacturer != nil {
		manufacturer = &inventoryV1.Manufacturer{
			Name:    part.Manufacturer.Name,
			Country: part.Manufacturer.Country,
			Website: part.Manufacturer.Website,
		}
	}

	var metadata map[string]*inventoryV1.Value
	if part.Metadata != nil {
		metadata = make(map[string]*inventoryV1.Value, len(part.Metadata))
		for key, value := range part.Metadata {
			if v := valueToProto(value); v != nil {
				metadata[key] = v
			}
		}
	}

	var createdAt *timestamppb.Timestamp
	if part.CreatedAt != nil {
		createdAt = timestamppb.New(*part.CreatedAt)
	}

	var updatedAt *timestamppb.Timestamp
	if part.UpdatedAt != nil {
		updatedAt = timestamppb.New(*part.UpdatedAt)
	}

	return &inventoryV1.Part{
		Uuid:          part.Uuid,
		Name:          part.Name,
		Description:   part.Description,
		Price:         part.Price,
		StockQuantity: part.StockQuantity,
		Category:      inventoryV1.Category(part.Category),
		Dimensions:    dimensions,
		Manufacturer:  manufacturer,
		Tags:          part.Tags,
		Metadata:      metadata,
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
	}
}

func valueToProto(value any) *inventoryV1.Value {
	switch v := value.(type) {
	case string:
		return &inventoryV1.Value{Kind: &inventoryV1.Value_StringValue{StringValue: v}}
	case int64:
		return &inventoryV1.Value{Kind: &inventoryV1.Value_Int64Value{Int64Value: v}}
	case float64:
		return &inventoryV1.Value{Kind: &inventoryV1.Value_DoubleValue{DoubleValue: v}}
	case bool:
		return &inventoryV1.Value{Kind: &inventoryV1.Value_BoolValue{BoolValue: v}}
	default:
		return nil
	}
}
