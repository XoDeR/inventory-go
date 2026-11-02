//go:build e2e

package e2e

import (
	"context"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"

	paymentV1 "inventory-go/shared/pkg/proto/payment/v1"
)

var _ = Describe("PaymentService", func() {
	var (
		ctx           context.Context
		cancel        context.CancelFunc
		conn          *grpc.ClientConn
		paymentClient paymentV1.PaymentServiceClient
	)

	BeforeEach(func() {
		ctx, cancel = context.WithCancel(suiteCtx)

		var err error
		conn, err = grpc.NewClient(
			env.App.Address(),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		Expect(err).ToNot(HaveOccurred(), "expected to connect to the gRPC app")

		paymentClient = paymentV1.NewPaymentServiceClient(conn)
	})

	AfterEach(func() {
		Expect(conn.Close()).To(Succeed())
		cancel()
	})

	Describe("PayOrder", func() {
		It("pays an order and returns a transaction UUID", func() {
			resp, err := paymentClient.PayOrder(ctx, &paymentV1.PayOrderRequest{
				OrderUuid:     uuid.NewString(),
				UserUuid:      uuid.NewString(),
				PaymentMethod: paymentV1.PaymentMethod_PAYMENT_METHOD_CARD,
			})

			Expect(err).ToNot(HaveOccurred())
			Expect(resp.GetTransactionUuid()).ToNot(BeEmpty())

			_, err = uuid.Parse(resp.GetTransactionUuid())
			Expect(err).ToNot(HaveOccurred(), "expected the transaction UUID to be a valid UUID")
		})

		It("returns a different transaction UUID for every payment", func() {
			req := &paymentV1.PayOrderRequest{
				OrderUuid:     uuid.NewString(),
				UserUuid:      uuid.NewString(),
				PaymentMethod: paymentV1.PaymentMethod_PAYMENT_METHOD_CREDIT_CARD,
			}

			first, err := paymentClient.PayOrder(ctx, req)
			Expect(err).ToNot(HaveOccurred())

			second, err := paymentClient.PayOrder(ctx, req)
			Expect(err).ToNot(HaveOccurred())

			Expect(first.GetTransactionUuid()).ToNot(Equal(second.GetTransactionUuid()))
		})

		It("rejects a malformed order UUID", func() {
			_, err := paymentClient.PayOrder(ctx, &paymentV1.PayOrderRequest{
				OrderUuid:     "not-a-uuid",
				UserUuid:      uuid.NewString(),
				PaymentMethod: paymentV1.PaymentMethod_PAYMENT_METHOD_CARD,
			})

			Expect(err).To(HaveOccurred())
			Expect(status.Code(err)).To(Equal(codes.InvalidArgument))
		})

		It("rejects an unspecified payment method", func() {
			_, err := paymentClient.PayOrder(ctx, &paymentV1.PayOrderRequest{
				OrderUuid:     uuid.NewString(),
				UserUuid:      uuid.NewString(),
				PaymentMethod: paymentV1.PaymentMethod_PAYMENT_METHOD_UNSPECIFIED,
			})

			Expect(err).To(HaveOccurred())
			Expect(status.Code(err)).To(Equal(codes.InvalidArgument))
		})
	})

	Describe("Health", func() {
		It("reports the service as serving", func() {
			resp, err := grpc_health_v1.NewHealthClient(conn).Check(ctx, &grpc_health_v1.HealthCheckRequest{})

			Expect(err).ToNot(HaveOccurred())
			Expect(resp.GetStatus()).To(Equal(grpc_health_v1.HealthCheckResponse_SERVING))
		})
	})
})
