package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"inventory-go/pkg/models"
	"io"
	"log"
	"math/rand/v2"
	"net/http"
	"time"

	"github.com/google/uuid"
)

const (
	serverURL      = "http://localhost:8080"
	requestTimeout = 5 * time.Second
	orderAPIPath   = "/api/v1/order/%s"
)

var (
	orderStatuses  = []string{"new", "pending_payment", "paid", "cancelled"}
	paymentMethods = []string{"card", "sbp", "credit_card", "investor_money"}
)

var httpClient = &http.Client{
	Timeout: requestTimeout,
}

func createOrder(ctx context.Context, order *models.Order) (*models.Order, error) {
	return sendOrder(ctx, http.MethodPost, "", order, http.StatusCreated)
}

func updateOrder(ctx context.Context, uuid string, order *models.Order) (*models.Order, error) {
	return sendOrder(ctx, http.MethodPut, uuid, order, http.StatusOK)
}

// sendOrder sends order as JSON and decodes the order from the response
func sendOrder(ctx context.Context, method, uuid string, order *models.Order, wantStatus int) (*models.Order, error) {
	payload, err := json.Marshal(order)
	if err != nil {
		return nil, fmt.Errorf("encode json: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		method,
		fmt.Sprintf("%s"+orderAPIPath, serverURL, uuid),
		bytes.NewReader(payload),
	)
	if err != nil {
		return nil, fmt.Errorf("create %s request: %w", method, err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute %s request: %w", method, err)
	}

	defer func() {
		cerr := resp.Body.Close()
		if cerr != nil {
			log.Printf("error closing request body: %v\n", cerr)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode != wantStatus {
		return nil, fmt.Errorf("%s order (status %d): %s", method, resp.StatusCode, string(body))
	}

	var result models.Order
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, fmt.Errorf("decode json: %w", err)
	}

	return &result, nil
}

func getOrder(ctx context.Context, uuid string) (*models.Order, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s"+orderAPIPath, serverURL, uuid),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("create get request: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute get request: %w", err)
	}

	defer func() {
		cerr := resp.Body.Close()
		if cerr != nil {
			log.Printf("error closing request body: %v\n", cerr)
			return
		}
	}()

	// if 404 return nil
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read request body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get order data (status %d): %s", resp.StatusCode, string(body))
	}

	var order models.Order
	err = json.Unmarshal(body, &order)
	if err != nil {
		return nil, fmt.Errorf("decode json: %w", err)
	}

	return &order, nil
}

// generateRandomOrder builds an order with random user, parts, price,
// status and payment data. Uuid and timestamps are assigned by the server.
func generateRandomOrder() *models.Order {
	partUuids := make([]string, 1+rand.IntN(5))
	for i := range partUuids {
		partUuids[i] = uuid.NewString()
	}

	order := &models.Order{
		UserUuid:   uuid.NewString(),
		PartUuids:  partUuids,
		TotalPrice: 100 + rand.IntN(100_000),
		Status:     orderStatuses[rand.IntN(len(orderStatuses))],
	}

	// half of the orders are already paid
	if rand.IntN(2) == 0 {
		transactionUuid := uuid.NewString()
		paymentMethod := paymentMethods[rand.IntN(len(paymentMethods))]
		order.TransactionUuid = &transactionUuid
		order.PaymentMethod = &paymentMethod
	}

	return order
}

func main() {
	ctx := context.Background()

	log.Println("Testing order API")

	created, err := createOrder(ctx, generateRandomOrder())
	if err != nil {
		log.Printf("error: %v\n", err)
		return
	}
	log.Printf("created: %+v\n", created)

	created.Status = "paid"
	updated, err := updateOrder(ctx, created.Uuid, created)
	if err != nil {
		log.Printf("error: %v\n", err)
		return
	}
	log.Printf("updated: %+v\n", updated)

	order, err := getOrder(ctx, created.Uuid)
	if err != nil {
		log.Printf("error: %v\n", err)
		return
	}
	log.Printf("fetched: %+v\n", order)
}
