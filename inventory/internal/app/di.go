package app

import (
	"context"
	"fmt"
	"time"

	redigo "github.com/gomodule/redigo/redis"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	inventoryV1API "inventory-go/inventory/internal/api/inventory/v1"
	"inventory-go/inventory/internal/config"
	"inventory-go/inventory/internal/repository"
	partRepository "inventory-go/inventory/internal/repository/part"
	partCacheRepository "inventory-go/inventory/internal/repository/part_cache"
	"inventory-go/inventory/internal/service"
	partService "inventory-go/inventory/internal/service/part"
	"inventory-go/platform/pkg/cache"
	cacheRedis "inventory-go/platform/pkg/cache/redis"
	"inventory-go/platform/pkg/closer"
	"inventory-go/platform/pkg/logger"
	authV1 "inventory-go/shared/pkg/proto/auth/v1"
	inventoryV1 "inventory-go/shared/pkg/proto/inventory/v1"
)

const mongoPingTimeout = 5 * time.Second

// diContainer lazily builds the dependencies of the service and keeps a single instance of each.
// Constructors panic on a failure: the container is only used while the app is being assembled.
type diContainer struct {
	inventoryV1API inventoryV1.InventoryServiceServer

	inventoryService service.InventoryService

	inventoryRepository      repository.InventoryRepository
	inventoryCacheRepository repository.InventoryCacheRepository

	redisPool   *redigo.Pool
	redisClient cache.RedisClient

	mongoDBClient *mongo.Client
	mongoDBHandle *mongo.Database

	authClient authV1.AuthServiceClient
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
		d.inventoryService = partService.NewService(
			d.InventoryRepository(ctx),
			d.InventoryCacheRepository(ctx),
			config.AppConfig().Redis.CacheTTL(),
		)
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

func (d *diContainer) InventoryCacheRepository(ctx context.Context) repository.InventoryCacheRepository {
	if d.inventoryCacheRepository == nil {
		d.inventoryCacheRepository = partCacheRepository.NewRepository(d.RedisClient(ctx))
	}

	return d.inventoryCacheRepository
}

func (d *diContainer) RedisPool() *redigo.Pool {
	if d.redisPool == nil {
		cfg := config.AppConfig().Redis

		pool := cacheRedis.NewPool(cfg.Address(), cfg.MaxIdle(), cfg.IdleTimeout(), cfg.ConnectionTimeout())

		closer.AddNamed("Redis pool", func(_ context.Context) error {
			return pool.Close()
		})

		d.redisPool = pool
	}

	return d.redisPool
}

// RedisClient does not ping Redis on start: the cache is optional, the service must start
// and work (from MongoDB) even when Redis is down.
func (d *diContainer) RedisClient(_ context.Context) cache.RedisClient {
	if d.redisClient == nil {
		d.redisClient = cacheRedis.NewClient(d.RedisPool(), logger.Default(), config.AppConfig().Redis.ConnectionTimeout())
	}

	return d.redisClient
}

// AuthClient is the client of the auth service, used by the auth interceptor to check sessions.
// The connection is lazy: the service starts even when auth is not up yet, and requests are
// rejected as unauthenticated until it is.
func (d *diContainer) AuthClient(_ context.Context) authV1.AuthServiceClient {
	if d.authClient == nil {
		conn, err := grpc.NewClient(
			config.AppConfig().Auth.Address(),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			panic(fmt.Sprintf("failed to connect to auth: %v", err))
		}

		closer.AddNamed("auth gRPC connection", func(_ context.Context) error {
			return conn.Close()
		})

		d.authClient = authV1.NewAuthServiceClient(conn)
	}

	return d.authClient
}
