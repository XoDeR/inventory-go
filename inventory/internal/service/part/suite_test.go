package part

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"inventory-go/inventory/internal/repository/mocks"
)

type ServiceSuite struct {
	suite.Suite

	inventoryRepository *mocks.InventoryRepository

	service *service
}

// SetupTest runs before every test, so each test gets a fresh mock and a fresh service
// and no expectations or call counts leak between tests.
func (s *ServiceSuite) SetupTest() {
	s.inventoryRepository = mocks.NewInventoryRepository(s.T())

	s.service = NewService(s.inventoryRepository)
}

func TestServiceSuite(t *testing.T) {
	suite.Run(t, new(ServiceSuite))
}
