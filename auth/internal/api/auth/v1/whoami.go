package v1

import (
	"context"

	authV1 "inventory-go/shared/pkg/proto/auth/v1"
)

func (a *api) Whoami(ctx context.Context, req *authV1.WhoamiRequest) (*authV1.WhoamiResponse, error) {
	user, err := a.authService.Whoami(ctx, req.GetSessionUuid())
	if err != nil {
		return nil, toStatus(err)
	}

	return &authV1.WhoamiResponse{
		UserUuid: user.UUID,
		Login:    user.Login,
		Email:    user.Email,
	}, nil
}
