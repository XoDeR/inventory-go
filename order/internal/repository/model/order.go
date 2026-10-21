package model

import "github.com/google/uuid"

type OrderStatus string

const (
	OrderStatusPendingPayment OrderStatus = "PENDING_PAYMENT"
	OrderStatusPaid           OrderStatus = "PAID"
	OrderStatusCancelled      OrderStatus = "CANCELLED"
)

type PaymentMethod string

type Order struct {
	Uuid            uuid.UUID
	UserUuid        uuid.UUID
	PartUuids       []uuid.UUID
	TotalPrice      float64
	TransactionUuid *uuid.UUID
	PaymentMethod   *PaymentMethod
	Status          OrderStatus
}
