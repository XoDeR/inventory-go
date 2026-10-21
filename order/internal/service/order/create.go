package order

import (
	"context"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"

	"inventory-go/order/internal/model"
)

func (s *service) Create(ctx context.Context, info model.CreateOrderInfo) (model.Order, error) {
	if len(info.PartUuids) == 0 {
		return model.Order{}, model.ErrEmptyPartUuids
	}

	// TODO: fetch parts via InventoryService.ListParts, return an error if any is missing
	// and sum their real prices. Until the inventory client exists, prices are generated.
	var totalPrice float64
	for range info.PartUuids {
		totalPrice += gofakeit.Price(100, 1000)
	}

	order := model.Order{
		Uuid:       uuid.New(),
		UserUuid:   info.UserUuid,
		PartUuids:  info.PartUuids,
		TotalPrice: totalPrice,
		Status:     model.OrderStatusPendingPayment,
	}

	if err := s.orderRepository.Create(ctx, order); err != nil {
		return model.Order{}, err
	}

	return order, nil
}
