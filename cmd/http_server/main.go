package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"

	orderV1 "inventory-go/pkg/openapi/order/v1"
)

const (
	httpPort          = "8080"
	urlParamOrder     = "order"
	urlParamUuid      = "uuid"
	readHeaderTimeout = 5 * time.Second
	shutdownTimeout   = 10 * time.Second
)

type OrderStorage struct {
	mu     sync.RWMutex
	orders map[string]*orderV1.Order
}

func NewOrderStorage() *OrderStorage {
	return &OrderStorage{
		orders: make(map[string]*orderV1.Order),
	}
}

func (s *OrderStorage) CreateOrder(order orderV1.Order) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.orders[order.OrderUUID.String()] = &order
}

func (s *OrderStorage) UpdateOrder(order orderV1.Order) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.orders[order.OrderUUID.String()] = &order
}

func (s *OrderStorage) GetOrder(uuid string) *orderV1.Order {
	s.mu.RLock()
	defer s.mu.RUnlock()

	order, ok := s.orders[uuid]
	if !ok {
		return nil
	}
	return order
}

// Implements orderV1.Handler interface
type OrderHandler struct {
	storage *OrderStorage
}

func NewOrderHandler(storage *OrderStorage) *OrderHandler {
	return &OrderHandler{
		storage: storage,
	}
}

// POST /api/v1/orders/{order_uuid}/cancel
func (h *OrderHandler) CancelOrder(_ context.Context, params orderV1.CancelOrderParams) (orderV1.CancelOrderRes, error) {
	order := h.storage.GetOrder(params.OrderUUID.String())
	if order == nil {
		return &orderV1.CancelOrderNotFound{
			Code:    http.StatusNotFound,
			Message: "order not found",
		}, nil
	}

	if order.Status == orderV1.OrderStatusPAID {
		return &orderV1.CancelOrderConflict{
			Code:    http.StatusConflict,
			Message: "order is already paid and cannot be cancelled",
		}, nil
	}

	// copy to avoid mutating the stored order outside of the storage lock
	updated := *order
	updated.Status = orderV1.OrderStatusCANCELLED
	h.storage.UpdateOrder(updated)

	return &orderV1.CancelOrderNoContent{}, nil
}

// POST /api/v1/orders
func (h *OrderHandler) CreateOrder(_ context.Context, req *orderV1.CreateOrderRequest) (orderV1.CreateOrderRes, error) {
	if len(req.PartUuids) == 0 {
		return &orderV1.CreateOrderBadRequest{
			Code:    http.StatusBadRequest,
			Message: "part_uuids must not be empty",
		}, nil
	}

	// TODO: fetch parts via InventoryService.ListParts, return an error if any is missing
	// and sum their real prices. Until the inventory client exists, prices are generated.
	var totalPrice float64
	for range req.PartUuids {
		totalPrice += gofakeit.Price(100, 1000)
	}

	order := orderV1.Order{
		OrderUUID:  uuid.New(),
		UserUUID:   req.UserUUID,
		PartUuids:  req.PartUuids,
		TotalPrice: totalPrice,
		Status:     orderV1.OrderStatusPENDINGPAYMENT,
	}
	h.storage.CreateOrder(order)

	return &orderV1.CreateOrderResponse{
		OrderUUID:  order.OrderUUID,
		TotalPrice: order.TotalPrice,
	}, nil
}

// GET /api/v1/orders/{order_uuid}
func (h *OrderHandler) GetOrderByUuid(_ context.Context, params orderV1.GetOrderByUuidParams) (orderV1.GetOrderByUuidRes, error) {
	order := h.storage.GetOrder(params.OrderUUID.String())
	if order == nil {
		return &orderV1.GetOrderByUuidNotFound{
			Code:    http.StatusNotFound,
			Message: "order not found",
		}, nil
	}

	return order, nil
}

// POST /api/v1/orders/{order_uuid}/pay
func (h *OrderHandler) PayOrder(_ context.Context, req *orderV1.PayOrderRequest, params orderV1.PayOrderParams) (orderV1.PayOrderRes, error) {
	order := h.storage.GetOrder(params.OrderUUID.String())
	if order == nil {
		return &orderV1.PayOrderNotFound{
			Code:    http.StatusNotFound,
			Message: "order not found",
		}, nil
	}

	// TODO: call PaymentService.PayOrder with user_uuid, order_uuid and payment_method.
	// Until the payment client exists, the transaction uuid is generated locally.
	transactionUUID := uuid.New()

	updated := *order
	updated.Status = orderV1.OrderStatusPAID
	updated.TransactionUUID = orderV1.NewOptNilUUID(transactionUUID)
	updated.PaymentMethod = orderV1.NewOptNilPaymentMethod(req.PaymentMethod)
	h.storage.UpdateOrder(updated)

	return &orderV1.PayOrderResponse{
		TransactionUUID: transactionUUID,
	}, nil
}

// Used for common default response.
func (h *OrderHandler) NewError(_ context.Context, err error) *orderV1.UnexpectedErrorStatusCode {
	return &orderV1.UnexpectedErrorStatusCode{
		StatusCode: http.StatusInternalServerError,
		Response: orderV1.Error{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		},
	}
}

func main() {
	storage := NewOrderStorage()

	orderHandler := NewOrderHandler(storage)

	orderServer, err := orderV1.NewServer(orderHandler)
	if err != nil {
		log.Fatalf("error creating OpenApi server: %v", err)
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
