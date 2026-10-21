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
	"slices"
	"sync"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	"inventory-go/shared/api/swagger"
	"inventory-go/shared/pkg/interceptor"
	inventoryV1 "inventory-go/shared/pkg/proto/inventory/v1"
)

const (
	grpcPort = 50051
	httpPort = 8081
)

type inventoryService struct {
	inventoryV1.UnimplementedInventoryServiceServer
	mu    sync.RWMutex
	parts map[string]*inventoryV1.Part
}

func (s *inventoryService) GetPart(_ context.Context, req *inventoryV1.GetPartRequest) (*inventoryV1.GetPartResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	part, ok := s.parts[req.GetUuid()]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "part with UUID %s not found", req.GetUuid())
	}

	return &inventoryV1.GetPartResponse{Part: part}, nil
}

// ListParts returns parts matching the filter: OR within a single filter
// field, AND between different fields. An empty field is not applied.
func (s *inventoryService) ListParts(_ context.Context, req *inventoryV1.ListPartsRequest) (*inventoryV1.ListPartsResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	filter := req.GetFilter()
	parts := make([]*inventoryV1.Part, 0, len(s.parts))
	for _, part := range s.parts {
		if matchesFilter(part, filter) {
			parts = append(parts, part)
		}
	}

	return &inventoryV1.ListPartsResponse{Parts: parts}, nil
}

func matchesFilter(part *inventoryV1.Part, filter *inventoryV1.PartsFilter) bool {
	if filter == nil {
		return true
	}

	if len(filter.GetUuids()) > 0 && !slices.Contains(filter.GetUuids(), part.GetUuid()) {
		return false
	}
	if len(filter.GetNames()) > 0 && !slices.Contains(filter.GetNames(), part.GetName()) {
		return false
	}
	if len(filter.GetCategories()) > 0 && !slices.Contains(filter.GetCategories(), part.GetCategory()) {
		return false
	}
	if len(filter.GetManufacturerCountries()) > 0 &&
		!slices.Contains(filter.GetManufacturerCountries(), part.GetManufacturer().GetCountry()) {
		return false
	}
	if len(filter.GetTags()) > 0 && !slices.ContainsFunc(part.GetTags(), func(tag string) bool {
		return slices.Contains(filter.GetTags(), tag)
	}) {
		return false
	}

	return true
}

func main() {
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

	service := &inventoryService{
		parts: make(map[string]*inventoryV1.Part),
	}

	for _, part := range seedParts() {
		service.parts[part.GetUuid()] = part
	}

	inventoryV1.RegisterInventoryServiceServer(s, service)

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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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
