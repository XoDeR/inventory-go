package repository

import (
	"context"
	"time"

	"inventory-go/auth/internal/model"
)

type UserRepository interface {
	// Create returns model.ErrUserAlreadyExists when the login or the email is taken.
	Create(ctx context.Context, user model.User) error
	// GetByUUID returns model.ErrUserNotFound when there is no such user.
	GetByUUID(ctx context.Context, uuid string) (model.User, error)
	// GetByLogin returns model.ErrUserNotFound when there is no such user.
	GetByLogin(ctx context.Context, login string) (model.User, error)
}

// SessionRepository stores user sessions (session UUID -> user UUID) with a TTL.
type SessionRepository interface {
	Create(ctx context.Context, sessionUUID, userUUID string, ttl time.Duration) error
	// Get returns the user UUID of the session, or model.ErrSessionNotFound.
	Get(ctx context.Context, sessionUUID string) (string, error)
}
