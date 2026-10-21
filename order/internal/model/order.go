package model

import "github.com/google/uuid"

type OrderStatus string

const (
	OrderStatusPendingPayment OrderStatus = "PENDING_PAYMENT"
	OrderStatusPaid           OrderStatus = "PAID"
	OrderStatusCancelled      OrderStatus = "CANCELLED"
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
