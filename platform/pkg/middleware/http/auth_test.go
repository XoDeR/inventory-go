package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

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

func serve(client AuthClient, header string) (*httptest.ResponseRecorder, bool) {
	reached := false

	next := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		reached = true

		user, ok := GetUserFromContext(r.Context())
		if !ok || user.UUID != "u1" {
			panic("user is not in the context")
		}
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if header != "" {
		req.Header.Set(SessionUUIDHeader, header)
	}

	rec := httptest.NewRecorder()
	NewAuthMiddleware(client).Handle(next).ServeHTTP(rec, req)

	return rec, reached
}

func TestHandleAuthenticates(t *testing.T) {
	client := fakeAuthClient{res: &authV1.WhoamiResponse{UserUuid: "u1", Login: "alice"}}

	rec, reached := serve(client, "s1")

	require.True(t, reached)
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestHandleMissingSession(t *testing.T) {
	rec, reached := serve(fakeAuthClient{}, "")

	require.False(t, reached)
	require.Equal(t, http.StatusUnauthorized, rec.Code)
	require.JSONEq(t, `{"error":{"code":"MISSING_SESSION","message":"Authentication required"}}`, rec.Body.String())
}

func TestHandleInvalidSession(t *testing.T) {
	rec, reached := serve(fakeAuthClient{err: errors.New("session not found")}, "s1")

	require.False(t, reached)
	require.Equal(t, http.StatusUnauthorized, rec.Code)
	require.JSONEq(t, `{"error":{"code":"INVALID_SESSION","message":"Authentication failed"}}`, rec.Body.String())
}
