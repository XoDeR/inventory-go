package app

import (
	"context"
	"fmt"

	redigo "github.com/gomodule/redigo/redis"
	"github.com/jackc/pgx/v5/pgxpool"

	authV1API "inventory-go/auth/internal/api/auth/v1"
	"inventory-go/auth/internal/config"
	"inventory-go/auth/internal/repository"
	sessionRepository "inventory-go/auth/internal/repository/session"
	userRepository "inventory-go/auth/internal/repository/user"
	"inventory-go/auth/internal/service"
	authService "inventory-go/auth/internal/service/auth"
	"inventory-go/platform/pkg/cache"
	cacheRedis "inventory-go/platform/pkg/cache/redis"
	"inventory-go/platform/pkg/closer"
	"inventory-go/platform/pkg/logger"
	authV1 "inventory-go/shared/pkg/proto/auth/v1"
)

// diContainer lazily builds the dependencies of the service and keeps a single instance of each.
// Constructors panic on a failure: the container is only used while the app is being assembled.
type diContainer struct {
	authV1API authV1.AuthServiceServer

	authService service.AuthService

	userRepository    repository.UserRepository
	sessionRepository repository.SessionRepository

	postgresPool *pgxpool.Pool

	redisPool   *redigo.Pool
	redisClient cache.RedisClient
}

func NewDiContainer() *diContainer {
	return &diContainer{}
}

func (d *diContainer) AuthV1API(ctx context.Context) authV1.AuthServiceServer {
	if d.authV1API == nil {
		d.authV1API = authV1API.NewAPI(d.AuthService(ctx))
	}

	return d.authV1API
}

func (d *diContainer) AuthService(ctx context.Context) service.AuthService {
	if d.authService == nil {
		d.authService = authService.NewService(
			d.UserRepository(ctx),
			d.SessionRepository(ctx),
			config.AppConfig().Session.TTL(),
		)
	}

	return d.authService
}

func (d *diContainer) UserRepository(ctx context.Context) repository.UserRepository {
	if d.userRepository == nil {
		d.userRepository = userRepository.NewRepository(d.PostgresPool(ctx))
	}

	return d.userRepository
}

func (d *diContainer) SessionRepository(ctx context.Context) repository.SessionRepository {
	if d.sessionRepository == nil {
		d.sessionRepository = sessionRepository.NewRepository(d.RedisClient(ctx))
	}

	return d.sessionRepository
}

func (d *diContainer) PostgresPool(ctx context.Context) *pgxpool.Pool {
	if d.postgresPool == nil {
		pool, err := pgxpool.New(ctx, config.AppConfig().Postgres.URI())
		if err != nil {
			panic(fmt.Sprintf("failed to connect to Postgres: %v", err))
		}

		if err = pool.Ping(ctx); err != nil {
			pool.Close()
			panic(fmt.Sprintf("failed to ping Postgres: %v", err))
		}

		closer.AddNamed("Postgres pool", func(_ context.Context) error {
			pool.Close()
			return nil
		})

		d.postgresPool = pool
	}

	return d.postgresPool
}

func (d *diContainer) RedisPool() *redigo.Pool {
	if d.redisPool == nil {
		cfg := config.AppConfig().Redis

		pool := cacheRedis.NewPool(cfg.Address(), cfg.MaxIdle(), cfg.IdleTimeout(), cfg.ConnectionTimeout())

		closer.AddNamed("Redis pool", func(_ context.Context) error {
			return pool.Close()
		})

		d.redisPool = pool
	}

	return d.redisPool
}

// RedisClient does not ping Redis on start: connections are dialed lazily. Sessions live only
// in Redis, so Login and Whoami fail while it is down, but Register and GetUser keep working.
func (d *diContainer) RedisClient(_ context.Context) cache.RedisClient {
	if d.redisClient == nil {
		d.redisClient = cacheRedis.NewClient(d.RedisPool(), logger.Default(), config.AppConfig().Redis.ConnectionTimeout())
	}

	return d.redisClient
}
