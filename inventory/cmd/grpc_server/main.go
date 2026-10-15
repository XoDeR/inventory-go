package main

import (
	"context"
	"fmt"
	inventoryV1 "inventory-go/shared/pkg/proto/inventory/v1"
	"log"
	"net"
	"os"
	"os/signal"
	"slices"
	"sync"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

const grpcPort = 50051

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

	s := grpc.NewServer()

	service := &inventoryService{
		parts: make(map[string]*inventoryV1.Part),
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

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down gRPC server...")
	s.GracefulStop()
	log.Println("Server stopped")
}
