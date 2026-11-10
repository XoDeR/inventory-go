package auth

import (
	"context"

	"inventory-go/auth/internal/model"
)

func (s *service) Whoami(ctx context.Context, sessionUUID string) (model.User, error) {
	userUUID, err := s.sessionRepository.Get(ctx, sessionUUID)
	if err != nil {
		return model.User{}, err
	}

	return s.userRepository.GetByUUID(ctx, userUUID)
}
