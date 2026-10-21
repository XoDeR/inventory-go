package order

import (
	"context"

	"github.com/google/uuid"

	"inventory-go/order/internal/model"
)

func (s *service) Create(ctx context.Context, info model.CreateOrderInfo) (model.Order, error) {
	if len(info.PartUuids) == 0 {
		return model.Order{}, model.ErrEmptyPartUuids
	}

	uuids := make([]string, 0, len(info.PartUuids))
	for _, partUUID := range info.PartUuids {
		uuids = append(uuids, partUUID.String())
	}

	parts, err := s.inventoryClient.ListParts(ctx, model.PartsFilter{Uuids: uuids})
	if err != nil {
		return model.Order{}, err
	}

	prices := make(map[string]float64, len(parts))
	for _, part := range parts {
		prices[part.Uuid] = part.Price
	}

	var totalPrice float64
	for _, id := range uuids {
		price, ok := prices[id]
		if !ok {
			return model.Order{}, model.ErrPartsNotFound
		}
		totalPrice += price
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
