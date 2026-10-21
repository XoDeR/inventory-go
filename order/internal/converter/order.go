package converter

import (
	"inventory-go/order/internal/model"
	orderV1 "inventory-go/shared/pkg/openapi/order/v1"
)

func CreateOrderInfoToModel(req *orderV1.CreateOrderRequest) model.CreateOrderInfo {
	return model.CreateOrderInfo{
		UserUuid:  req.UserUUID,
		PartUuids: req.PartUuids,
	}
}

func PaymentMethodToModel(method orderV1.PaymentMethod) model.PaymentMethod {
	return model.PaymentMethod(method)
}

func OrderToOpenAPI(order model.Order) *orderV1.Order {
	result := &orderV1.Order{
		UUID:       order.Uuid,
		UserUUID:   order.UserUuid,
		PartUuids:  order.PartUuids,
		TotalPrice: order.TotalPrice,
		Status:     orderV1.OrderStatus(order.Status),
	}

	if order.TransactionUuid != nil {
		result.TransactionUUID = orderV1.NewOptNilUUID(*order.TransactionUuid)
	}
	if order.PaymentMethod != nil {
		result.PaymentMethod = orderV1.NewOptNilPaymentMethod(orderV1.PaymentMethod(*order.PaymentMethod))
	}

	return result
}
