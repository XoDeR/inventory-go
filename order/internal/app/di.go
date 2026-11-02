package app

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	orderV1API "inventory-go/order/internal/api/order/v1"
	grpcClient "inventory-go/order/internal/client/grpc"
	inventoryClientV1 "inventory-go/order/internal/client/grpc/inventory/v1"
	paymentClientV1 "inventory-go/order/internal/client/grpc/payment/v1"
	"inventory-go/order/internal/config"
	"inventory-go/order/internal/repository"
	orderRepository "inventory-go/order/internal/repository/order"
	"inventory-go/order/internal/service"
	orderService "inventory-go/order/internal/service/order"
	"inventory-go/platform/pkg/closer"
	orderV1 "inventory-go/shared/pkg/openapi/order/v1"
	inventoryV1 "inventory-go/shared/pkg/proto/inventory/v1"
	paymentV1 "inventory-go/shared/pkg/proto/payment/v1"
)

// diContainer lazily builds the dependencies of the service and keeps a single instance of each.
// Constructors panic on a failure: the container is only used while the app is being assembled.
type diContainer struct {
	orderV1API orderV1.Handler

	orderService service.OrderService

	orderRepository repository.OrderRepository

	inventoryClient grpcClient.InventoryClient
	paymentClient   grpcClient.PaymentClient

	inventoryConn *grpc.ClientConn
	paymentConn   *grpc.ClientConn

	postgresPool *pgxpool.Pool
}

func NewDiContainer() *diContainer {
	return &diContainer{}
}

func (d *diContainer) OrderV1API(ctx context.Context) orderV1.Handler {
	if d.orderV1API == nil {
		d.orderV1API = orderV1API.NewAPI(d.OrderService(ctx))
	}

	return d.orderV1API
}

func (d *diContainer) OrderService(ctx context.Context) service.OrderService {
	if d.orderService == nil {
		d.orderService = orderService.NewService(
			d.OrderRepository(ctx),
			d.InventoryClient(ctx),
			d.PaymentClient(ctx),
		)
	}

	return d.orderService
}

func (d *diContainer) OrderRepository(ctx context.Context) repository.OrderRepository {
	if d.orderRepository == nil {
		d.orderRepository = orderRepository.NewRepository(d.PostgresPool(ctx))
	}

	return d.orderRepository
}

func (d *diContainer) InventoryClient(ctx context.Context) grpcClient.InventoryClient {
	if d.inventoryClient == nil {
		d.inventoryClient = inventoryClientV1.NewClient(inventoryV1.NewInventoryServiceClient(d.InventoryConn(ctx)))
	}

	return d.inventoryClient
}

func (d *diContainer) PaymentClient(ctx context.Context) grpcClient.PaymentClient {
	if d.paymentClient == nil {
		d.paymentClient = paymentClientV1.NewClient(paymentV1.NewPaymentServiceClient(d.PaymentConn(ctx)))
	}

	return d.paymentClient
}

func (d *diContainer) InventoryConn(_ context.Context) *grpc.ClientConn {
	if d.inventoryConn == nil {
		conn, err := grpc.NewClient(
			config.AppConfig().Inventory.Address(),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			panic(fmt.Sprintf("failed to connect to inventory: %v", err))
		}

		closer.AddNamed("inventory gRPC connection", func(_ context.Context) error {
			return conn.Close()
		})

		d.inventoryConn = conn
	}

	return d.inventoryConn
}

func (d *diContainer) PaymentConn(_ context.Context) *grpc.ClientConn {
	if d.paymentConn == nil {
		conn, err := grpc.NewClient(
			config.AppConfig().Payment.Address(),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			panic(fmt.Sprintf("failed to connect to payment: %v", err))
		}

		closer.AddNamed("payment gRPC connection", func(_ context.Context) error {
			return conn.Close()
		})

		d.paymentConn = conn
	}

	return d.paymentConn
}

func (d *diContainer) PostgresPool(ctx context.Context) *pgxpool.Pool {
	if d.postgresPool == nil {
		pool, err := pgxpool.New(ctx, config.AppConfig().Postgres.URI())
		if err != nil {
			panic(fmt.Sprintf("failed to connect to Postgres: %v", err))
		}

		if err = pool.Ping(ctx); err != nil {
			pool.Close()
			panic(fmt.Sprintf("failed to ping Postgres: %v", err))
		}

		closer.AddNamed("Postgres pool", func(_ context.Context) error {
			pool.Close()
			return nil
		})

		d.postgresPool = pool
	}

	return d.postgresPool
}
