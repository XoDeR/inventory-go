package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	orderV1API "inventory-go/order/internal/api/order/v1"
	inventoryClientV1 "inventory-go/order/internal/client/grpc/inventory/v1"
	paymentClientV1 "inventory-go/order/internal/client/grpc/payment/v1"
	orderRepository "inventory-go/order/internal/repository/order"
	orderService "inventory-go/order/internal/service/order"
	orderV1 "inventory-go/shared/pkg/openapi/order/v1"
	inventoryV1 "inventory-go/shared/pkg/proto/inventory/v1"
	paymentV1 "inventory-go/shared/pkg/proto/payment/v1"
)

const (
	inventoryServerAddress = "localhost:50051"
	paymentServerAddress   = "localhost:50052"

	httpPort          = "8080"
	readHeaderTimeout = 5 * time.Second
	shutdownTimeout   = 10 * time.Second
)

func main() {
	inventoryConn, err := grpc.NewClient(
		inventoryServerAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Printf("failed to connect to inventory: %v\n", err)
		return
	}
	defer func() {
		if cerr := inventoryConn.Close(); cerr != nil {
			log.Printf("failed to close inventory connection: %v\n", cerr)
		}
	}()

	paymentConn, err := grpc.NewClient(
		paymentServerAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Printf("failed to connect to payment: %v\n", err)
		return
	}
	defer func() {
		if cerr := paymentConn.Close(); cerr != nil {
			log.Printf("failed to close payment connection: %v\n", cerr)
		}
	}()

	inventoryClient := inventoryClientV1.NewClient(inventoryV1.NewInventoryServiceClient(inventoryConn))
	paymentClient := paymentClientV1.NewClient(paymentV1.NewPaymentServiceClient(paymentConn))

	repo := orderRepository.NewRepository()
	service := orderService.NewService(repo, inventoryClient, paymentClient)
	api := orderV1API.NewAPI(service)

	orderServer, err := orderV1.NewServer(api)
	if err != nil {
		log.Printf("error creating OpenApi server: %v\n", err)
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
		Addr:              net.JoinHostPort("localhost", httpPort),
		Handler:           r,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	// launch server in go routine
	go func() {
		log.Printf("HTTP server running on port %s\n", httpPort)
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("Error launching server: %v\n", err)
		}
	}()

	// graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Server shutdown...")

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	err = server.Shutdown(ctx)
	if err != nil {
		log.Printf("Server shutdown error: %v\n", err)
	}

	log.Println("Server stopped.")
}
