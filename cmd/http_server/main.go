package main

import (
	"context"
	"errors"
	"inventory-go/pkg/models"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

const (
	httpPort          = "8080"
	urlParamOrder     = "order"
	urlParamUuid      = "uuid"
	readHeaderTimeout = 5 * time.Second
	shutdownTimeout   = 10 * time.Second
)

func main() {
	storage := models.NewOrderStorage()

	// router
	r := chi.NewRouter()

	// middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(10 * time.Second))
	r.Use(render.SetContentType(render.ContentTypeJSON))

	// routes
	r.Route("/api/order", func(r chi.Router) {
		r.Post("/", createOrderHandler(storage))
		r.Get("/{uuid}", getOrderHandler(storage))
		r.Put("/{uuid}", updateOrderHandler(storage))
	})

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

	err := server.Shutdown(ctx)
	if err != nil {
		log.Printf("Server shutdown error: %v\n", err)
	}

	log.Println("Server stopped.")
}

func createOrderHandler(storage *models.OrderStorage) http.HandlerFunc {

}

func updateOrderHandler(storage *models.OrderStorage) http.HandlerFunc {

}

func getOrderHandler(storage *models.OrderStorage) http.HandlerFunc {

}
