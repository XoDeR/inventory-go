package app

import (
	"context"
	"fmt"

	"github.com/IBM/sarama"

	"inventory-go/assembly/internal/api/consumer/order_paid"
	"inventory-go/assembly/internal/config"
	"inventory-go/assembly/internal/service"
	assemblyService "inventory-go/assembly/internal/service/assembly"
	producerService "inventory-go/assembly/internal/service/producer/assembly_producer"
	"inventory-go/platform/pkg/closer"
	"inventory-go/platform/pkg/kafka"
	kafkaConsumer "inventory-go/platform/pkg/kafka/consumer"
	kafkaProducer "inventory-go/platform/pkg/kafka/producer"
	"inventory-go/platform/pkg/logger"
	kafkaMiddleware "inventory-go/platform/pkg/middleware/kafka"
)

// diContainer lazily builds the dependencies of the service and keeps a single instance of each.
// Constructors panic on a failure: the container is only used while the app is being assembled.
type diContainer struct {
	orderPaidConsumer *order_paid.Consumer

	assemblyService         service.AssemblyService
	assemblyProducerService service.AssemblyProducerService

	kafkaOrderPaidConsumer    kafka.Consumer
	computerAssembledProducer kafka.Producer

	consumerGroup sarama.ConsumerGroup
	syncProducer  sarama.SyncProducer
}

func NewDiContainer() *diContainer {
	return &diContainer{}
}

func (d *diContainer) OrderPaidConsumer() *order_paid.Consumer {
	if d.orderPaidConsumer == nil {
		d.orderPaidConsumer = order_paid.NewConsumer(d.KafkaOrderPaidConsumer(), d.AssemblyService())
	}

	return d.orderPaidConsumer
}

func (d *diContainer) AssemblyService() service.AssemblyService {
	if d.assemblyService == nil {
		d.assemblyService = assemblyService.NewService(
			d.AssemblyProducerService(),
			config.AppConfig().Assembly.Duration(),
		)
	}

	return d.assemblyService
}

func (d *diContainer) AssemblyProducerService() service.AssemblyProducerService {
	if d.assemblyProducerService == nil {
		d.assemblyProducerService = producerService.NewService(d.ComputerAssembledProducer())
	}

	return d.assemblyProducerService
}

func (d *diContainer) KafkaOrderPaidConsumer() kafka.Consumer {
	if d.kafkaOrderPaidConsumer == nil {
		d.kafkaOrderPaidConsumer = kafkaConsumer.NewConsumer(
			d.ConsumerGroup(),
			[]string{config.AppConfig().OrderPaidConsumer.Topic()},
			logger.Default(),
			kafkaMiddleware.Logging(logger.Default()),
		)
	}

	return d.kafkaOrderPaidConsumer
}

func (d *diContainer) ComputerAssembledProducer() kafka.Producer {
	if d.computerAssembledProducer == nil {
		d.computerAssembledProducer = kafkaProducer.NewProducer(
			d.SyncProducer(),
			config.AppConfig().ComputerAssembledProducer.Topic(),
			logger.Default(),
		)
	}

	return d.computerAssembledProducer
}

func (d *diContainer) ConsumerGroup() sarama.ConsumerGroup {
	if d.consumerGroup == nil {
		group, err := sarama.NewConsumerGroup(
			config.AppConfig().Kafka.Brokers(),
			config.AppConfig().OrderPaidConsumer.GroupID(),
			config.AppConfig().OrderPaidConsumer.Config(),
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
			config.AppConfig().ComputerAssembledProducer.Config(),
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
