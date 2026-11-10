package service

import (
	"context"

	"inventory-go/auth/internal/model"
)

type AuthService interface {
	// Register creates a user and returns its UUID.
	Register(ctx context.Context, info model.RegisterInfo) (string, error)
	// Login checks the credentials, opens a session and returns its UUID.
	Login(ctx context.Context, login, password string) (string, error)
	// Whoami returns the owner of the session.
	Whoami(ctx context.Context, sessionUUID string) (model.User, error)
	// GetUser returns a user by UUID.
	GetUser(ctx context.Context, userUUID string) (model.User, error)
}
