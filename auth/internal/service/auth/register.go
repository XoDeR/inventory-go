package auth

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"inventory-go/auth/internal/model"
)

func (s *service) Register(ctx context.Context, info model.RegisterInfo) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(info.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}

	user := model.User{
		UUID:                uuid.NewString(),
		Login:               info.Login,
		Email:               info.Email,
		PasswordHash:        string(hash),
		NotificationMethods: info.NotificationMethods,
	}

	if err = s.userRepository.Create(ctx, user); err != nil {
		return "", err
	}

	return user.UUID, nil
}
