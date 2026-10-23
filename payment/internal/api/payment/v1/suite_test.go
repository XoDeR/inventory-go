package v1

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"inventory-go/payment/internal/service/mocks"
)

type APISuite struct {
	suite.Suite

	paymentService *mocks.PaymentService

	api *api
}

// SetupTest runs before every test, so each test gets a fresh mock and a fresh api
// and no expectations or call counts leak between tests.
func (s *APISuite) SetupTest() {
	s.paymentService = mocks.NewPaymentService(s.T())

	s.api = NewAPI(s.paymentService)
}

func TestAPISuite(t *testing.T) {
	suite.Run(t, new(APISuite))
}
