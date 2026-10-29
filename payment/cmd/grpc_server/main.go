package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	paymentV1API "inventory-go/payment/internal/api/payment/v1"
	"inventory-go/payment/internal/config"
	paymentService "inventory-go/payment/internal/service/payment"
	"inventory-go/shared/api/swagger"
	"inventory-go/shared/pkg/interceptor"
	"inventory-go/shared/pkg/logger"
	paymentV1 "inventory-go/shared/pkg/proto/payment/v1"
)

const configPath = "./deploy/compose/payment/.env"

func main() {
	err := config.Load(configPath)
	if err != nil {
		slog.Error("failed to load config", "error", err)
		return
	}

	err = logger.Init(config.AppConfig().Logger.Level(), config.AppConfig().Logger.AsJSON())
	if err != nil {
		slog.Error("failed to init logger", "error", err)
		return
	}

	// listener
	lis, err := net.Listen("tcp", config.AppConfig().GRPC.Address())
	if err != nil {
		slog.Error("failed to listen", "error", err)
		return
	}
	defer func() {
		if cerr := lis.Close(); cerr != nil {
			slog.Error("failed to close listener", "error", cerr)
		}
	}()

	s := grpc.NewServer(grpc.UnaryInterceptor(interceptor.UnaryValidate()))

	service := paymentService.NewService()
	api := paymentV1API.NewAPI(service)

	paymentV1.RegisterPaymentServiceServer(s, api)

	reflection.Register(s)

	go func() {
		slog.Info("gRPC server listening", "address", config.AppConfig().GRPC.Address())
		err = s.Serve(lis)
		if err != nil {
			slog.Error("failed to serve", "error", err)
			return
		}
	}()

	// HTTP gateway (REST -> gRPC) + Swagger UI
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	gwMux := runtime.NewServeMux()
	err = paymentV1.RegisterPaymentServiceHandlerFromEndpoint(
		ctx,
		gwMux,
		config.AppConfig().GRPC.Address(),
		[]grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())},
	)
	if err != nil {
		slog.Error("failed to register gateway", "error", err)
		return
	}

	httpMux := http.NewServeMux()
	httpMux.Handle("/api/", gwMux)
	httpMux.Handle("/", swagger.Handler("payment/v1/payment.swagger.json"))

	gwServer := &http.Server{
		Addr:              config.AppConfig().HTTP.Address(),
		Handler:           httpMux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		slog.Info("HTTP gateway listening", "address", config.AppConfig().HTTP.Address(), "swagger_ui", "/swagger-ui.html")
		if serr := gwServer.ListenAndServe(); serr != nil && !errors.Is(serr, http.ErrServerClosed) {
			slog.Error("failed to serve HTTP", "error", serr)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("shutting down servers")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := gwServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("HTTP server shutdown error", "error", err)
	}

	s.GracefulStop()
	slog.Info("server stopped")
}
