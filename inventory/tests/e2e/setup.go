//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"os"

	"github.com/testcontainers/testcontainers-go/wait"

	"inventory-go/platform/pkg/logger"
	"inventory-go/platform/pkg/testcontainers/app"
	"inventory-go/platform/pkg/testcontainers/authstack"
	"inventory-go/platform/pkg/testcontainers/mongo"
	"inventory-go/platform/pkg/testcontainers/network"
	"inventory-go/platform/pkg/testcontainers/path"
	"inventory-go/platform/pkg/testcontainers/redis"
)

// TestEnvironment holds the resources of the test environment.
type TestEnvironment struct {
	Network *network.Network
	Mongo   *mongo.Container
	Redis   *redis.Container
	Auth    *authstack.Stack
	App     *app.Container
}

// setupTestEnvironment prepares the test environment: the network, MongoDB and the inventory service.
func setupTestEnvironment(ctx context.Context) (*TestEnvironment, error) {
	logger.Info(ctx, "Preparing the test environment...")

	// Step 1: create a shared Docker network
	generatedNetwork, err := network.NewNetwork(ctx, projectName)
	if err != nil {
		return nil, fmt.Errorf("failed to create the network: %w", err)
	}
	logger.Info(ctx, "Network created")

	// Step 2: start MongoDB
	generatedMongo, err := mongo.NewContainer(ctx,
		mongo.WithNetworkName(generatedNetwork.Name()),
		mongo.WithContainerName(mongoName),
		mongo.WithImageName(mongoImage),
		mongo.WithDatabase(mongoDatabase),
		mongo.WithAuth(mongoUser, mongoPassword),
		mongo.WithAuthDB(mongoAuthDB),
		mongo.WithLogger(logger.Default()),
	)
	if err != nil {
		cleanupTestEnvironment(ctx, &TestEnvironment{Network: generatedNetwork})
		return nil, fmt.Errorf("failed to start the MongoDB container: %w", err)
	}
	logger.Info(ctx, "MongoDB container started")

	// Step 2.1: start Redis (the cache)
	generatedRedis, err := redis.NewContainer(ctx,
		redis.WithNetworkName(generatedNetwork.Name()),
		redis.WithContainerName(redisName),
		redis.WithImageName(redisImage),
		redis.WithLogger(logger.Default()),
	)
	if err != nil {
		cleanupTestEnvironment(ctx, &TestEnvironment{Network: generatedNetwork, Mongo: generatedMongo})
		return nil, fmt.Errorf("failed to start the Redis container: %w", err)
	}
	logger.Info(ctx, "Redis container started")

	// Step 2.2: start the auth service (the inventory service validates sessions against it)
	generatedAuth, err := authstack.Start(ctx, generatedNetwork.Name(), projectName)
	if err != nil {
		cleanupTestEnvironment(ctx, &TestEnvironment{Network: generatedNetwork, Mongo: generatedMongo, Redis: generatedRedis})
		return nil, fmt.Errorf("failed to start the auth service: %w", err)
	}
	logger.Info(ctx, "Auth service started")

	// Step 3: build and start the service
	appEnv := map[string]string{
		"GRPC_HOST": "0.0.0.0",
		"GRPC_PORT": grpcPort,
		"HTTP_HOST": "0.0.0.0",
		"HTTP_PORT": httpPort,

		"LOGGER_LEVEL":   loggerLevelValue,
		"LOGGER_AS_JSON": "true",

		// the service reaches MongoDB by the container name inside the Docker network
		"MONGO_HOST":                 generatedMongo.Config().ContainerName,
		"MONGO_PORT":                 mongoPort,
		"MONGO_INITDB_DATABASE":      mongoDatabase,
		"MONGO_INITDB_ROOT_USERNAME": mongoUser,
		"MONGO_INITDB_ROOT_PASSWORD": mongoPassword,
		"MONGO_AUTH_DB":              mongoAuthDB,

		"REDIS_HOST":               redisName,
		"REDIS_PORT":               redisPort,
		"REDIS_CONNECTION_TIMEOUT": redisConnectionTimeout,
		"REDIS_MAX_IDLE":           redisMaxIdle,
		"REDIS_IDLE_TIMEOUT":       redisIdleTimeout,
		"REDIS_CACHE_TTL":          redisCacheTTL,

		"AUTH_GRPC_HOST": generatedAuth.Host,
		"AUTH_GRPC_PORT": authstack.GRPCPort,
	}

	waitStrategy := wait.ForListeningPort(grpcPort + "/tcp").WithStartupTimeout(startupTimeout)

	appContainer, err := app.NewContainer(ctx,
		app.WithName(appName),
		app.WithPort(grpcPort),
		app.WithDockerfile(path.GetProjectRoot(), dockerfile),
		app.WithBuildArg("SERVICE", serviceBuild),
		app.WithBuildArg("CMD", cmdBuild),
		app.WithNetwork(generatedNetwork.Name()),
		app.WithEnv(appEnv),
		app.WithLogOutput(os.Stdout),
		app.WithStartupWait(waitStrategy),
		app.WithLogger(logger.Default()),
	)
	if err != nil {
		cleanupTestEnvironment(ctx, &TestEnvironment{Network: generatedNetwork, Mongo: generatedMongo, Redis: generatedRedis, Auth: generatedAuth})
		return nil, fmt.Errorf("failed to start the app container: %w", err)
	}
	logger.Info(ctx, "App container started")

	logger.Info(ctx, "Test environment is ready")

	return &TestEnvironment{
		Network: generatedNetwork,
		Mongo:   generatedMongo,
		Redis:   generatedRedis,
		Auth:    generatedAuth,
		App:     appContainer,
	}, nil
}
