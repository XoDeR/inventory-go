package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	inventoryV1 "inventory-go/shared/pkg/proto/inventory/v1"
)

const (
	serverAddress  = "localhost:50051"
	requestTimeout = 5 * time.Second
)

func main() {
	conn, err := grpc.NewClient(
		serverAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		slog.Error("failed to connect", "error", err)
		return
	}
	defer func() {
		if cerr := conn.Close(); cerr != nil {
			slog.Error("failed to close connection", "error", cerr)
		}
	}()

	client := inventoryV1.NewInventoryServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	parts, err := listParts(ctx, client, nil)
	if err != nil {
		slog.Error("failed to list parts", "error", err)
		return
	}
	slog.Info("ListParts", "parts", len(parts))
	for _, part := range parts {
		fmt.Printf("  %s  %s  %s\n", part.GetUuid(), part.GetName(), part.GetCategory()) //nolint:forbidigo // CLI client prints its result to stdout
	}

	// CPU or GPU
	filtered, err := listParts(ctx, client, &inventoryV1.PartsFilter{
		Categories: []inventoryV1.Category{
			inventoryV1.Category_CATEGORY_CPU,
			inventoryV1.Category_CATEGORY_GPU,
		},
	})
	if err != nil {
		slog.Error("failed to list filtered parts", "error", err)
		return
	}
	slog.Info("ListParts (CPU/GPU)", "parts", len(filtered))

	// The first listed part, or an unknown UUID to show NotFound
	uuid := "00000000-0000-0000-0000-000000000000"
	if len(parts) > 0 {
		uuid = parts[0].GetUuid()
	}

	part, err := getPart(ctx, client, uuid)
	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
			slog.Warn("GetPart: part not found", "uuid", uuid)
			return
		}
		slog.Error("failed to get part", "error", err)
		return
	}
	slog.Info("GetPart", "part", part)
}

func listParts(ctx context.Context, client inventoryV1.InventoryServiceClient, filter *inventoryV1.PartsFilter) ([]*inventoryV1.Part, error) {
	resp, err := client.ListParts(ctx, &inventoryV1.ListPartsRequest{Filter: filter})
	if err != nil {
		return nil, err
	}

	return resp.GetParts(), nil
}

func getPart(ctx context.Context, client inventoryV1.InventoryServiceClient, uuid string) (*inventoryV1.Part, error) {
	resp, err := client.GetPart(ctx, &inventoryV1.GetPartRequest{Uuid: uuid})
	if err != nil {
		return nil, err
	}

	return resp.GetPart(), nil
}
