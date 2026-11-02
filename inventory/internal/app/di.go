package app

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"

	inventoryV1API "inventory-go/inventory/internal/api/inventory/v1"
	"inventory-go/inventory/internal/config"
	"inventory-go/inventory/internal/repository"
	partRepository "inventory-go/inventory/internal/repository/part"
	"inventory-go/inventory/internal/service"
	partService "inventory-go/inventory/internal/service/part"
	"inventory-go/platform/pkg/closer"
	inventoryV1 "inventory-go/shared/pkg/proto/inventory/v1"
)

const mongoPingTimeout = 5 * time.Second

// diContainer lazily builds the dependencies of the service and keeps a single instance of each.
// Constructors panic on a failure: the container is only used while the app is being assembled.
type diContainer struct {
	inventoryV1API inventoryV1.InventoryServiceServer

	inventoryService service.InventoryService

	inventoryRepository repository.InventoryRepository

	mongoDBClient *mongo.Client
	mongoDBHandle *mongo.Database
}

func NewDiContainer() *diContainer {
	return &diContainer{}
}

func (d *diContainer) InventoryV1API(ctx context.Context) inventoryV1.InventoryServiceServer {
	if d.inventoryV1API == nil {
		d.inventoryV1API = inventoryV1API.NewAPI(d.InventoryService(ctx))
	}

	return d.inventoryV1API
}

func (d *diContainer) InventoryService(ctx context.Context) service.InventoryService {
	if d.inventoryService == nil {
		d.inventoryService = partService.NewService(d.InventoryRepository(ctx))
	}

	return d.inventoryService
}

func (d *diContainer) InventoryRepository(ctx context.Context) repository.InventoryRepository {
	if d.inventoryRepository == nil {
		repo := partRepository.NewRepository(d.MongoDBHandle(ctx))
		if err := repo.Seed(ctx); err != nil {
			panic(fmt.Sprintf("failed to seed the inventory: %v", err))
		}

		d.inventoryRepository = repo
	}

	return d.inventoryRepository
}

func (d *diContainer) MongoDBClient(ctx context.Context) *mongo.Client {
	if d.mongoDBClient == nil {
		client, err := mongo.Connect(options.Client().ApplyURI(config.AppConfig().Mongo.URI()))
		if err != nil {
			panic(fmt.Sprintf("failed to connect to MongoDB: %v", err))
		}

		pingCtx, cancel := context.WithTimeout(ctx, mongoPingTimeout)
		defer cancel()

		if err = client.Ping(pingCtx, readpref.Primary()); err != nil {
			panic(fmt.Sprintf("failed to ping MongoDB: %v", err))
		}

		closer.AddNamed("MongoDB client", func(ctx context.Context) error {
			return client.Disconnect(ctx)
		})

		d.mongoDBClient = client
	}

	return d.mongoDBClient
}

func (d *diContainer) MongoDBHandle(ctx context.Context) *mongo.Database {
	if d.mongoDBHandle == nil {
		d.mongoDBHandle = d.MongoDBClient(ctx).Database(config.AppConfig().Mongo.DatabaseName())
	}

	return d.mongoDBHandle
}
