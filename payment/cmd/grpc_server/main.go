package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	"inventory-go/shared/api/swagger"
	"inventory-go/shared/pkg/interceptor"
	paymentV1 "inventory-go/shared/pkg/proto/payment/v1"
)

const (
	grpcPort = 50052
	httpPort = 8082
)

type paymentService struct {
	paymentV1.UnimplementedPaymentServiceServer
}

// PayOrder simulates a payment: it validates the request, logs it and
// returns a new transaction UUID. No state is stored.
func (s *paymentService) PayOrder(_ context.Context, req *paymentV1.PayOrderRequest) (*paymentV1.PayOrderResponse, error) {
	if req.GetOrderUuid() == "" {
		return nil, status.Error(codes.InvalidArgument, "order_uuid is required")
	}
	if req.GetUserUuid() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_uuid is required")
	}
	if req.GetPaymentMethod() == paymentV1.PaymentMethod_PAYMENT_METHOD_UNSPECIFIED {
		return nil, status.Error(codes.InvalidArgument, "payment_method is required")
	}

	transactionUUID := uuid.NewString()
	log.Printf("Payment successful, order_uuid: %s, user_uuid: %s, method: %s, transaction_uuid: %s\n",
		req.GetOrderUuid(), req.GetUserUuid(), req.GetPaymentMethod(), transactionUUID)

	return &paymentV1.PayOrderResponse{TransactionUuid: transactionUUID}, nil
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

	s := grpc.NewServer(grpc.UnaryInterceptor(interceptor.UnaryValidate()))

	paymentV1.RegisterPaymentServiceServer(s, &paymentService{})

	reflection.Register(s)

	go func() {
		log.Printf("gRPC server listening on %d\n", grpcPort)
		err = s.Serve(lis)
		if err != nil {
			log.Printf("failed to serve: %v\n", err)
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
		fmt.Sprintf("localhost:%d", grpcPort),
		[]grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())},
	)
	if err != nil {
		log.Printf("failed to register gateway: %v\n", err)
		return
	}

	httpMux := http.NewServeMux()
	httpMux.Handle("/api/", gwMux)
	httpMux.Handle("/", swagger.Handler("payment/v1/payment.swagger.json"))

	gwServer := &http.Server{
		Addr:              fmt.Sprintf(":%d", httpPort),
		Handler:           httpMux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("HTTP gateway listening on %d (Swagger UI: /swagger-ui.html)\n", httpPort)
		if serr := gwServer.ListenAndServe(); serr != nil && !errors.Is(serr, http.ErrServerClosed) {
			log.Printf("failed to serve HTTP: %v\n", serr)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down servers...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := gwServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v\n", err)
	}

	s.GracefulStop()
	log.Println("Server stopped")
}
