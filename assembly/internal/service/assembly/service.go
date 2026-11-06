package assembly

import (
	"context"
	"time"

	"github.com/google/uuid"

	"inventory-go/assembly/internal/model"
	def "inventory-go/assembly/internal/service"
	"inventory-go/platform/pkg/logger"
)

var _ def.AssemblyService = (*service)(nil)

type service struct {
	producer def.AssemblyProducerService
	duration time.Duration
}

func NewService(producer def.AssemblyProducerService, duration time.Duration) *service {
	return &service{producer: producer, duration: duration}
}

func (s *service) Assemble(ctx context.Context, event model.OrderPaidEvent) error {
	logger.Info(ctx, "Assembly started", "order_uuid", event.OrderUUID, "duration", s.duration)

	timer := time.NewTimer(s.duration)
	defer timer.Stop()

	select {
	case <-timer.C:
	case <-ctx.Done():
		// the message is not marked as processed, so it is redelivered after a restart
		return ctx.Err()
	}

	err := s.producer.ProduceComputerAssembled(ctx, model.ComputerAssembledEvent{
		EventUUID: uuid.New(),
		OrderUUID: event.OrderUUID,
		UserUUID:  event.UserUUID,
		BuildTime: s.duration,
	})
	if err != nil {
		return err
	}

	logger.Info(ctx, "Assembly finished", "order_uuid", event.OrderUUID)

	return nil
}
