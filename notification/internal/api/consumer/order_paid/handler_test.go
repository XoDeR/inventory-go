package order_paid

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"inventory-go/notification/internal/model"
	"inventory-go/notification/internal/service/mocks"
	"inventory-go/platform/pkg/kafka"
	"inventory-go/shared/pkg/events"
)

func TestHandle(t *testing.T) {
	ctx := context.Background()
	svc := mocks.NewNotificationService(t)
	c := NewConsumer(nil, svc)

	e := events.OrderPaid{
		EventUUID:       uuid.NewString(),
		OrderUUID:       uuid.NewString(),
		UserUUID:        uuid.NewString(),
		PaymentMethod:   "CARD",
		TransactionUUID: uuid.NewString(),
	}
	payload, err := events.Encode(e)
	require.NoError(t, err)

	svc.EXPECT().NotifyOrderPaid(ctx, model.OrderPaidEvent{
		EventUUID:       uuid.MustParse(e.EventUUID),
		OrderUUID:       uuid.MustParse(e.OrderUUID),
		UserUUID:        uuid.MustParse(e.UserUUID),
		PaymentMethod:   "CARD",
		TransactionUUID: uuid.MustParse(e.TransactionUUID),
	}).Return(nil)

	require.NoError(t, c.Handle(ctx, kafka.Message{Value: payload}))
}

func TestHandleInvalidMessage(t *testing.T) {
	c := NewConsumer(nil, mocks.NewNotificationService(t))

	require.Error(t, c.Handle(context.Background(), kafka.Message{Value: []byte("not json")}))
}
