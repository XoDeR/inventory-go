package main

import (
	"context"
	inventoryV1 "inventory-go/shared/pkg/proto/inventory/v1"
	"slices"
	"sync"

	"google.golang.org/grpc/codes"
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
