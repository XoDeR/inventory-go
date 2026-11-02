package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/stdlib"

	"inventory-go/order/internal/config"
	"inventory-go/order/internal/migrator"
	"inventory-go/order/migrations"
	"inventory-go/platform/pkg/closer"
	"inventory-go/platform/pkg/logger"
	orderV1 "inventory-go/shared/pkg/openapi/order/v1"
)

const (
	readHeaderTimeout = 5 * time.Second
	requestTimeout    = 10 * time.Second
)

type App struct {
	diContainer *diContainer
	httpServer  *http.Server
}

func New(ctx context.Context) (*App, error) {
	a := &App{}

	err := a.initDeps(ctx)
	if err != nil {
		return nil, err
	}

	return a, nil
}

// Run serves HTTP until the server fails or ctx is cancelled.
func (a *App) Run(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() { errCh <- a.runHTTPServer(ctx) }()

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
		a.initHTTPServer,
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

func (a *App) initHTTPServer(ctx context.Context) error {
	orderServer, err := orderV1.NewServer(a.diContainer.OrderV1API(ctx))
	if err != nil {
		return fmt.Errorf("failed to create the OpenAPI server: %w", err)
	}

	r := chi.NewRouter()

	r.Use(middleware.Logger) // log like: PUT http://localhost:8080/api/v1/order/3c516345-47df-4eba-a9d6-da1e3d80860b HTTP/1.1" from 127.0.0.
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(requestTimeout))

	r.Mount("/", orderServer)

	a.httpServer = &http.Server{
		Addr:              config.AppConfig().HTTP.Address(),
		Handler:           r,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	closer.AddNamed("HTTP server", func(ctx context.Context) error {
		return a.httpServer.Shutdown(ctx)
	})

	return nil
}

func (a *App) runHTTPServer(ctx context.Context) error {
	logger.Info(ctx, "HTTP server running", "address", config.AppConfig().HTTP.Address())

	err := a.httpServer.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}
