package v1

import (
	"context"

	"inventory-go/auth/internal/converter"
	authV1 "inventory-go/shared/pkg/proto/auth/v1"
)

func (a *api) GetUser(ctx context.Context, req *authV1.GetUserRequest) (*authV1.GetUserResponse, error) {
	user, err := a.authService.GetUser(ctx, req.GetUserUuid())
	if err != nil {
		return nil, toStatus(err)
	}

	return &authV1.GetUserResponse{
		UserUuid:            user.UUID,
		Login:               user.Login,
		Email:               user.Email,
		NotificationMethods: converter.NotificationMethodsToProto(user.NotificationMethods),
	}, nil
}
