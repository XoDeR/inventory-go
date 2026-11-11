package v1

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"inventory-go/order/internal/service/mocks"
	grpcAuth "inventory-go/platform/pkg/middleware/grpc"
)

type APISuite struct {
	suite.Suite

	orderService *mocks.OrderService

	// userUUID is the user the auth middleware would have put into the request context
	userUUID uuid.UUID

	api *api
}

// SetupTest runs before every test, so each test gets a fresh mock and a fresh api
// and no expectations or call counts leak between tests.
func (s *APISuite) SetupTest() {
	s.orderService = mocks.NewOrderService(s.T())
	s.userUUID = uuid.New()

	s.api = NewAPI(s.orderService)
}

// authCtx returns a context of an authenticated request.
func (s *APISuite) authCtx() context.Context {
	return grpcAuth.WithSession(context.Background(), &grpcAuth.User{UUID: s.userUUID.String()}, uuid.NewString())
}

func TestAPISuite(t *testing.T) {
	suite.Run(t, new(APISuite))
}
