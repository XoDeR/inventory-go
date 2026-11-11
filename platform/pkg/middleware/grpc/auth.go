// Package grpc contains the gRPC authentication interceptor. It validates the session
// with AuthService.Whoami and puts the user into the request context.
package grpc

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	authV1 "inventory-go/shared/pkg/proto/auth/v1"
)

const (
	// SessionUUIDMetadataKey is the key for passing the session UUID in gRPC metadata
	SessionUUIDMetadataKey = "session-uuid"
)

type contextKey string

const (
	// userContextKey is the key for storing the user in the context
	userContextKey contextKey = "user"
	// sessionUUIDContextKey is the key for storing the session UUID in the context
	sessionUUIDContextKey contextKey = "session-uuid"
)

// User is the authenticated user, taken from the Whoami response.
type User struct {
	UUID  string
	Login string
	Email string
}

// AuthClient is an alias for the generated gRPC client of AuthService
type AuthClient = authV1.AuthServiceClient

// AuthInterceptor is an interceptor for authenticating gRPC requests
type AuthInterceptor struct {
	authClient AuthClient
}

// NewAuthInterceptor creates a new authentication interceptor
func NewAuthInterceptor(authClient AuthClient) *AuthInterceptor {
	return &AuthInterceptor{authClient: authClient}
}

// Unary returns a unary server interceptor for authentication
func (i *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		_ *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		authCtx, err := i.authenticate(ctx)
		if err != nil {
			return nil, err
		}

		return handler(authCtx, req)
	}
}

// authenticate performs authentication and adds the user to the context
func (i *AuthInterceptor) authenticate(ctx context.Context) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing metadata")
	}

	sessionUUIDs := md.Get(SessionUUIDMetadataKey)
	if len(sessionUUIDs) == 0 {
		return nil, status.Error(codes.Unauthenticated, "missing session-uuid in metadata")
	}

	sessionUUID := sessionUUIDs[0]
	if sessionUUID == "" {
		return nil, status.Error(codes.Unauthenticated, "empty session-uuid")
	}

	// Validate the session via the auth service
	whoamiRes, err := i.authClient.Whoami(ctx, &authV1.WhoamiRequest{SessionUuid: sessionUUID})
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, fmt.Sprintf("invalid session: %v", err))
	}

	user := &User{
		UUID:  whoamiRes.GetUserUuid(),
		Login: whoamiRes.GetLogin(),
		Email: whoamiRes.GetEmail(),
	}

	return WithSession(ctx, user, sessionUUID), nil
}

// WithSession puts the user and the session UUID into the context.
func WithSession(ctx context.Context, user *User, sessionUUID string) context.Context {
	ctx = context.WithValue(ctx, userContextKey, user)
	return context.WithValue(ctx, sessionUUIDContextKey, sessionUUID)
}

// GetUserFromContext extracts the user from the context
func GetUserFromContext(ctx context.Context) (*User, bool) {
	user, ok := ctx.Value(userContextKey).(*User)
	return user, ok
}

// GetSessionUUIDFromContext extracts the session UUID from the context
func GetSessionUUIDFromContext(ctx context.Context) (string, bool) {
	sessionUUID, ok := ctx.Value(sessionUUIDContextKey).(string)
	return sessionUUID, ok
}

// ForwardSessionUUIDToGRPC adds the session UUID from the context to the outgoing gRPC metadata,
// so a service can pass the caller's session on to the services it calls.
func ForwardSessionUUIDToGRPC(ctx context.Context) context.Context {
	sessionUUID, ok := GetSessionUUIDFromContext(ctx)
	if !ok || sessionUUID == "" {
		return ctx
	}

	return metadata.AppendToOutgoingContext(ctx, SessionUUIDMetadataKey, sessionUUID)
}
