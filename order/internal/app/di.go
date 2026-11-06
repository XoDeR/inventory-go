package app

import (
	"context"
	"fmt"

	"github.com/IBM/sarama"
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
	consumerService "inventory-go/order/internal/service/consumer/order_consumer"
	orderService "inventory-go/order/internal/service/order"
	producerService "inventory-go/order/internal/service/producer/order_producer"
	"inventory-go/platform/pkg/closer"
	"inventory-go/platform/pkg/kafka"
	kafkaConsumer "inventory-go/platform/pkg/kafka/consumer"
	kafkaProducer "inventory-go/platform/pkg/kafka/producer"
	"inventory-go/platform/pkg/logger"
	kafkaMiddleware "inventory-go/platform/pkg/middleware/kafka"
	orderV1 "inventory-go/shared/pkg/openapi/order/v1"
	inventoryV1 "inventory-go/shared/pkg/proto/inventory/v1"
	paymentV1 "inventory-go/shared/pkg/proto/payment/v1"
)

// diContainer lazily builds the dependencies of the service and keeps a single instance of each.
// Constructors panic on a failure: the container is only used while the app is being assembled.
type diContainer struct {
	orderV1API orderV1.Handler

	orderService service.OrderService

	orderProducerService service.OrderProducerService
	orderConsumerService service.OrderConsumerService

	computerAssembledConsumer kafka.Consumer
	orderPaidProducer         kafka.Producer

	consumerGroup sarama.ConsumerGroup
	syncProducer  sarama.SyncProducer

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
			d.OrderProducerService(ctx),
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

func (d *diContainer) OrderProducerService(ctx context.Context) service.OrderProducerService {
	if d.orderProducerService == nil {
		d.orderProducerService = producerService.NewService(d.OrderPaidProducer(ctx))
	}

	return d.orderProducerService
}

func (d *diContainer) OrderConsumerService(ctx context.Context) service.OrderConsumerService {
	if d.orderConsumerService == nil {
		d.orderConsumerService = consumerService.NewService(d.ComputerAssembledConsumer(), d.OrderRepository(ctx))
	}

	return d.orderConsumerService
}

func (d *diContainer) OrderPaidProducer(_ context.Context) kafka.Producer {
	if d.orderPaidProducer == nil {
		d.orderPaidProducer = kafkaProducer.NewProducer(d.SyncProducer(), config.AppConfig().OrderPaidProducer.Topic(), logger.Default())
	}

	return d.orderPaidProducer
}

func (d *diContainer) ComputerAssembledConsumer() kafka.Consumer {
	if d.computerAssembledConsumer == nil {
		d.computerAssembledConsumer = kafkaConsumer.NewConsumer(
			d.ConsumerGroup(),
			[]string{config.AppConfig().OrderAssembledConsumer.Topic()},
			logger.Default(),
			kafkaMiddleware.Logging(logger.Default()),
		)
	}

	return d.computerAssembledConsumer
}

func (d *diContainer) ConsumerGroup() sarama.ConsumerGroup {
	if d.consumerGroup == nil {
		group, err := sarama.NewConsumerGroup(
			config.AppConfig().Kafka.Brokers(),
			config.AppConfig().OrderAssembledConsumer.GroupID(),
			config.AppConfig().OrderAssembledConsumer.Config(),
		)
		if err != nil {
			panic(fmt.Sprintf("failed to create the consumer group: %v", err))
		}

		closer.AddNamed("Kafka consumer group", func(_ context.Context) error {
			return group.Close()
		})

		d.consumerGroup = group
	}

	return d.consumerGroup
}

func (d *diContainer) SyncProducer() sarama.SyncProducer {
	if d.syncProducer == nil {
		producer, err := sarama.NewSyncProducer(
			config.AppConfig().Kafka.Brokers(),
			config.AppConfig().OrderPaidProducer.Config(),
		)
		if err != nil {
			panic(fmt.Sprintf("failed to create the sync producer: %v", err))
		}

		closer.AddNamed("Kafka sync producer", func(_ context.Context) error {
			return producer.Close()
		})

		d.syncProducer = producer
	}

	return d.syncProducer
}
