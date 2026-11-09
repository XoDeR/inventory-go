package notification

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"inventory-go/notification/internal/client/notifier/mocks"
	"inventory-go/notification/internal/model"
)

func TestNotifyOrderPaid(t *testing.T) {
	ctx := context.Background()
	n := mocks.NewNotifier(t)
	s := NewService(n)

	event := model.OrderPaidEvent{
		OrderUUID:       uuid.New(),
		UserUUID:        uuid.New(),
		PaymentMethod:   "CARD",
		TransactionUUID: uuid.New(),
	}

	n.EXPECT().Send(ctx, ":moneybag: *Order paid*\n"+
		"Order: `"+event.OrderUUID.String()+"`\n"+
		"User: `"+event.UserUUID.String()+"`\n"+
		"Payment method: CARD\n"+
		"Transaction: `"+event.TransactionUUID.String()+"`").Return(nil)

	require.NoError(t, s.NotifyOrderPaid(ctx, event))
}

func TestNotifyComputerAssembled(t *testing.T) {
	ctx := context.Background()
	n := mocks.NewNotifier(t)
	s := NewService(n)

	event := model.ComputerAssembledEvent{OrderUUID: uuid.New(), UserUUID: uuid.New(), BuildTimeSec: 10}

	n.EXPECT().Send(ctx, ":computer: *Computer assembled*\n"+
		"Order: `"+event.OrderUUID.String()+"`\n"+
		"User: `"+event.UserUUID.String()+"`\n"+
		"Build time: 10 s").Return(nil)

	require.NoError(t, s.NotifyComputerAssembled(ctx, event))
}

func TestNotifySendError(t *testing.T) {
	n := mocks.NewNotifier(t)
	s := NewService(n)
	sendErr := errors.New("channel is unavailable")

	n.EXPECT().Send(context.Background(), mock.Anything).Return(sendErr)

	err := s.NotifyComputerAssembled(context.Background(), model.ComputerAssembledEvent{})
	require.ErrorIs(t, err, sendErr)
}
