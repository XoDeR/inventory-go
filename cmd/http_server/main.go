package main

import (
	"inventory-go/pkg/models"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

const (
	httpPort          = "8080"
	readHeaderTimeout = 5 * time.Second
	shutdownTimeout   = 10 * time.Second
)

func main() {

	// router
	r := chi.NewRouter()

	// middleware

	// routes
}

func createOrderHandler(storage *models.OrderStorage) http.HandlerFunc {

}

func updateOrderHandler(storage *models.OrderStorage) http.HandlerFunc {

}

func getOrderHandler(storage *models.OrderStorage) http.HandlerFunc {

}
