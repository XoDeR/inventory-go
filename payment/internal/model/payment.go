package model

import "errors"

var ErrInvalidArgument = errors.New("invalid argument")

type PaymentMethod int32

const (
	PaymentMethodUnspecified PaymentMethod = iota
	PaymentMethodCard
	PaymentMethodCreditCard
	PaymentMethodStoreCredit
)

type PayOrderInfo struct {
	OrderUuid     string
	UserUuid      string
	PaymentMethod PaymentMethod
}
