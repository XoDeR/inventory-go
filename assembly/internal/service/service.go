package service

import (
	"context"

	"inventory-go/assembly/internal/model"
)

// AssemblyService assembles computers of paid orders.
type AssemblyService interface {
	// Assemble builds the computer of the paid order (takes the configured time)
	// and publishes ComputerAssembled afterwards.
	Assemble(ctx context.Context, event model.OrderPaidEvent) error
}

// AssemblyProducerService publishes assembly events to Kafka.
type AssemblyProducerService interface {
	ProduceComputerAssembled(ctx context.Context, event model.ComputerAssembledEvent) error
}
