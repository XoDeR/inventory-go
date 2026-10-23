package v1

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"inventory-go/inventory/internal/service/mocks"
)

type APISuite struct {
	suite.Suite

	inventoryService *mocks.InventoryService

	api *api
}

// SetupTest runs before every test, so each test gets a fresh mock and a fresh api
// and no expectations or call counts leak between tests.
func (s *APISuite) SetupTest() {
	s.inventoryService = mocks.NewInventoryService(s.T())

	s.api = NewAPI(s.inventoryService)
}

func TestAPISuite(t *testing.T) {
	suite.Run(t, new(APISuite))
}
