package auth

import (
	"context"

	"inventory-go/auth/internal/model"
)

func (s *service) GetUser(ctx context.Context, userUUID string) (model.User, error) {
	return s.userRepository.GetByUUID(ctx, userUUID)
}
