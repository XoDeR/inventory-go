package v1

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"inventory-go/auth/internal/model"
	"inventory-go/auth/internal/service"
	authV1 "inventory-go/shared/pkg/proto/auth/v1"
)

type api struct {
	authV1.UnimplementedAuthServiceServer

	authService service.AuthService
}

func NewAPI(authService service.AuthService) *api {
	return &api{authService: authService}
}

// toStatus maps domain errors to gRPC statuses. Unknown errors are returned as is.
func toStatus(err error) error {
	switch {
	case errors.Is(err, model.ErrUserAlreadyExists):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, model.ErrUserNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, model.ErrInvalidCredentials), errors.Is(err, model.ErrSessionNotFound):
		return status.Error(codes.Unauthenticated, err.Error())
	default:
		return err
	}
}
