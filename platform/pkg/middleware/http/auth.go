// Package http contains the HTTP authentication middleware. It validates the session
// with AuthService.Whoami and puts the user into the request context.
package http

import (
	"context"
	"net/http"

	grpcAuth "inventory-go/platform/pkg/middleware/grpc"
	authV1 "inventory-go/shared/pkg/proto/auth/v1"
)

const SessionUUIDHeader = "X-Session-Uuid"

// AuthClient is an alias for the generated gRPC client of AuthService
type AuthClient = authV1.AuthServiceClient

// AuthMiddleware is a middleware for authenticating HTTP requests
type AuthMiddleware struct {
	authClient AuthClient
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(authClient AuthClient) *AuthMiddleware {
	return &AuthMiddleware{authClient: authClient}
}

// Handle handles an HTTP request with authentication
func (m *AuthMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionUUID := r.Header.Get(SessionUUIDHeader)
		if sessionUUID == "" {
			writeErrorResponse(w, http.StatusUnauthorized, "MISSING_SESSION", "Authentication required")
			return
		}

		// Validate the session via the auth service
		whoamiRes, err := m.authClient.Whoami(r.Context(), &authV1.WhoamiRequest{SessionUuid: sessionUUID})
		if err != nil {
			writeErrorResponse(w, http.StatusUnauthorized, "INVALID_SESSION", "Authentication failed")
			return
		}

		user := &grpcAuth.User{
			UUID:  whoamiRes.GetUserUuid(),
			Login: whoamiRes.GetLogin(),
			Email: whoamiRes.GetEmail(),
		}

		ctx := grpcAuth.WithSession(r.Context(), user, sessionUUID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserFromContext extracts the user from the context
func GetUserFromContext(ctx context.Context) (*grpcAuth.User, bool) {
	return grpcAuth.GetUserFromContext(ctx)
}

// GetSessionUUIDFromContext extracts the session UUID from the context
func GetSessionUUIDFromContext(ctx context.Context) (string, bool) {
	return grpcAuth.GetSessionUUIDFromContext(ctx)
}
