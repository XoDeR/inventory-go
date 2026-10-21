package converter

import (
	"inventory-go/order/internal/model"
	repoModel "inventory-go/order/internal/repository/model"
)

func OrderToRepoModel(order model.Order) repoModel.Order {
	var paymentMethod *repoModel.PaymentMethod
	if order.PaymentMethod != nil {
		method := repoModel.PaymentMethod(*order.PaymentMethod)
		paymentMethod = &method
	}

	return repoModel.Order{
		Uuid:            order.Uuid,
		UserUuid:        order.UserUuid,
		PartUuids:       order.PartUuids,
		TotalPrice:      order.TotalPrice,
		TransactionUuid: order.TransactionUuid,
		PaymentMethod:   paymentMethod,
		Status:          repoModel.OrderStatus(order.Status),
	}
}

func OrderToModel(order repoModel.Order) model.Order {
	var paymentMethod *model.PaymentMethod
	if order.PaymentMethod != nil {
		method := model.PaymentMethod(*order.PaymentMethod)
		paymentMethod = &method
	}

	return model.Order{
		Uuid:            order.Uuid,
		UserUuid:        order.UserUuid,
		PartUuids:       order.PartUuids,
		TotalPrice:      order.TotalPrice,
		TransactionUuid: order.TransactionUuid,
		PaymentMethod:   paymentMethod,
		Status:          model.OrderStatus(order.Status),
	}
}
