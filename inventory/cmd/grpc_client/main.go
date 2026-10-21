package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	inventoryV1 "inventory-go/shared/pkg/proto/inventory/v1"
)

const serverAddress = "localhost:50051"

func main() {
	conn, err := grpc.NewClient(
		serverAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Printf("failed to connect: %v\n", err)
		return
	}
	defer func() {
		if cerr := conn.Close(); cerr != nil {
			log.Printf("failed to close connection: %v\n", cerr)
		}
	}()

	client := inventoryV1.NewInventoryServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// List all parts (empty filter)
	listResp, err := client.ListParts(ctx, &inventoryV1.ListPartsRequest{})
	if err != nil {
		log.Printf("failed to list parts: %v\n", err)
		return
	}
	log.Printf("ListParts: %d part(s)\n", len(listResp.GetParts()))
	for _, part := range listResp.GetParts() {
		fmt.Printf("  %s  %s  %s\n", part.GetUuid(), part.GetName(), part.GetCategory())
	}

	// List parts with a filter: CPU or GPU
	filteredResp, err := client.ListParts(ctx, &inventoryV1.ListPartsRequest{
		Filter: &inventoryV1.PartsFilter{
			Categories: []inventoryV1.Category{
				inventoryV1.Category_CATEGORY_CPU,
				inventoryV1.Category_CATEGORY_GPU,
			},
		},
	})
	if err != nil {
		log.Printf("failed to list filtered parts: %v\n", err)
		return
	}
	log.Printf("ListParts (CPU/GPU): %d part(s)\n", len(filteredResp.GetParts()))

	// Get one part: the first listed one, or an unknown UUID to show NotFound
	uuid := "00000000-0000-0000-0000-000000000000"
	if len(listResp.GetParts()) > 0 {
		uuid = listResp.GetParts()[0].GetUuid()
	}

	getResp, err := client.GetPart(ctx, &inventoryV1.GetPartRequest{Uuid: uuid})
	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
			log.Printf("GetPart: part %s not found\n", uuid)
			return
		}
		log.Printf("failed to get part: %v\n", err)
		return
	}
	log.Printf("GetPart: %v\n", getResp.GetPart())
}
