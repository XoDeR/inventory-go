package model

import "github.com/google/uuid"

type PayOrderInfo struct {
	OrderUuid     uuid.UUID
	UserUuid      uuid.UUID
	PaymentMethod PaymentMethod
}
