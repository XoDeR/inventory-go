//go:build e2e

package e2e

import "time"

const (
	// projectName is the name of the project for the Docker containers and the network
	projectName = "inventory-service-e2e"

	// partsCollectionName is the name of the MongoDB collection with the parts
	partsCollectionName = "parts"

	// Container parameters
	appName       = "e2e-inventory-app"
	mongoName     = "e2e-inventory-mongo"
	dockerfile    = "deploy/docker/Dockerfile"
	serviceBuild  = "inventory"
	cmdBuild      = "grpc_server"
	mongoImage    = "mongo:7.0.5"
	mongoDatabase = "inventory-service"
	mongoUser     = "inventory-service-user"
	mongoPassword = "inventory-service-password"
	mongoAuthDB   = "admin"
	mongoPort     = "27017"

	redisName              = "e2e-inventory-redis"
	redisImage             = "redis:7.2.5-alpine3.20"
	redisPort              = "6379"
	redisConnectionTimeout = "5s"
	redisMaxIdle           = "10"
	redisIdleTimeout       = "240s"
	redisCacheTTL          = "1h"

	// ports the app listens on inside its container
	grpcPort = "50051"
	httpPort = "8081"

	loggerLevelValue = "debug"
	startupTimeout   = 3 * time.Minute
	testsTimeout     = 10 * time.Minute
)
