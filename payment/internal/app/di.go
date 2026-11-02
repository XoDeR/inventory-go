package app

import (
	"context"

	paymentV1API "inventory-go/payment/internal/api/payment/v1"
	"inventory-go/payment/internal/service"
	paymentService "inventory-go/payment/internal/service/payment"
	paymentV1 "inventory-go/shared/pkg/proto/payment/v1"
)

// diContainer lazily builds the dependencies of the service and keeps a single instance of each.
type diContainer struct {
	paymentV1API paymentV1.PaymentServiceServer

	paymentService service.PaymentService
}

func NewDiContainer() *diContainer {
	return &diContainer{}
}

func (d *diContainer) PaymentV1API(ctx context.Context) paymentV1.PaymentServiceServer {
	if d.paymentV1API == nil {
		d.paymentV1API = paymentV1API.NewAPI(d.PaymentService(ctx))
	}

	return d.paymentV1API
}

func (d *diContainer) PaymentService(_ context.Context) service.PaymentService {
	if d.paymentService == nil {
		d.paymentService = paymentService.NewService()
	}

	return d.paymentService
}
