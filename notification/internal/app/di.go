package app

import (
	"context"
	"fmt"

	"github.com/IBM/sarama"

	"inventory-go/notification/internal/api/consumer/computer_assembled"
	"inventory-go/notification/internal/api/consumer/order_paid"
	"inventory-go/notification/internal/client/notifier"
	logNotifier "inventory-go/notification/internal/client/notifier/log"
	"inventory-go/notification/internal/config"
	"inventory-go/notification/internal/service"
	notificationService "inventory-go/notification/internal/service/notification"
	"inventory-go/platform/pkg/closer"
	"inventory-go/platform/pkg/kafka"
	kafkaConsumer "inventory-go/platform/pkg/kafka/consumer"
	"inventory-go/platform/pkg/logger"
	kafkaMiddleware "inventory-go/platform/pkg/middleware/kafka"
)

// diContainer lazily builds the dependencies of the service and keeps a single instance of each.
// Constructors panic on a failure: the container is only used while the app is being assembled.
type diContainer struct {
	orderPaidConsumer         *order_paid.Consumer
	computerAssembledConsumer *computer_assembled.Consumer

	notificationService service.NotificationService
	notifier            notifier.Notifier

	kafkaOrderPaidConsumer         kafka.Consumer
	kafkaComputerAssembledConsumer kafka.Consumer

	orderPaidGroup         sarama.ConsumerGroup
	computerAssembledGroup sarama.ConsumerGroup
}

func NewDiContainer() *diContainer {
	return &diContainer{}
}

func (d *diContainer) OrderPaidConsumer() *order_paid.Consumer {
	if d.orderPaidConsumer == nil {
		d.orderPaidConsumer = order_paid.NewConsumer(d.KafkaOrderPaidConsumer(), d.NotificationService())
	}

	return d.orderPaidConsumer
}

func (d *diContainer) ComputerAssembledConsumer() *computer_assembled.Consumer {
	if d.computerAssembledConsumer == nil {
		d.computerAssembledConsumer = computer_assembled.NewConsumer(d.KafkaComputerAssembledConsumer(), d.NotificationService())
	}

	return d.computerAssembledConsumer
}

func (d *diContainer) NotificationService() service.NotificationService {
	if d.notificationService == nil {
		d.notificationService = notificationService.NewService(d.Notifier())
	}

	return d.notificationService
}

// Notifier is the place to plug in the Slack notifier later.
func (d *diContainer) Notifier() notifier.Notifier {
	if d.notifier == nil {
		d.notifier = logNotifier.NewNotifier()
	}

	return d.notifier
}

func (d *diContainer) KafkaOrderPaidConsumer() kafka.Consumer {
	if d.kafkaOrderPaidConsumer == nil {
		d.kafkaOrderPaidConsumer = kafkaConsumer.NewConsumer(
			d.OrderPaidGroup(),
			[]string{config.AppConfig().OrderPaidConsumer.Topic()},
			logger.Default(),
			kafkaMiddleware.Logging(logger.Default()),
		)
	}

	return d.kafkaOrderPaidConsumer
}

func (d *diContainer) KafkaComputerAssembledConsumer() kafka.Consumer {
	if d.kafkaComputerAssembledConsumer == nil {
		d.kafkaComputerAssembledConsumer = kafkaConsumer.NewConsumer(
			d.ComputerAssembledGroup(),
			[]string{config.AppConfig().ComputerAssembledConsumer.Topic()},
			logger.Default(),
			kafkaMiddleware.Logging(logger.Default()),
		)
	}

	return d.kafkaComputerAssembledConsumer
}

func (d *diContainer) OrderPaidGroup() sarama.ConsumerGroup {
	if d.orderPaidGroup == nil {
		d.orderPaidGroup = newConsumerGroup(
			"order paid",
			config.AppConfig().OrderPaidConsumer.GroupID(),
			config.AppConfig().OrderPaidConsumer.Config(),
		)
	}

	return d.orderPaidGroup
}

func (d *diContainer) ComputerAssembledGroup() sarama.ConsumerGroup {
	if d.computerAssembledGroup == nil {
		d.computerAssembledGroup = newConsumerGroup(
			"computer assembled",
			config.AppConfig().ComputerAssembledConsumer.GroupID(),
			config.AppConfig().ComputerAssembledConsumer.Config(),
		)
	}

	return d.computerAssembledGroup
}

func newConsumerGroup(name, groupID string, cfg *sarama.Config) sarama.ConsumerGroup {
	group, err := sarama.NewConsumerGroup(config.AppConfig().Kafka.Brokers(), groupID, cfg)
	if err != nil {
		panic(fmt.Sprintf("failed to create the %s consumer group: %v", name, err))
	}

	closer.AddNamed("Kafka consumer group "+name, func(_ context.Context) error {
		return group.Close()
	})

	return group
}
