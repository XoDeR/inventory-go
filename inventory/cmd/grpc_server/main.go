package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	inventoryV1API "inventory-go/inventory/internal/api/inventory/v1"
	"inventory-go/inventory/internal/config"
	inventoryRepository "inventory-go/inventory/internal/repository/part"
	inventoryService "inventory-go/inventory/internal/service/part"
	"inventory-go/shared/api/swagger"
	"inventory-go/shared/pkg/interceptor"
	"inventory-go/shared/pkg/logger"
	inventoryV1 "inventory-go/shared/pkg/proto/inventory/v1"
)

const configPath = "./deploy/compose/inventory/.env"

func main() {
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mongoClient, err := mongo.Connect(options.Client().ApplyURI(config.AppConfig().Mongo.URI()))
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		return
	}
	defer func() {
		if cerr := mongoClient.Disconnect(context.Background()); cerr != nil {
			slog.Error("failed to disconnect from database", "error", cerr)
		}
	}()

	pingCtx, pingCancel := context.WithTimeout(ctx, 5*time.Second)
	defer pingCancel()
	if err = mongoClient.Ping(pingCtx, nil); err != nil {
		slog.Error("database is unavailable", "error", err)
		return
	}

	repo := inventoryRepository.NewRepository(mongoClient.Database(config.AppConfig().Mongo.DatabaseName()))
	if err = repo.Seed(ctx); err != nil {
		slog.Error("failed to seed the inventory", "error", err)
		return
	}

	// listener
	lis, err := net.Listen("tcp", config.AppConfig().GRPC.Address())
	if err != nil {
		slog.Error("failed to listen", "error", err)
		return
	}
	defer func() {
		if cerr := lis.Close(); cerr != nil {
			slog.Error("failed to close listener", "error", cerr)
		}
	}()

	s := grpc.NewServer(grpc.UnaryInterceptor(interceptor.UnaryValidate()))

	service := inventoryService.NewService(repo)
	api := inventoryV1API.NewAPI(service)

	inventoryV1.RegisterInventoryServiceServer(s, api)

	reflection.Register(s)

	go func() {
		slog.Info("gRPC server listening", "address", config.AppConfig().GRPC.Address())
		err = s.Serve(lis)
		if err != nil {
			slog.Error("failed to serve", "error", err)
			return
		}
	}()

	// HTTP gateway (REST -> gRPC) + Swagger UI
	gwMux := runtime.NewServeMux()
	err = inventoryV1.RegisterInventoryServiceHandlerFromEndpoint(
		ctx,
		gwMux,
		config.AppConfig().GRPC.Address(),
		[]grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())},
	)
	if err != nil {
		slog.Error("failed to register gateway", "error", err)
		return
	}

	httpMux := http.NewServeMux()
	httpMux.Handle("/api/", gwMux)
	httpMux.Handle("/", swagger.Handler("inventory/v1/inventory.swagger.json"))

	gwServer := &http.Server{
		Addr:              config.AppConfig().HTTP.Address(),
		Handler:           httpMux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		slog.Info("HTTP gateway listening", "address", config.AppConfig().HTTP.Address(), "swagger_ui", "/swagger-ui.html")
		if serr := gwServer.ListenAndServe(); serr != nil && !errors.Is(serr, http.ErrServerClosed) {
			slog.Error("failed to serve HTTP", "error", serr)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("shutting down servers")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := gwServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("HTTP server shutdown error", "error", err)
	}

	s.GracefulStop()
	slog.Info("server stopped")
}
