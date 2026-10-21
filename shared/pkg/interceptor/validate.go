// Package interceptor contains shared gRPC interceptors.
package interceptor

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type validator interface {
	Validate() error
}

// UnaryValidate rejects requests that fail protoc-gen-validate rules
// with InvalidArgument before they reach the handler.
func UnaryValidate() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if v, ok := req.(validator); ok {
			if err := v.Validate(); err != nil {
				return nil, status.Error(codes.InvalidArgument, err.Error())
			}
		}

		return handler(ctx, req)
	}
}
