//go:build e2e

package e2e

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// InsertTestPart inserts a part into the MongoDB collection and returns its UUID.
// The document layout mirrors the repository model of the service.
func (env *TestEnvironment) InsertTestPart(ctx context.Context, name string, category int32, country string) (string, error) {
	partUUID := uuid.NewString()
	now := time.Now().UTC()

	doc := bson.M{
		"_id":            partUUID,
		"name":           name,
		"description":    "part inserted by the e2e tests",
		"price":          42.5,
		"stock_quantity": int64(7),
		"category":       category,
		"manufacturer": bson.M{
			"name":    "E2E Corp",
			"country": country,
		},
		"tags":       bson.A{"e2e"},
		"created_at": now,
		"updated_at": now,
	}

	_, err := env.Mongo.Client().Database(mongoDatabase).Collection(partsCollectionName).InsertOne(ctx, doc)
	if err != nil {
		return "", err
	}

	return partUUID, nil
}

// DeleteTestPart removes a part from the MongoDB collection.
func (env *TestEnvironment) DeleteTestPart(ctx context.Context, partUUID string) error {
	_, err := env.Mongo.Client().Database(mongoDatabase).Collection(partsCollectionName).DeleteOne(ctx, bson.M{"_id": partUUID})

	return err
}
