package part

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"inventory-go/inventory/internal/repository/mocks"
)

const testCacheTTL = time.Hour

type ServiceSuite struct {
	suite.Suite

	inventoryRepository      *mocks.InventoryRepository
	inventoryCacheRepository *mocks.InventoryCacheRepository

	service *service
}

// SetupTest runs before every test, so each test gets fresh mocks and a fresh service
// and no expectations or call counts leak between tests.
func (s *ServiceSuite) SetupTest() {
	s.inventoryRepository = mocks.NewInventoryRepository(s.T())
	s.inventoryCacheRepository = mocks.NewInventoryCacheRepository(s.T())

	s.service = NewService(s.inventoryRepository, s.inventoryCacheRepository, testCacheTTL)
}

func TestServiceSuite(t *testing.T) {
	suite.Run(t, new(ServiceSuite))
}
