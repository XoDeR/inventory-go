package v1

import (
	"context"

	"inventory-go/auth/internal/converter"
	"inventory-go/auth/internal/model"
	authV1 "inventory-go/shared/pkg/proto/auth/v1"
)

func (a *api) Register(ctx context.Context, req *authV1.RegisterRequest) (*authV1.RegisterResponse, error) {
	userUUID, err := a.authService.Register(ctx, model.RegisterInfo{
		Login:               req.GetLogin(),
		Password:            req.GetPassword(),
		Email:               req.GetEmail(),
		NotificationMethods: converter.NotificationMethodsFromProto(req.GetNotificationMethods()),
	})
	if err != nil {
		return nil, toStatus(err)
	}

	return &authV1.RegisterResponse{UserUuid: userUUID}, nil
}
