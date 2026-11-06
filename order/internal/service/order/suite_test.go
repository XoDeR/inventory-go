package order

import (
	"testing"

	"github.com/stretchr/testify/suite"

	grpcMocks "inventory-go/order/internal/client/grpc/mocks"
	repoMocks "inventory-go/order/internal/repository/mocks"
	serviceMocks "inventory-go/order/internal/service/mocks"
)

type ServiceSuite struct {
	suite.Suite

	orderRepository *repoMocks.OrderRepository

	inventoryClient *grpcMocks.InventoryClient
	paymentClient   *grpcMocks.PaymentClient
	orderProducer   *serviceMocks.OrderProducerService

	service *service
}

// SetupTest runs before every test, so each test gets fresh mocks and a fresh service
// and no expectations or call counts leak between tests.
func (s *ServiceSuite) SetupTest() {
	s.orderRepository = repoMocks.NewOrderRepository(s.T())

	s.inventoryClient = grpcMocks.NewInventoryClient(s.T())
	s.paymentClient = grpcMocks.NewPaymentClient(s.T())
	s.orderProducer = serviceMocks.NewOrderProducerService(s.T())

	s.service = NewService(s.orderRepository, s.inventoryClient, s.paymentClient, s.orderProducer)
}

func TestServiceSuite(t *testing.T) {
	suite.Run(t, new(ServiceSuite))
}
