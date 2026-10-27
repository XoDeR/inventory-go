package part

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"

	"inventory-go/inventory/internal/model"
	repoConverter "inventory-go/inventory/internal/repository/converter"
	repoModel "inventory-go/inventory/internal/repository/model"
)

// List returns parts matching the filter: OR within a single filter field,
// AND between different fields. An empty field is not applied.
func (r *repository) List(ctx context.Context, filter model.PartsFilter) ([]model.Part, error) {
	cursor, err := r.collection.Find(ctx, buildFilter(filter))
	if err != nil {
		return nil, err
	}

	var repoParts []repoModel.Part
	if err = cursor.All(ctx, &repoParts); err != nil {
		return nil, err
	}

	parts := make([]model.Part, 0, len(repoParts))
	for _, part := range repoParts {
		parts = append(parts, repoConverter.PartToModel(part))
	}

	return parts, nil
}

// buildFilter turns the filter into a MongoDB query. Top-level conditions are combined with AND,
// and $in matches any of the values, which gives OR within a field (for the tags array $in
// matches when any element is in the list).
func buildFilter(filter model.PartsFilter) bson.M {
	query := bson.M{}

	if len(filter.Uuids) > 0 {
		query["_id"] = bson.M{"$in": filter.Uuids}
	}
	if len(filter.Names) > 0 {
		query["name"] = bson.M{"$in": filter.Names}
	}
	if len(filter.Categories) > 0 {
		query["category"] = bson.M{"$in": filter.Categories}
	}
	if len(filter.ManufacturerCountries) > 0 {
		query["manufacturer.country"] = bson.M{"$in": filter.ManufacturerCountries}
	}
	if len(filter.Tags) > 0 {
		query["tags"] = bson.M{"$in": filter.Tags}
	}

	return query
}
