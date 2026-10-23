package v1

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"inventory-go/order/internal/service/mocks"
)

type APISuite struct {
	suite.Suite

	orderService *mocks.OrderService

	api *api
}

// SetupTest runs before every test, so each test gets a fresh mock and a fresh api
// and no expectations or call counts leak between tests.
func (s *APISuite) SetupTest() {
	s.orderService = mocks.NewOrderService(s.T())

	s.api = NewAPI(s.orderService)
}

func TestAPISuite(t *testing.T) {
	suite.Run(t, new(APISuite))
}
