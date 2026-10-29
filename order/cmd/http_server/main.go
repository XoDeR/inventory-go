package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	orderV1API "inventory-go/order/internal/api/order/v1"
	inventoryClientV1 "inventory-go/order/internal/client/grpc/inventory/v1"
	paymentClientV1 "inventory-go/order/internal/client/grpc/payment/v1"
	"inventory-go/order/internal/config"
	"inventory-go/order/internal/migrator"
	orderRepository "inventory-go/order/internal/repository/order"
	orderService "inventory-go/order/internal/service/order"
	"inventory-go/order/migrations"
	"inventory-go/shared/pkg/logger"
	orderV1 "inventory-go/shared/pkg/openapi/order/v1"
	inventoryV1 "inventory-go/shared/pkg/proto/inventory/v1"
	paymentV1 "inventory-go/shared/pkg/proto/payment/v1"
)

const (
	configPath = "./deploy/compose/order/.env"

	readHeaderTimeout = 5 * time.Second
	shutdownTimeout   = 10 * time.Second
)

func main() {
	ctx := context.Background()

	err := config.Load(configPath)
	if err != nil {
		slog.Error("failed to load config", "error", err)
		return
	}

	err = logger.Init(config.AppConfig().Logger.Level(), config.AppConfig().Logger.AsJSON())
	if err != nil {
		slog.Error("failed to init logger", "error", err)
		return
	}

	pool, err := pgxpool.New(ctx, config.AppConfig().Postgres.URI())
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		return
	}
	defer pool.Close()

	if err = pool.Ping(ctx); err != nil {
		slog.Error("database is unavailable", "error", err)
		return
	}

	// goose works on top of database/sql, so run migrations through a stdlib wrapper of the pool
	migrationsDB := stdlib.OpenDBFromPool(pool)
	err = migrator.NewMigrator(migrationsDB).Up(ctx, migrations.FS)
	if err != nil {
		slog.Error("failed to run migrations", "error", err)
		return
	}
	if err = migrationsDB.Close(); err != nil {
		slog.Error("failed to close migrations connection", "error", err)
	}

	inventoryConn, err := grpc.NewClient(
		config.AppConfig().Inventory.Address(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		slog.Error("failed to connect to inventory", "error", err)
		return
	}
	defer func() {
		if cerr := inventoryConn.Close(); cerr != nil {
			slog.Error("failed to close inventory connection", "error", cerr)
		}
	}()

	paymentConn, err := grpc.NewClient(
		config.AppConfig().Payment.Address(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		slog.Error("failed to connect to payment", "error", err)
		return
	}
	defer func() {
		if cerr := paymentConn.Close(); cerr != nil {
			slog.Error("failed to close payment connection", "error", cerr)
		}
	}()

	inventoryClient := inventoryClientV1.NewClient(inventoryV1.NewInventoryServiceClient(inventoryConn))
	paymentClient := paymentClientV1.NewClient(paymentV1.NewPaymentServiceClient(paymentConn))

	repo := orderRepository.NewRepository(pool)
	service := orderService.NewService(repo, inventoryClient, paymentClient)
	api := orderV1API.NewAPI(service)

	orderServer, err := orderV1.NewServer(api)
	if err != nil {
		slog.Error("error creating OpenApi server", "error", err)
		return
	}

	// router
	r := chi.NewRouter()

	// middleware
	r.Use(middleware.Logger) // log like: PUT http://localhost:8080/api/v1/order/3c516345-47df-4eba-a9d6-da1e3d80860b HTTP/1.1" from 127.0.0.
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(10 * time.Second))

	r.Mount("/", orderServer)

	// create server
	server := &http.Server{
		Addr:              config.AppConfig().HTTP.Address(),
		Handler:           r,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	// launch server in go routine
	go func() {
		slog.Info("HTTP server running", "address", config.AppConfig().HTTP.Address())
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Error launching server", "error", err)
		}
	}()

	// graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server")

	ctx, cancel := context.WithTimeout(ctx, shutdownTimeout)
	defer cancel()

	err = server.Shutdown(ctx)
	if err != nil {
		slog.Error("Server shutdown error", "error", err)
	}

	slog.Info("server stopped")
}
