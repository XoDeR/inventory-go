package assembly_producer

import (
	"context"

	"inventory-go/assembly/internal/model"
	def "inventory-go/assembly/internal/service"
	"inventory-go/platform/pkg/kafka"
	"inventory-go/shared/pkg/events"
)

var _ def.AssemblyProducerService = (*service)(nil)

type service struct {
	computerAssembledProducer kafka.Producer
}

func NewService(computerAssembledProducer kafka.Producer) *service {
	return &service{computerAssembledProducer: computerAssembledProducer}
}

// ProduceComputerAssembled sends the ComputerAssembled event, keyed by the order UUID.
func (s *service) ProduceComputerAssembled(ctx context.Context, event model.ComputerAssembledEvent) error {
	payload, err := events.Encode(events.ComputerAssembled{
		EventUUID:    event.EventUUID.String(),
		OrderUUID:    event.OrderUUID.String(),
		UserUUID:     event.UserUUID.String(),
		BuildTimeSec: int64(event.BuildTime.Seconds()),
	})
	if err != nil {
		return err
	}

	return s.computerAssembledProducer.Send(ctx, []byte(event.OrderUUID.String()), payload)
}
