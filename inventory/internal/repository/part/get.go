package part

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"inventory-go/inventory/internal/model"
	repoConverter "inventory-go/inventory/internal/repository/converter"
	repoModel "inventory-go/inventory/internal/repository/model"
)

func (r *repository) Get(ctx context.Context, uuid string) (model.Part, error) {
	var part repoModel.Part

	err := r.collection.FindOne(ctx, bson.M{"_id": uuid}).Decode(&part)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return model.Part{}, model.ErrPartNotFound
		}

		return model.Part{}, err
	}

	return repoConverter.PartToModel(part), nil
}
