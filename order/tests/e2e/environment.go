//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/testcontainers/testcontainers-go/wait"

	"inventory-go/platform/pkg/logger"
	"inventory-go/platform/pkg/testcontainers/app"
	"inventory-go/platform/pkg/testcontainers/kafka"
	"inventory-go/platform/pkg/testcontainers/mongo"
	"inventory-go/platform/pkg/testcontainers/network"
	"inventory-go/platform/pkg/testcontainers/path"
	"inventory-go/platform/pkg/testcontainers/postgres"
	"inventory-go/platform/pkg/testcontainers/redis"
)

const (
	// projectName is the name of the project for the Docker containers and the network
	projectName = "order-service-e2e"

	dockerfile = "deploy/docker/Dockerfile"

	// Order service
	orderAppName = "e2e-order-app"
	orderHTTP    = "8080"

	// Inventory service
	inventoryAppName = "e2e-order-inventory-app"
	inventoryGRPC    = "50051"
	inventoryHTTP    = "8081"

	// Payment service
	paymentAppName = "e2e-order-payment-app"
	paymentGRPC    = "50052"
	paymentHTTP    = "8082"

	// Assembly service (a Kafka consumer without a server port)
	assemblyAppName  = "e2e-order-assembly-app"
	assemblyDuration = "2s"

	// Kafka
	kafkaName  = "e2e-order-kafka"
	kafkaImage = "confluentinc/cp-kafka:7.9.0"

	orderPaidTopic         = "order.paid"
	computerAssembledTopic = "assembly.computer-assembled"

	// PostgreSQL of the order service
	postgresName     = "e2e-order-postgres"
	postgresImage    = "postgres:17-alpine"
	postgresDatabase = "order-service"
	postgresUser     = "order-service-user"
	postgresPassword = "order-service-password"
	postgresPort     = "5432"

	// MongoDB of the inventory service
	mongoName     = "e2e-order-mongo"
	mongoImage    = "mongo:7.0.5"
	mongoDatabase = "inventory-service"
	mongoUser     = "inventory-service-user"
	mongoPassword = "inventory-service-password"
	mongoAuthDB   = "admin"
	mongoPort     = "27017"

	// Redis shared by the inventory and order caches (their keys have different prefixes)
	redisName  = "e2e-order-redis"
	redisImage = "redis:7.2.5-alpine3.20"
	redisPort  = "6379"

	redisConnectionTimeout = "5s"
	redisMaxIdle           = "10"
	redisIdleTimeout       = "240s"
	redisCacheTTL          = "1h"

	loggerLevelValue = "debug"
	startupTimeout   = 3 * time.Minute
	testsTimeout     = 15 * time.Minute
)

// TestEnvironment holds the resources of the test environment.
type TestEnvironment struct {
	Network   *network.Network
	Postgres  *postgres.Container
	Kafka     *kafka.Container
	Mongo     *mongo.Container
	Redis     *redis.Container
	Inventory *app.Container
	Payment   *app.Container
	Assembly  *app.Container
	Order     *app.Container
}

// setupTestEnvironment prepares the whole chain: PostgreSQL and MongoDB, then the inventory,
// payment and order services. The containers are started one after another to keep the load low.
func setupTestEnvironment(ctx context.Context) (*TestEnvironment, error) {
	logger.Info(ctx, "Preparing the test environment...")

	env := &TestEnvironment{}

	fail := func(msg string, err error) (*TestEnvironment, error) {
		cleanupTestEnvironment(ctx, env)
		return nil, fmt.Errorf("%s: %w", msg, err)
	}

	var err error

	env.Network, err = network.NewNetwork(ctx, projectName)
	if err != nil {
		return fail("failed to create the network", err)
	}
	logger.Info(ctx, "Network created")

	env.Postgres, err = postgres.NewContainer(ctx,
		postgres.WithNetworkName(env.Network.Name()),
		postgres.WithContainerName(postgresName),
		postgres.WithImageName(postgresImage),
		postgres.WithDatabase(postgresDatabase),
		postgres.WithAuth(postgresUser, postgresPassword),
		postgres.WithLogger(logger.Default()),
	)
	if err != nil {
		return fail("failed to start the PostgreSQL container", err)
	}

	env.Kafka, err = kafka.NewContainer(ctx,
		kafka.WithNetworkName(env.Network.Name()),
		kafka.WithContainerName(kafkaName),
		kafka.WithImageName(kafkaImage),
		kafka.WithLogger(logger.Default()),
	)
	if err != nil {
		return fail("failed to start the Kafka container", err)
	}

	env.Mongo, err = mongo.NewContainer(ctx,
		mongo.WithNetworkName(env.Network.Name()),
		mongo.WithContainerName(mongoName),
		mongo.WithImageName(mongoImage),
		mongo.WithDatabase(mongoDatabase),
		mongo.WithAuth(mongoUser, mongoPassword),
		mongo.WithAuthDB(mongoAuthDB),
		mongo.WithLogger(logger.Default()),
	)
	if err != nil {
		return fail("failed to start the MongoDB container", err)
	}

	env.Redis, err = redis.NewContainer(ctx,
		redis.WithNetworkName(env.Network.Name()),
		redis.WithContainerName(redisName),
		redis.WithImageName(redisImage),
		redis.WithLogger(logger.Default()),
	)
	if err != nil {
		return fail("failed to start the Redis container", err)
	}

	env.Inventory, err = startService(ctx, env.Network.Name(), "inventory", "grpc_server", inventoryAppName, inventoryGRPC, map[string]string{
		"GRPC_HOST": "0.0.0.0",
		"GRPC_PORT": inventoryGRPC,
		"HTTP_HOST": "0.0.0.0",
		"HTTP_PORT": inventoryHTTP,

		"MONGO_HOST":                 mongoName,
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
	})
	if err != nil {
		return fail("failed to start the inventory service", err)
	}

	env.Payment, err = startService(ctx, env.Network.Name(), "payment", "grpc_server", paymentAppName, paymentGRPC, map[string]string{
		"GRPC_HOST": "0.0.0.0",
		"GRPC_PORT": paymentGRPC,
		"HTTP_HOST": "0.0.0.0",
		"HTTP_PORT": paymentHTTP,
	})
	if err != nil {
		return fail("failed to start the payment service", err)
	}

	// the assembly service turns OrderPaid into ComputerAssembled after a short delay;
	// it has no server, so the container is considered started once the consumer runs
	env.Assembly, err = app.NewContainer(ctx,
		app.WithName(assemblyAppName),
		app.WithPort("8080"),
		app.WithDockerfile(path.GetProjectRoot(), dockerfile),
		app.WithBuildArg("SERVICE", "assembly"),
		app.WithBuildArg("CMD", "consumer"),
		app.WithNetwork(env.Network.Name()),
		app.WithEnv(map[string]string{
			"KAFKA_BROKERS":                 env.Kafka.Address(),
			"ORDER_PAID_TOPIC_NAME":         orderPaidTopic,
			"ORDER_PAID_CONSUMER_GROUP_ID":  "e2e-assembly-service",
			"COMPUTER_ASSEMBLED_TOPIC_NAME": computerAssembledTopic,
			"ASSEMBLY_DURATION":             assemblyDuration,
			"LOGGER_LEVEL":                  loggerLevelValue,
			"LOGGER_AS_JSON":                "true",
		}),
		app.WithLogOutput(os.Stdout),
		app.WithStartupWait(wait.ForLog("Starting the OrderPaid consumer").WithStartupTimeout(startupTimeout)),
		app.WithLogger(logger.Default()),
	)
	if err != nil {
		return fail("failed to start the assembly service", err)
	}

	// the order service reaches the other containers by their names inside the Docker network
	env.Order, err = startService(ctx, env.Network.Name(), "order", "http_server", orderAppName, orderHTTP, map[string]string{
		"HTTP_HOST": "0.0.0.0",
		"HTTP_PORT": orderHTTP,

		"INVENTORY_GRPC_HOST": inventoryAppName,
		"INVENTORY_GRPC_PORT": inventoryGRPC,
		"PAYMENT_GRPC_HOST":   paymentAppName,
		"PAYMENT_GRPC_PORT":   paymentGRPC,

		"POSTGRES_HOST":     postgresName,
		"POSTGRES_PORT":     postgresPort,
		"POSTGRES_DB":       postgresDatabase,
		"POSTGRES_USER":     postgresUser,
		"POSTGRES_PASSWORD": postgresPassword,

		"REDIS_HOST":               redisName,
		"REDIS_PORT":               redisPort,
		"REDIS_CONNECTION_TIMEOUT": redisConnectionTimeout,
		"REDIS_MAX_IDLE":           redisMaxIdle,
		"REDIS_IDLE_TIMEOUT":       redisIdleTimeout,
		"REDIS_CACHE_TTL":          redisCacheTTL,

		"KAFKA_BROKERS":                     env.Kafka.Address(),
		"ORDER_PAID_TOPIC_NAME":             orderPaidTopic,
		"COMPUTER_ASSEMBLED_TOPIC_NAME":     computerAssembledTopic,
		"ORDER_ASSEMBLED_CONSUMER_GROUP_ID": "e2e-order-service",
	})
	if err != nil {
		return fail("failed to start the order service", err)
	}

	logger.Info(ctx, "Test environment is ready")

	return env, nil
}

// startService builds a service from the shared Dockerfile and starts it in the network.
func startService(ctx context.Context, networkName, service, cmd, name, port string, appEnv map[string]string) (*app.Container, error) {
	appEnv["LOGGER_LEVEL"] = loggerLevelValue
	appEnv["LOGGER_AS_JSON"] = "true"

	return app.NewContainer(ctx,
		app.WithName(name),
		app.WithPort(port),
		app.WithDockerfile(path.GetProjectRoot(), dockerfile),
		app.WithBuildArg("SERVICE", service),
		app.WithBuildArg("CMD", cmd),
		app.WithNetwork(networkName),
		app.WithEnv(appEnv),
		app.WithLogOutput(os.Stdout),
		app.WithStartupWait(wait.ForListeningPort(port+"/tcp").WithStartupTimeout(startupTimeout)),
		app.WithLogger(logger.Default()),
	)
}

// teardownTestEnvironment releases all resources of the test environment.
func teardownTestEnvironment(ctx context.Context, env *TestEnvironment) {
	logger.Info(ctx, "Cleaning up the test environment...")

	cleanupTestEnvironment(ctx, env)

	logger.Info(ctx, "Test environment cleaned up")
}

// cleanupTestEnvironment releases the resources that were created, in the reverse order of the start.
func cleanupTestEnvironment(ctx context.Context, env *TestEnvironment) {
	for name, c := range map[string]*app.Container{
		"order":     env.Order,
		"assembly":  env.Assembly,
		"payment":   env.Payment,
		"inventory": env.Inventory,
	} {
		if c == nil {
			continue
		}

		if err := c.Terminate(ctx); err != nil {
			logger.Error(ctx, "failed to stop the service container", "service", name, "error", err)
		} else {
			logger.Info(ctx, "Service container stopped", "service", name)
		}
	}

	if env.Kafka != nil {
		if err := env.Kafka.Terminate(ctx); err != nil {
			logger.Error(ctx, "failed to stop the Kafka container", "error", err)
		}
	}

	if env.Redis != nil {
		if err := env.Redis.Terminate(ctx); err != nil {
			logger.Error(ctx, "failed to stop the Redis container", "error", err)
		}
	}

	if env.Mongo != nil {
		if err := env.Mongo.Terminate(ctx); err != nil {
			logger.Error(ctx, "failed to stop the MongoDB container", "error", err)
		}
	}

	if env.Postgres != nil {
		if err := env.Postgres.Terminate(ctx); err != nil {
			logger.Error(ctx, "failed to stop the PostgreSQL container", "error", err)
		}
	}

	if env.Network != nil {
		if err := env.Network.Remove(ctx); err != nil {
			logger.Error(ctx, "failed to remove the network", "error", err)
		} else {
			logger.Info(ctx, "Network removed")
		}
	}
}
