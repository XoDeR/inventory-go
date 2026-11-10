package order_consumer

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"inventory-go/order/internal/model"
	repoMocks "inventory-go/order/internal/repository/mocks"
	"inventory-go/platform/pkg/kafka"
	"inventory-go/shared/pkg/events"
)

func newMessage(t *testing.T, orderUUID uuid.UUID) kafka.Message {
	t.Helper()

	payload, err := events.Encode(events.ComputerAssembled{
		EventUUID:    uuid.NewString(),
		OrderUUID:    orderUUID.String(),
		UserUUID:     uuid.NewString(),
		BuildTimeSec: 10,
	})
	require.NoError(t, err)

	return kafka.Message{Value: payload}
}

func TestHandlerCompletesPaidOrder(t *testing.T) {
	ctx := context.Background()
	orderUUID := uuid.New()
	repo := repoMocks.NewOrderRepository(t)
	cacheRepo := repoMocks.NewOrderCacheRepository(t)
	s := NewService(nil, repo, cacheRepo)

	repo.EXPECT().Get(ctx, orderUUID.String()).Return(model.Order{Uuid: orderUUID, Status: model.OrderStatusPaid}, nil)
	repo.EXPECT().Update(ctx, model.Order{Uuid: orderUUID, Status: model.OrderStatusCompleted}).Return(nil)
	cacheRepo.EXPECT().Delete(ctx, orderUUID.String()).Return(nil)

	require.NoError(t, s.ComputerAssembledHandler(ctx, newMessage(t, orderUUID)))
}

func TestHandlerIsIdempotent(t *testing.T) {
	ctx := context.Background()
	orderUUID := uuid.New()
	repo := repoMocks.NewOrderRepository(t)
	cacheRepo := repoMocks.NewOrderCacheRepository(t)
	s := NewService(nil, repo, cacheRepo)

	// no Update expectation: an already completed order is left untouched
	repo.EXPECT().Get(ctx, orderUUID.String()).Return(model.Order{Uuid: orderUUID, Status: model.OrderStatusCompleted}, nil)

	require.NoError(t, s.ComputerAssembledHandler(ctx, newMessage(t, orderUUID)))
}

func TestHandlerRejectsUnpaidOrder(t *testing.T) {
	ctx := context.Background()
	orderUUID := uuid.New()
	repo := repoMocks.NewOrderRepository(t)
	cacheRepo := repoMocks.NewOrderCacheRepository(t)
	s := NewService(nil, repo, cacheRepo)

	repo.EXPECT().Get(ctx, orderUUID.String()).Return(model.Order{Uuid: orderUUID, Status: model.OrderStatusPendingPayment}, nil)

	require.Error(t, s.ComputerAssembledHandler(ctx, newMessage(t, orderUUID)))
}

func TestHandlerRepositoryError(t *testing.T) {
	ctx := context.Background()
	repo := repoMocks.NewOrderRepository(t)
	cacheRepo := repoMocks.NewOrderCacheRepository(t)
	s := NewService(nil, repo, cacheRepo)
	repoErr := errors.New("db is down")

	repo.EXPECT().Get(ctx, mock.Anything).Return(model.Order{}, repoErr)

	require.ErrorIs(t, s.ComputerAssembledHandler(ctx, newMessage(t, uuid.New())), repoErr)
}

func TestHandlerInvalidPayload(t *testing.T) {
	s := NewService(nil, repoMocks.NewOrderRepository(t), repoMocks.NewOrderCacheRepository(t))

	require.Error(t, s.ComputerAssembledHandler(context.Background(), kafka.Message{Value: []byte("not json")}))
}
