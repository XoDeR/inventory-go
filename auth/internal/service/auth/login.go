package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"inventory-go/auth/internal/model"
)

func (s *service) Login(ctx context.Context, login, password string) (string, error) {
	user, err := s.userRepository.GetByLogin(ctx, login)
	if err != nil {
		// An unknown login and a wrong password look the same to the caller.
		if errors.Is(err, model.ErrUserNotFound) {
			return "", model.ErrInvalidCredentials
		}

		return "", err
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", model.ErrInvalidCredentials
	}

	sessionUUID := uuid.NewString()

	if err = s.sessionRepository.Create(ctx, sessionUUID, user.UUID, s.sessionTTL); err != nil {
		return "", fmt.Errorf("create session: %w", err)
	}

	return sessionUUID, nil
}
