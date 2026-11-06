//go:build e2e

package e2e

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	orderV1 "inventory-go/shared/pkg/openapi/order/v1"
)

// UUIDs of the parts the inventory service seeds on the first start
var (
	seededCPU = uuid.MustParse("6f1c8a4e-3b52-4d7a-9c10-1a2b3c4d5e01") // 399.99
	seededGPU = uuid.MustParse("6f1c8a4e-3b52-4d7a-9c10-1a2b3c4d5e02") // 549.00
)

var _ = Describe("OrderService", func() {
	var (
		ctx    context.Context
		cancel context.CancelFunc
		client *orderV1.Client
	)

	BeforeEach(func() {
		ctx, cancel = context.WithCancel(suiteCtx)

		var err error
		client, err = orderV1.NewClient("http://" + env.Order.Address())
		Expect(err).ToNot(HaveOccurred(), "expected to create the order API client")
	})

	AfterEach(func() {
		cancel()
	})

	// createOrder creates an order for the given parts and returns its response
	createOrder := func(partUUIDs ...uuid.UUID) *orderV1.CreateOrderResponse {
		res, err := client.CreateOrder(ctx, &orderV1.CreateOrderRequest{
			UserUUID:  uuid.New(),
			PartUuids: partUUIDs,
		})
		Expect(err).ToNot(HaveOccurred())

		created, ok := res.(*orderV1.CreateOrderResponse)
		Expect(ok).To(BeTrue(), "expected a successful CreateOrder response, got %T", res)

		return created
	}

	Describe("CreateOrder", func() {
		It("creates an order and sums the prices of the parts", func() {
			created := createOrder(seededCPU, seededGPU)

			Expect(created.UUID).ToNot(Equal(uuid.Nil))
			Expect(created.TotalPrice).To(BeNumerically("~", 948.99, 0.001))
		})

		It("rejects an order for an unknown part", func() {
			res, err := client.CreateOrder(ctx, &orderV1.CreateOrderRequest{
				UserUUID:  uuid.New(),
				PartUuids: []uuid.UUID{uuid.New()},
			})

			Expect(err).ToNot(HaveOccurred())

			badRequest, ok := res.(*orderV1.CreateOrderBadRequest)
			Expect(ok).To(BeTrue(), "expected a bad request response, got %T", res)
			Expect(badRequest.Code).To(Equal(http.StatusBadRequest))
		})
	})

	Describe("GetOrderByUuid", func() {
		It("returns a pending order", func() {
			created := createOrder(seededCPU)

			res, err := client.GetOrderByUuid(ctx, orderV1.GetOrderByUuidParams{OrderUUID: created.UUID})
			Expect(err).ToNot(HaveOccurred())

			order, ok := res.(*orderV1.Order)
			Expect(ok).To(BeTrue(), "expected an order, got %T", res)
			Expect(order.UUID).To(Equal(created.UUID))
			Expect(order.Status).To(Equal(orderV1.OrderStatusPENDINGPAYMENT))
			Expect(order.PartUuids).To(ConsistOf(seededCPU))
			Expect(order.TransactionUUID.IsSet() && !order.TransactionUUID.IsNull()).To(BeFalse())
		})

		It("returns NotFound for an unknown order", func() {
			res, err := client.GetOrderByUuid(ctx, orderV1.GetOrderByUuidParams{OrderUUID: uuid.New()})

			Expect(err).ToNot(HaveOccurred())

			notFound, ok := res.(*orderV1.GetOrderByUuidNotFound)
			Expect(ok).To(BeTrue(), "expected a not found response, got %T", res)
			Expect(notFound.Code).To(Equal(http.StatusNotFound))
		})
	})

	Describe("PayOrder", func() {
		It("pays an order through the payment service", func() {
			created := createOrder(seededCPU)

			res, err := client.PayOrder(ctx,
				&orderV1.PayOrderRequest{PaymentMethod: orderV1.PaymentMethodCARD},
				orderV1.PayOrderParams{OrderUUID: created.UUID},
			)
			Expect(err).ToNot(HaveOccurred())

			paid, ok := res.(*orderV1.PayOrderResponse)
			Expect(ok).To(BeTrue(), "expected a successful PayOrder response, got %T", res)
			Expect(paid.TransactionUUID).ToNot(Equal(uuid.Nil))

			getRes, err := client.GetOrderByUuid(ctx, orderV1.GetOrderByUuidParams{OrderUUID: created.UUID})
			Expect(err).ToNot(HaveOccurred())

			order, ok := getRes.(*orderV1.Order)
			Expect(ok).To(BeTrue())
			Expect(order.Status).To(Equal(orderV1.OrderStatusPAID))

			transactionUUID, set := order.TransactionUUID.Get()
			Expect(set).To(BeTrue())
			Expect(transactionUUID).To(Equal(paid.TransactionUUID))
		})

		It("completes a paid order once the computer is assembled", func() {
			created := createOrder(seededCPU)

			_, err := client.PayOrder(ctx,
				&orderV1.PayOrderRequest{PaymentMethod: orderV1.PaymentMethodCARD},
				orderV1.PayOrderParams{OrderUUID: created.UUID},
			)
			Expect(err).ToNot(HaveOccurred())

			// OrderPaid -> assembly service (waits) -> ComputerAssembled -> order service
			Eventually(func() orderV1.OrderStatus {
				getRes, err := client.GetOrderByUuid(ctx, orderV1.GetOrderByUuidParams{OrderUUID: created.UUID})
				Expect(err).ToNot(HaveOccurred())

				order, ok := getRes.(*orderV1.Order)
				Expect(ok).To(BeTrue())

				return order.Status
			}).WithTimeout(time.Minute).WithPolling(time.Second).Should(Equal(orderV1.OrderStatusCOMPLETED))
		})

		It("returns NotFound when paying an unknown order", func() {
			res, err := client.PayOrder(ctx,
				&orderV1.PayOrderRequest{PaymentMethod: orderV1.PaymentMethodCARD},
				orderV1.PayOrderParams{OrderUUID: uuid.New()},
			)

			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(BeAssignableToTypeOf(&orderV1.PayOrderNotFound{}))
		})
	})

	Describe("CancelOrder", func() {
		It("cancels a pending order", func() {
			created := createOrder(seededGPU)

			res, err := client.CancelOrder(ctx, orderV1.CancelOrderParams{OrderUUID: created.UUID})
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(BeAssignableToTypeOf(&orderV1.CancelOrderNoContent{}))

			getRes, err := client.GetOrderByUuid(ctx, orderV1.GetOrderByUuidParams{OrderUUID: created.UUID})
			Expect(err).ToNot(HaveOccurred())

			order, ok := getRes.(*orderV1.Order)
			Expect(ok).To(BeTrue())
			Expect(order.Status).To(Equal(orderV1.OrderStatusCANCELLED))
		})

		It("does not cancel a paid order", func() {
			created := createOrder(seededCPU)

			_, err := client.PayOrder(ctx,
				&orderV1.PayOrderRequest{PaymentMethod: orderV1.PaymentMethodCREDITCARD},
				orderV1.PayOrderParams{OrderUUID: created.UUID},
			)
			Expect(err).ToNot(HaveOccurred())

			res, err := client.CancelOrder(ctx, orderV1.CancelOrderParams{OrderUUID: created.UUID})
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(BeAssignableToTypeOf(&orderV1.CancelOrderConflict{}))
		})

		It("returns NotFound when cancelling an unknown order", func() {
			res, err := client.CancelOrder(ctx, orderV1.CancelOrderParams{OrderUUID: uuid.New()})

			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(BeAssignableToTypeOf(&orderV1.CancelOrderNotFound{}))
		})
	})
})
