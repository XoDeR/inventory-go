package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"inventory-go/auth/internal/repository/mocks"
)

const testSessionTTL = 24 * time.Hour

type ServiceSuite struct {
	suite.Suite

	userRepository    *mocks.UserRepository
	sessionRepository *mocks.SessionRepository

	service *service
}

func (s *ServiceSuite) SetupTest() {
	s.userRepository = mocks.NewUserRepository(s.T())
	s.sessionRepository = mocks.NewSessionRepository(s.T())

	s.service = NewService(s.userRepository, s.sessionRepository, testSessionTTL)
}

func TestServiceSuite(t *testing.T) {
	suite.Run(t, new(ServiceSuite))
}
