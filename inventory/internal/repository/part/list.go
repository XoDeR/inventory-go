package part

import (
	"context"
	"slices"

	"inventory-go/inventory/internal/model"
	repoConverter "inventory-go/inventory/internal/repository/converter"
	repoModel "inventory-go/inventory/internal/repository/model"
)

// List returns parts matching the filter: OR within a single filter field,
// AND between different fields. An empty field is not applied.
func (r *repository) List(_ context.Context, filter model.PartsFilter) ([]model.Part, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	parts := make([]model.Part, 0, len(r.data))
	for _, part := range r.data {
		if matchesFilter(part, filter) {
			parts = append(parts, repoConverter.PartToModel(part))
		}
	}

	return parts, nil
}

func matchesFilter(part repoModel.Part, filter model.PartsFilter) bool {
	if len(filter.Uuids) > 0 && !slices.Contains(filter.Uuids, part.Uuid) {
		return false
	}
	if len(filter.Names) > 0 && !slices.Contains(filter.Names, part.Name) {
		return false
	}
	if len(filter.Categories) > 0 && !slices.Contains(filter.Categories, model.Category(part.Category)) {
		return false
	}
	if len(filter.ManufacturerCountries) > 0 {
		country := ""
		if part.Manufacturer != nil {
			country = part.Manufacturer.Country
		}
		if !slices.Contains(filter.ManufacturerCountries, country) {
			return false
		}
	}
	if len(filter.Tags) > 0 && !slices.ContainsFunc(part.Tags, func(tag string) bool {
		return slices.Contains(filter.Tags, tag)
	}) {
		return false
	}

	return true
}
