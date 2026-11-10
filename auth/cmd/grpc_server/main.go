package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"
	"time"

	"inventory-go/auth/internal/app"
	"inventory-go/auth/internal/config"
	"inventory-go/platform/pkg/closer"
	"inventory-go/platform/pkg/logger"
)

const (
	configPath = "./deploy/compose/auth/.env"

	shutdownTimeout = 5 * time.Second
)

func main() {
	err := config.Load(configPath)
	if err != nil {
		slog.Error("failed to load config", "error", err)
		return
	}

	appCtx, appCancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer appCancel()
	defer gracefulShutdown()

	closer.Configure(syscall.SIGINT, syscall.SIGTERM)

	a, err := app.New(appCtx)
	if err != nil {
		logger.Error(appCtx, "failed to create the app", "error", err)
		return
	}

	err = a.Run(appCtx)
	if err != nil {
		logger.Error(appCtx, "the app failed", "error", err)
		return
	}
}

func gracefulShutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := closer.CloseAll(ctx); err != nil {
		logger.Error(ctx, "failed to shut down gracefully", "error", err)
	}
}
