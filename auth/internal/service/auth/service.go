package auth

import (
	"time"

	"inventory-go/auth/internal/repository"
	def "inventory-go/auth/internal/service"
)

var _ def.AuthService = (*service)(nil)

type service struct {
	userRepository    repository.UserRepository
	sessionRepository repository.SessionRepository
	sessionTTL        time.Duration
}

func NewService(
	userRepository repository.UserRepository,
	sessionRepository repository.SessionRepository,
	sessionTTL time.Duration,
) *service {
	return &service{
		userRepository:    userRepository,
		sessionRepository: sessionRepository,
		sessionTTL:        sessionTTL,
	}
}
