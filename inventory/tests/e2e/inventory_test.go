//go:build e2e

package e2e

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	inventoryV1 "inventory-go/shared/pkg/proto/inventory/v1"
)

// UUID of the first part of the catalog the service seeds on the first start
const seededCPUUUID = "6f1c8a4e-3b52-4d7a-9c10-1a2b3c4d5e01"

var _ = Describe("InventoryService", func() {
	var (
		ctx             context.Context
		cancel          context.CancelFunc
		conn            *grpc.ClientConn
		inventoryClient inventoryV1.InventoryServiceClient
	)

	BeforeEach(func() {
		ctx, cancel = context.WithCancel(suiteCtx)

		var err error
		conn, err = grpc.NewClient(
			env.App.Address(),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		Expect(err).ToNot(HaveOccurred(), "expected to connect to the gRPC app")

		inventoryClient = inventoryV1.NewInventoryServiceClient(conn)
	})

	AfterEach(func() {
		Expect(conn.Close()).To(Succeed())
		cancel()
	})

	Describe("GetPart", func() {
		It("returns a seeded part by UUID", func() {
			resp, err := inventoryClient.GetPart(ctx, &inventoryV1.GetPartRequest{Uuid: seededCPUUUID})

			Expect(err).ToNot(HaveOccurred())
			Expect(resp.GetPart()).ToNot(BeNil())
			Expect(resp.GetPart().GetUuid()).To(Equal(seededCPUUUID))
			Expect(resp.GetPart().GetCategory()).To(Equal(inventoryV1.Category_CATEGORY_CPU))
			Expect(resp.GetPart().GetName()).ToNot(BeEmpty())
			Expect(resp.GetPart().GetPrice()).To(BeNumerically(">", 0))
		})

		It("returns NotFound for an unknown UUID", func() {
			_, err := inventoryClient.GetPart(ctx, &inventoryV1.GetPartRequest{
				Uuid: "11111111-1111-4111-8111-111111111111",
			})

			Expect(err).To(HaveOccurred())
			Expect(status.Code(err)).To(Equal(codes.NotFound))
		})

		It("rejects a malformed UUID", func() {
			_, err := inventoryClient.GetPart(ctx, &inventoryV1.GetPartRequest{Uuid: "not-a-uuid"})

			Expect(err).To(HaveOccurred())
			Expect(status.Code(err)).To(Equal(codes.InvalidArgument))
		})

		It("returns a part inserted directly into MongoDB", func() {
			partUUID, err := env.InsertTestPart(ctx, "E2E part", int32(inventoryV1.Category_CATEGORY_GPU), "Norway")
			Expect(err).ToNot(HaveOccurred(), "expected the part to be inserted into MongoDB")
			DeferCleanup(func() {
				Expect(env.DeleteTestPart(context.Background(), partUUID)).To(Succeed())
			})

			resp, err := inventoryClient.GetPart(ctx, &inventoryV1.GetPartRequest{Uuid: partUUID})

			Expect(err).ToNot(HaveOccurred())
			Expect(resp.GetPart().GetName()).To(Equal("E2E part"))
			Expect(resp.GetPart().GetManufacturer().GetCountry()).To(Equal("Norway"))
		})
	})

	Describe("ListParts", func() {
		It("returns the whole seeded catalog without a filter", func() {
			resp, err := inventoryClient.ListParts(ctx, &inventoryV1.ListPartsRequest{})

			Expect(err).ToNot(HaveOccurred())
			Expect(len(resp.GetParts())).To(BeNumerically(">=", 4))
		})

		It("filters by category", func() {
			resp, err := inventoryClient.ListParts(ctx, &inventoryV1.ListPartsRequest{
				Filter: &inventoryV1.PartsFilter{
					Categories: []inventoryV1.Category{inventoryV1.Category_CATEGORY_CPU},
				},
			})

			Expect(err).ToNot(HaveOccurred())
			Expect(resp.GetParts()).ToNot(BeEmpty())
			for _, part := range resp.GetParts() {
				Expect(part.GetCategory()).To(Equal(inventoryV1.Category_CATEGORY_CPU))
			}
		})

		It("filters by UUIDs", func() {
			resp, err := inventoryClient.ListParts(ctx, &inventoryV1.ListPartsRequest{
				Filter: &inventoryV1.PartsFilter{Uuids: []string{seededCPUUUID}},
			})

			Expect(err).ToNot(HaveOccurred())
			Expect(resp.GetParts()).To(HaveLen(1))
			Expect(resp.GetParts()[0].GetUuid()).To(Equal(seededCPUUUID))
		})

		It("returns an empty list when nothing matches", func() {
			resp, err := inventoryClient.ListParts(ctx, &inventoryV1.ListPartsRequest{
				Filter: &inventoryV1.PartsFilter{Names: []string{"no such part"}},
			})

			Expect(err).ToNot(HaveOccurred())
			Expect(resp.GetParts()).To(BeEmpty())
		})
	})
})
