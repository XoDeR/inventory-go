package v1

import (
	"context"

	authV1 "inventory-go/shared/pkg/proto/auth/v1"
)

func (a *api) Login(ctx context.Context, req *authV1.LoginRequest) (*authV1.LoginResponse, error) {
	sessionUUID, err := a.authService.Login(ctx, req.GetLogin(), req.GetPassword())
	if err != nil {
		return nil, toStatus(err)
	}

	return &authV1.LoginResponse{SessionUuid: sessionUUID}, nil
}
