package converter

import (
	"inventory-go/order/internal/model"
	paymentV1 "inventory-go/shared/pkg/proto/payment/v1"
)

func PayOrderInfoToProto(info model.PayOrderInfo) *paymentV1.PayOrderRequest {
	return &paymentV1.PayOrderRequest{
		OrderUuid:     info.OrderUuid.String(),
		UserUuid:      info.UserUuid.String(),
		PaymentMethod: paymentMethodToProto(info.PaymentMethod),
	}
}

func paymentMethodToProto(method model.PaymentMethod) paymentV1.PaymentMethod {
	switch method {
	case model.PaymentMethodCard:
		return paymentV1.PaymentMethod_PAYMENT_METHOD_CARD
	case model.PaymentMethodCreditCard:
		return paymentV1.PaymentMethod_PAYMENT_METHOD_CREDIT_CARD
	case model.PaymentMethodStoreCredit:
		return paymentV1.PaymentMethod_PAYMENT_METHOD_STORE_CREDIT
	default:
		return paymentV1.PaymentMethod_PAYMENT_METHOD_UNSPECIFIED
	}
}
