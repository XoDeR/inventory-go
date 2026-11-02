//go:build e2e

package e2e

import (
	"context"

	"inventory-go/platform/pkg/logger"
)

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

	if env.Mongo != nil {
		if err := env.Mongo.Terminate(ctx); err != nil {
			logger.Error(ctx, "failed to stop the MongoDB container", "error", err)
		} else {
			logger.Info(ctx, "MongoDB container stopped")
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
