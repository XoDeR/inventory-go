package app

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	paymentV1API "inventory-go/payment/internal/api/payment/v1"
	"inventory-go/payment/internal/config"
	"inventory-go/payment/internal/service"
	paymentService "inventory-go/payment/internal/service/payment"
	"inventory-go/platform/pkg/closer"
	authV1 "inventory-go/shared/pkg/proto/auth/v1"
	paymentV1 "inventory-go/shared/pkg/proto/payment/v1"
)

// diContainer lazily builds the dependencies of the service and keeps a single instance of each.
type diContainer struct {
	paymentV1API paymentV1.PaymentServiceServer

	paymentService service.PaymentService

	authClient authV1.AuthServiceClient
}

func NewDiContainer() *diContainer {
	return &diContainer{}
}

func (d *diContainer) PaymentV1API(ctx context.Context) paymentV1.PaymentServiceServer {
	if d.paymentV1API == nil {
		d.paymentV1API = paymentV1API.NewAPI(d.PaymentService(ctx))
	}

	return d.paymentV1API
}

func (d *diContainer) PaymentService(_ context.Context) service.PaymentService {
	if d.paymentService == nil {
		d.paymentService = paymentService.NewService()
	}

	return d.paymentService
}

// AuthClient is the client of the auth service, used by the auth interceptor to check sessions.
// The connection is lazy: the service starts even when auth is not up yet, and requests are
// rejected as unauthenticated until it is.
func (d *diContainer) AuthClient(_ context.Context) authV1.AuthServiceClient {
	if d.authClient == nil {
		conn, err := grpc.NewClient(
			config.AppConfig().Auth.Address(),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			panic(fmt.Sprintf("failed to connect to auth: %v", err))
		}

		closer.AddNamed("auth gRPC connection", func(_ context.Context) error {
			return conn.Close()
		})

		d.authClient = authV1.NewAuthServiceClient(conn)
	}

	return d.authClient
}
