package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"

	orderV1 "inventory-go/shared/pkg/openapi/order/v1"
)

const (
	serverURL      = "http://localhost:8080"
	requestTimeout = 5 * time.Second
)

// createOrder creates an order and returns its uuid and total price
func createOrder(ctx context.Context, client *orderV1.Client, req *orderV1.CreateOrderRequest) (*orderV1.CreateOrderResponse, error) {
	res, err := client.CreateOrder(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("execute create request: %w", err)
	}

	created, ok := res.(*orderV1.CreateOrderResponse)
	if !ok {
		return nil, fmt.Errorf("create order: unexpected response %T: %+v", res, res)
	}

	return created, nil
}

// payOrder pays an order and returns the transaction uuid
func payOrder(ctx context.Context, client *orderV1.Client, orderUUID uuid.UUID, method orderV1.PaymentMethod) (*orderV1.PayOrderResponse, error) {
	res, err := client.PayOrder(
		ctx,
		&orderV1.PayOrderRequest{PaymentMethod: method},
		orderV1.PayOrderParams{OrderUUID: orderUUID},
	)
	if err != nil {
		return nil, fmt.Errorf("execute pay request: %w", err)
	}

	paid, ok := res.(*orderV1.PayOrderResponse)
	if !ok {
		return nil, fmt.Errorf("pay order: unexpected response %T: %+v", res, res)
	}

	return paid, nil
}

// getOrder returns the order or nil if it was not found
func getOrder(ctx context.Context, client *orderV1.Client, orderUUID uuid.UUID) (*orderV1.Order, error) {
	res, err := client.GetOrderByUuid(ctx, orderV1.GetOrderByUuidParams{OrderUUID: orderUUID})
	if err != nil {
		return nil, fmt.Errorf("execute get request: %w", err)
	}

	switch r := res.(type) {
	case *orderV1.Order:
		return r, nil
	case *orderV1.GetOrderByUuidNotFound:
		return nil, nil
	default:
		return nil, fmt.Errorf("get order: unexpected response %T: %+v", res, res)
	}
}

// generateRandomOrderRequest builds a request with random user and parts
func generateRandomOrderRequest() *orderV1.CreateOrderRequest {
	partUuids := make([]uuid.UUID, gofakeit.IntRange(1, 5))
	for i := range partUuids {
		partUuids[i] = uuid.New()
	}

	return &orderV1.CreateOrderRequest{
		UserUUID:  uuid.New(),
		PartUuids: partUuids,
	}
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	log.Println("Testing order API")

	client, err := orderV1.NewClient(serverURL)
	if err != nil {
		log.Printf("error creating client: %v\n", err)
		return
	}

	created, err := createOrder(ctx, client, generateRandomOrderRequest())
	if err != nil {
		log.Printf("error: %v\n", err)
		return
	}
	log.Printf("created: %+v\n", created)

	paid, err := payOrder(ctx, client, created.UUID, orderV1.PaymentMethodCARD)
	if err != nil {
		log.Printf("error: %v\n", err)
		return
	}
	log.Printf("paid: %+v\n", paid)

	order, err := getOrder(ctx, client, created.UUID)
	if err != nil {
		log.Printf("error: %v\n", err)
		return
	}
	log.Printf("fetched: %+v\n", order)
}
