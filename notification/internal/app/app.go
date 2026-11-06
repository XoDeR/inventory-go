package app

import (
	"context"
	"errors"

	"inventory-go/notification/internal/config"
	"inventory-go/platform/pkg/closer"
	"inventory-go/platform/pkg/logger"
)

type App struct {
	diContainer *diContainer
}

func New(ctx context.Context) (*App, error) {
	a := &App{}

	err := a.initDeps(ctx)
	if err != nil {
		return nil, err
	}

	return a, nil
}

// Run consumes OrderPaid and ComputerAssembled events until a consumer fails or ctx is cancelled.
func (a *App) Run(ctx context.Context) error {
	errCh := make(chan error, 2)

	go func() { errCh <- a.diContainer.OrderPaidConsumer().Run(ctx) }()
	go func() { errCh <- a.diContainer.ComputerAssembledConsumer().Run(ctx) }()

	select {
	case <-ctx.Done():
		return nil
	case err := <-errCh:
		if err != nil && !errors.Is(err, context.Canceled) {
			return err
		}

		return nil
	}
}

func (a *App) initDeps(ctx context.Context) error {
	inits := []func(context.Context) error{
		a.initDI,
		a.initLogger,
		a.initCloser,
		a.initKafka,
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

// initKafka builds the consumers eagerly, so a misconfiguration fails the startup.
func (a *App) initKafka(_ context.Context) error {
	a.diContainer.OrderPaidConsumer()
	a.diContainer.ComputerAssembledConsumer()

	return nil
}
