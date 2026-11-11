package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	authV1 "inventory-go/shared/pkg/proto/auth/v1"
)

type fakeAuthClient struct {
	authV1.AuthServiceClient

	res *authV1.WhoamiResponse
	err error
}

func (f fakeAuthClient) Whoami(context.Context, *authV1.WhoamiRequest, ...grpc.CallOption) (*authV1.WhoamiResponse, error) {
	return f.res, f.err
}

func call(t *testing.T, client AuthClient, ctx context.Context) (*User, string, error) {
	t.Helper()

	var (
		user    *User
		session string
	)

	handler := func(ctx context.Context, _ any) (any, error) {
		user, _ = GetUserFromContext(ctx)
		session, _ = GetSessionUUIDFromContext(ctx)

		return nil, nil
	}

	_, err := NewAuthInterceptor(client).Unary()(ctx, nil, &grpc.UnaryServerInfo{}, handler)

	return user, session, err
}

func TestUnaryAuthenticates(t *testing.T) {
	client := fakeAuthClient{res: &authV1.WhoamiResponse{UserUuid: "u1", Login: "alice", Email: "a@x.io"}}
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(SessionUUIDMetadataKey, "s1"))

	user, session, err := call(t, client, ctx)

	require.NoError(t, err)
	require.Equal(t, &User{UUID: "u1", Login: "alice", Email: "a@x.io"}, user)
	require.Equal(t, "s1", session)
}

func TestUnaryRejects(t *testing.T) {
	withSession := metadata.NewIncomingContext(context.Background(), metadata.Pairs(SessionUUIDMetadataKey, "s1"))
	emptySession := metadata.NewIncomingContext(context.Background(), metadata.Pairs(SessionUUIDMetadataKey, ""))
	noSession := metadata.NewIncomingContext(context.Background(), metadata.MD{})

	tests := map[string]struct {
		ctx    context.Context
		client fakeAuthClient
	}{
		"no metadata":   {ctx: context.Background()},
		"no session":    {ctx: noSession},
		"empty session": {ctx: emptySession},
		"auth rejects":  {ctx: withSession, client: fakeAuthClient{err: errors.New("session not found")}},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, _, err := call(t, tt.client, tt.ctx)

			require.Equal(t, codes.Unauthenticated, status.Code(err))
		})
	}
}

func TestForwardSessionUUID(t *testing.T) {
	ctx := WithSession(context.Background(), &User{UUID: "u1"}, "s1")

	md, ok := metadata.FromOutgoingContext(ForwardSessionUUIDToGRPC(ctx))

	require.True(t, ok)
	require.Equal(t, []string{"s1"}, md.Get(SessionUUIDMetadataKey))

	_, ok = metadata.FromOutgoingContext(ForwardSessionUUIDToGRPC(context.Background()))
	require.False(t, ok)
}
