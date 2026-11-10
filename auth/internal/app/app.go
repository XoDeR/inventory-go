package app

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/jackc/pgx/v5/stdlib"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"inventory-go/auth/internal/config"
	"inventory-go/auth/internal/migrator"
	"inventory-go/auth/migrations"
	"inventory-go/platform/pkg/closer"
	"inventory-go/platform/pkg/grpc/health"
	"inventory-go/platform/pkg/logger"
	"inventory-go/shared/pkg/interceptor"
	authV1 "inventory-go/shared/pkg/proto/auth/v1"
)

type App struct {
	diContainer *diContainer
	grpcServer  *grpc.Server
	listener    net.Listener
}

func New(ctx context.Context) (*App, error) {
	a := &App{}

	err := a.initDeps(ctx)
	if err != nil {
		return nil, err
	}

	return a, nil
}

// Run serves gRPC until it fails or ctx is cancelled.
func (a *App) Run(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() { errCh <- a.runGRPCServer(ctx) }()

	select {
	case <-ctx.Done():
		return nil
	case err := <-errCh:
		return err
	}
}

func (a *App) initDeps(ctx context.Context) error {
	inits := []func(context.Context) error{
		a.initDI,
		a.initLogger,
		a.initCloser,
		a.runMigrations,
		a.initListener,
		a.initGRPCServer,
	}

	for _, f := range inits {
		err := f(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *App) initDI(_ context.Context) error {
	a.diContainer = NewDiContainer()
	return nil
}

func (a *App) initLogger(_ context.Context) error {
	return logger.Init(
		config.AppConfig().Logger.Level(),
		config.AppConfig().Logger.AsJSON(),
	)
}

func (a *App) initCloser(_ context.Context) error {
	closer.SetLogger(logger.Default())
	return nil
}

// runMigrations applies the pending SQL migrations before the service starts serving.
func (a *App) runMigrations(ctx context.Context) error {
	// goose works on top of database/sql, so run migrations through a stdlib wrapper of the pool
	migrationsDB := stdlib.OpenDBFromPool(a.diContainer.PostgresPool(ctx))

	err := migrator.NewMigrator(migrationsDB).Up(ctx, migrations.FS)
	if err != nil {
		_ = migrationsDB.Close()
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	if err = migrationsDB.Close(); err != nil {
		logger.Error(ctx, "failed to close migrations connection", "error", err)
	}

	return nil
}

func (a *App) initListener(_ context.Context) error {
	listener, err := net.Listen("tcp", config.AppConfig().GRPC.Address())
	if err != nil {
		return err
	}

	closer.AddNamed("TCP listener", func(_ context.Context) error {
		lerr := listener.Close()
		if lerr != nil && !errors.Is(lerr, net.ErrClosed) {
			return lerr
		}

		return nil
	})

	a.listener = listener

	return nil
}

func (a *App) initGRPCServer(ctx context.Context) error {
	a.grpcServer = grpc.NewServer(grpc.UnaryInterceptor(interceptor.UnaryValidate()))

	closer.AddNamed("gRPC server", func(_ context.Context) error {
		a.grpcServer.GracefulStop()
		return nil
	})

	reflection.Register(a.grpcServer)

	// the health service lets orchestrators probe the server
	health.RegisterService(a.grpcServer)

	authV1.RegisterAuthServiceServer(a.grpcServer, a.diContainer.AuthV1API(ctx))

	return nil
}

func (a *App) runGRPCServer(ctx context.Context) error {
	logger.Info(ctx, "gRPC server listening", "address", config.AppConfig().GRPC.Address())

	return a.grpcServer.Serve(a.listener)
}
