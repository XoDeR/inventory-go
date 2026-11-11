// Package authstack starts the auth service with its PostgreSQL and Redis for e2e tests of the
// services that sit behind the auth middleware, and hands out ready-to-use user sessions.
package authstack

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go/wait"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"inventory-go/platform/pkg/logger"
	"inventory-go/platform/pkg/testcontainers/app"
	"inventory-go/platform/pkg/testcontainers/path"
	"inventory-go/platform/pkg/testcontainers/postgres"
	"inventory-go/platform/pkg/testcontainers/redis"
	authV1 "inventory-go/shared/pkg/proto/auth/v1"
)

const (
	// GRPCPort is the port the auth service listens on inside its container.
	GRPCPort = "50053"

	dockerfile = "deploy/docker/Dockerfile"

	postgresImage    = "postgres:17-alpine"
	postgresDatabase = "auth-service"
	postgresUser     = "auth-service-user"
	postgresPassword = "auth-service-password"
	postgresPort     = "5432"

	redisImage = "redis:7.2.5-alpine3.20"
	redisPort  = "6379"

	startupTimeout = 3 * time.Minute
	callTimeout    = 10 * time.Second
	testPassword   = "e2e-password-1"
)

// Stack is the running auth service with its storages.
type Stack struct {
	Postgres *postgres.Container
	Redis    *redis.Container
	App      *app.Container

	// Host is the name the auth service has inside the Docker network.
	Host string
}

// User is a registered user with an open session.
type User struct {
	UUID        string
	SessionUUID string
}

// Start builds and starts the auth service in the given network. prefix keeps the container
// names unique between test suites. On a failure everything started so far is stopped.
func Start(ctx context.Context, networkName, prefix string) (*Stack, error) {
	s := &Stack{Host: prefix + "-auth-app"}

	var err error

	s.Postgres, err = postgres.NewContainer(ctx,
		postgres.WithNetworkName(networkName),
		postgres.WithContainerName(prefix+"-auth-postgres"),
		postgres.WithImageName(postgresImage),
		postgres.WithDatabase(postgresDatabase),
		postgres.WithAuth(postgresUser, postgresPassword),
		postgres.WithLogger(logger.Default()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start the auth PostgreSQL: %w", err)
	}

	s.Redis, err = redis.NewContainer(ctx,
		redis.WithNetworkName(networkName),
		redis.WithContainerName(prefix+"-auth-redis"),
		redis.WithImageName(redisImage),
		redis.WithLogger(logger.Default()),
	)
	if err != nil {
		s.Terminate(ctx)
		return nil, fmt.Errorf("failed to start the auth Redis: %w", err)
	}

	s.App, err = app.NewContainer(ctx,
		app.WithName(s.Host),
		app.WithPort(GRPCPort),
		app.WithDockerfile(path.GetProjectRoot(), dockerfile),
		app.WithBuildArg("SERVICE", "auth"),
		app.WithBuildArg("CMD", "grpc_server"),
		app.WithNetwork(networkName),
		app.WithEnv(map[string]string{
			"GRPC_HOST": "0.0.0.0",
			"GRPC_PORT": GRPCPort,

			"LOGGER_LEVEL":   "debug",
			"LOGGER_AS_JSON": "true",

			"POSTGRES_HOST":     prefix + "-auth-postgres",
			"POSTGRES_PORT":     postgresPort,
			"POSTGRES_DB":       postgresDatabase,
			"POSTGRES_USER":     postgresUser,
			"POSTGRES_PASSWORD": postgresPassword,

			"REDIS_HOST":               prefix + "-auth-redis",
			"REDIS_PORT":               redisPort,
			"REDIS_CONNECTION_TIMEOUT": "5s",
			"REDIS_MAX_IDLE":           "10",
			"REDIS_IDLE_TIMEOUT":       "240s",

			"SESSION_TTL": "24h",
		}),
		app.WithLogOutput(os.Stdout),
		app.WithStartupWait(wait.ForListeningPort(GRPCPort+"/tcp").WithStartupTimeout(startupTimeout)),
		app.WithLogger(logger.Default()),
	)
	if err != nil {
		s.Terminate(ctx)
		return nil, fmt.Errorf("failed to start the auth service: %w", err)
	}

	return s, nil
}

// NewUser registers a new random user and logs them in.
func (s *Stack) NewUser(ctx context.Context) (User, error) {
	conn, err := grpc.NewClient(s.App.Address(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return User{}, fmt.Errorf("failed to connect to the auth service: %w", err)
	}
	defer func() { _ = conn.Close() }()

	ctx, cancel := context.WithTimeout(ctx, callTimeout)
	defer cancel()

	client := authV1.NewAuthServiceClient(conn)
	login := "user-" + uuid.NewString()[:8]

	registered, err := client.Register(ctx, &authV1.RegisterRequest{
		Login:    login,
		Password: testPassword,
		Email:    login + "@example.com",
	})
	if err != nil {
		return User{}, fmt.Errorf("failed to register: %w", err)
	}

	session, err := client.Login(ctx, &authV1.LoginRequest{Login: login, Password: testPassword})
	if err != nil {
		return User{}, fmt.Errorf("failed to log in: %w", err)
	}

	return User{UUID: registered.GetUserUuid(), SessionUUID: session.GetSessionUuid()}, nil
}

// Terminate stops everything that was started. Errors are logged, not returned.
func (s *Stack) Terminate(ctx context.Context) {
	if s.App != nil {
		if err := s.App.Terminate(ctx); err != nil {
			logger.Error(ctx, "failed to stop the auth service", "error", err)
		}
	}

	if s.Redis != nil {
		if err := s.Redis.Terminate(ctx); err != nil {
			logger.Error(ctx, "failed to stop the auth Redis", "error", err)
		}
	}

	if s.Postgres != nil {
		if err := s.Postgres.Terminate(ctx); err != nil {
			logger.Error(ctx, "failed to stop the auth PostgreSQL", "error", err)
		}
	}
}
