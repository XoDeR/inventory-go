package model

import "github.com/google/uuid"

// OrderPaidEvent is the event received after an order was paid.
type OrderPaidEvent struct {
	EventUUID       uuid.UUID
	OrderUUID       uuid.UUID
	UserUUID        uuid.UUID
	PaymentMethod   string
	TransactionUUID uuid.UUID
}

// ComputerAssembledEvent is the event received after a computer was assembled.
type ComputerAssembledEvent struct {
	EventUUID    uuid.UUID
	OrderUUID    uuid.UUID
	UserUUID     uuid.UUID
	BuildTimeSec int64
}
