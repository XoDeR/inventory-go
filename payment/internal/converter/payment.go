package converter

import (
	"inventory-go/payment/internal/model"
	paymentV1 "inventory-go/shared/pkg/proto/payment/v1"
)

func PayOrderInfoToModel(req *paymentV1.PayOrderRequest) model.PayOrderInfo {
	return model.PayOrderInfo{
		OrderUuid:     req.GetOrderUuid(),
		UserUuid:      req.GetUserUuid(),
		PaymentMethod: model.PaymentMethod(req.GetPaymentMethod()),
	}
}
