package main

import (
	"context"
	"errors"
	"fmt"
	"log"
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
	inventoryRepository "inventory-go/inventory/internal/repository/part"
	inventoryService "inventory-go/inventory/internal/service/part"
	"inventory-go/shared/api/swagger"
	"inventory-go/shared/pkg/interceptor"
	inventoryV1 "inventory-go/shared/pkg/proto/inventory/v1"
)

const (
	grpcPort = 50051
	httpPort = 8081

	defaultMongoURI = "mongodb://inventory-service-user:inventory-service-password@localhost:27017/?authSource=admin"
	databaseName    = "inventory-service"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = defaultMongoURI
	}

	mongoClient, err := mongo.Connect(options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Printf("failed to connect to database: %v\n", err)
		return
	}
	defer func() {
		if cerr := mongoClient.Disconnect(context.Background()); cerr != nil {
			log.Printf("failed to disconnect from database: %v\n", cerr)
		}
	}()

	pingCtx, pingCancel := context.WithTimeout(ctx, 5*time.Second)
	defer pingCancel()
	if err = mongoClient.Ping(pingCtx, nil); err != nil {
		log.Printf("database is unavailable: %v\n", err)
		return
	}

	repo := inventoryRepository.NewRepository(mongoClient.Database(databaseName))
	if err = repo.Seed(ctx); err != nil {
		log.Printf("failed to seed the inventory: %v\n", err)
		return
	}

	// listener
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		log.Printf("failed to listen: %v\n", err)
		return
	}
	defer func() {
		if cerr := lis.Close(); cerr != nil {
			log.Printf("failed to close listener: %v\n", cerr)
		}
	}()

	s := grpc.NewServer(grpc.UnaryInterceptor(interceptor.UnaryValidate()))

	service := inventoryService.NewService(repo)
	api := inventoryV1API.NewAPI(service)

	inventoryV1.RegisterInventoryServiceServer(s, api)

	reflection.Register(s)

	go func() {
		log.Printf("gRPC server listening on %d\n", grpcPort)
		err = s.Serve(lis)
		if err != nil {
			log.Printf("failed to serve: %v\n", err)
			return
		}
	}()

	// HTTP gateway (REST -> gRPC) + Swagger UI
	gwMux := runtime.NewServeMux()
	err = inventoryV1.RegisterInventoryServiceHandlerFromEndpoint(
		ctx,
		gwMux,
		fmt.Sprintf("localhost:%d", grpcPort),
		[]grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())},
	)
	if err != nil {
		log.Printf("failed to register gateway: %v\n", err)
		return
	}

	httpMux := http.NewServeMux()
	httpMux.Handle("/api/", gwMux)
	httpMux.Handle("/", swagger.Handler("inventory/v1/inventory.swagger.json"))

	gwServer := &http.Server{
		Addr:              fmt.Sprintf(":%d", httpPort),
		Handler:           httpMux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("HTTP gateway listening on %d (Swagger UI: /swagger-ui.html)\n", httpPort)
		if serr := gwServer.ListenAndServe(); serr != nil && !errors.Is(serr, http.ErrServerClosed) {
			log.Printf("failed to serve HTTP: %v\n", serr)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down servers...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := gwServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v\n", err)
	}

	s.GracefulStop()
	log.Println("Server stopped")
}
