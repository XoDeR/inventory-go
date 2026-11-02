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
	"inventory-go/platform/pkg/testcontainers/network"
	"inventory-go/platform/pkg/testcontainers/path"
)

const (
	// projectName is the name of the project for the Docker containers and the network
	projectName = "payment-service-e2e"

	appName      = "e2e-payment-app"
	dockerfile   = "deploy/docker/Dockerfile"
	serviceBuild = "payment"
	cmdBuild     = "grpc_server"

	// ports the app listens on inside its container
	grpcPort = "50052"
	httpPort = "8082"

	loggerLevelValue = "debug"
	startupTimeout   = 3 * time.Minute
	testsTimeout     = 10 * time.Minute
)

// TestEnvironment holds the resources of the test environment.
type TestEnvironment struct {
	Network *network.Network
	App     *app.Container
}

// setupTestEnvironment prepares the test environment: the network and the payment service.
func setupTestEnvironment(ctx context.Context) (*TestEnvironment, error) {
	logger.Info(ctx, "Preparing the test environment...")

	generatedNetwork, err := network.NewNetwork(ctx, projectName)
	if err != nil {
		return nil, fmt.Errorf("failed to create the network: %w", err)
	}
	logger.Info(ctx, "Network created")

	appEnv := map[string]string{
		"GRPC_HOST": "0.0.0.0",
		"GRPC_PORT": grpcPort,
		"HTTP_HOST": "0.0.0.0",
		"HTTP_PORT": httpPort,

		"LOGGER_LEVEL":   loggerLevelValue,
		"LOGGER_AS_JSON": "true",
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
		cleanupTestEnvironment(ctx, &TestEnvironment{Network: generatedNetwork})
		return nil, fmt.Errorf("failed to start the app container: %w", err)
	}
	logger.Info(ctx, "App container started")

	return &TestEnvironment{Network: generatedNetwork, App: appContainer}, nil
}

// teardownTestEnvironment releases all resources of the test environment.
func teardownTestEnvironment(ctx context.Context, env *TestEnvironment) {
	logger.Info(ctx, "Cleaning up the test environment...")

	cleanupTestEnvironment(ctx, env)

	logger.Info(ctx, "Test environment cleaned up")
}

// cleanupTestEnvironment releases the resources that were created.
func cleanupTestEnvironment(ctx context.Context, env *TestEnvironment) {
	if env.App != nil {
		if err := env.App.Terminate(ctx); err != nil {
			logger.Error(ctx, "failed to stop the app container", "error", err)
		} else {
			logger.Info(ctx, "App container stopped")
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
