package app

import (
	"context"

	"google.golang.org/grpc"

	authV1 "inventory-go/shared/pkg/proto/auth/v1"
)

// localAuthClient lets the auth interceptor of this service check sessions in-process,
// instead of calling its own gRPC server over the network. Only Whoami is implemented:
// that is the only method the interceptor uses.
type localAuthClient struct {
	authV1.AuthServiceClient

	server authV1.AuthServiceServer
}

func (c localAuthClient) Whoami(ctx context.Context, req *authV1.WhoamiRequest, _ ...grpc.CallOption) (*authV1.WhoamiResponse, error) {
	return c.server.Whoami(ctx, req)
}
