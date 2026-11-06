package model

import "github.com/google/uuid"

type OrderStatus string

const (
	OrderStatusPendingPayment OrderStatus = "PENDING_PAYMENT"
	OrderStatusPaid           OrderStatus = "PAID"
	OrderStatusCancelled      OrderStatus = "CANCELLED"
	OrderStatusCompleted      OrderStatus = "COMPLETED"
)

type PaymentMethod string

const (
	PaymentMethodUnknown     PaymentMethod = "UNKNOWN"
	PaymentMethodCard        PaymentMethod = "CARD"
	PaymentMethodCreditCard  PaymentMethod = "CREDIT_CARD"
	PaymentMethodStoreCredit PaymentMethod = "STORE_CREDIT"
)

type Order struct {
	Uuid            uuid.UUID
	UserUuid        uuid.UUID
	PartUuids       []uuid.UUID
	TotalPrice      float64
	TransactionUuid *uuid.UUID
	PaymentMethod   *PaymentMethod
	Status          OrderStatus
}

type CreateOrderInfo struct {
	UserUuid  uuid.UUID
	PartUuids []uuid.UUID
}

// OrderPaidEvent is the event published after an order was paid.
type OrderPaidEvent struct {
	EventUUID       uuid.UUID
	OrderUUID       uuid.UUID
	UserUUID        uuid.UUID
	PaymentMethod   PaymentMethod
	TransactionUUID uuid.UUID
}

// ComputerAssembledEvent is the event received after the computer of an order was assembled.
type ComputerAssembledEvent struct {
	EventUUID    uuid.UUID
	OrderUUID    uuid.UUID
	UserUUID     uuid.UUID
	BuildTimeSec int64
}
