package assembly

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"inventory-go/assembly/internal/model"
	"inventory-go/assembly/internal/service/mocks"
)

func TestAssemblePublishesAfterTheDelay(t *testing.T) {
	ctx := context.Background()
	producer := mocks.NewAssemblyProducerService(t)
	duration := 50 * time.Millisecond
	s := NewService(producer, duration)

	event := model.OrderPaidEvent{EventUUID: uuid.New(), OrderUUID: uuid.New(), UserUUID: uuid.New()}

	producer.EXPECT().ProduceComputerAssembled(ctx, mock.MatchedBy(func(e model.ComputerAssembledEvent) bool {
		return e.OrderUUID == event.OrderUUID && e.UserUUID == event.UserUUID && e.BuildTime == duration && e.EventUUID != uuid.Nil
	})).Return(nil)

	start := time.Now()
	require.NoError(t, s.Assemble(ctx, event))
	require.GreaterOrEqual(t, time.Since(start), duration)
}

func TestAssembleProducerError(t *testing.T) {
	producer := mocks.NewAssemblyProducerService(t)
	s := NewService(producer, time.Millisecond)
	producerErr := errors.New("kafka is down")

	producer.EXPECT().ProduceComputerAssembled(mock.Anything, mock.Anything).Return(producerErr)

	require.ErrorIs(t, s.Assemble(context.Background(), model.OrderPaidEvent{}), producerErr)
}

func TestAssembleCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// no producer expectation: nothing may be published when the context is cancelled
	s := NewService(mocks.NewAssemblyProducerService(t), time.Hour)

	require.ErrorIs(t, s.Assemble(ctx, model.OrderPaidEvent{}), context.Canceled)
}
