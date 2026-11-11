package app

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	"inventory-go/payment/internal/config"
	"inventory-go/platform/pkg/closer"
	"inventory-go/platform/pkg/grpc/health"
	"inventory-go/platform/pkg/logger"
	grpcAuth "inventory-go/platform/pkg/middleware/grpc"
	httpAuth "inventory-go/platform/pkg/middleware/http"
	"inventory-go/shared/api/swagger"
	"inventory-go/shared/pkg/interceptor"
	paymentV1 "inventory-go/shared/pkg/proto/payment/v1"
)

const readHeaderTimeout = 10 * time.Second

type App struct {
	diContainer *diContainer
	grpcServer  *grpc.Server
	httpServer  *http.Server
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

// Run serves gRPC and the HTTP gateway until one of them fails or ctx is cancelled.
func (a *App) Run(ctx context.Context) error {
	errCh := make(chan error, 2)

	go func() { errCh <- a.runGRPCServer(ctx) }()
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
		a.initListener,
		a.initGRPCServer,
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
	a.grpcServer = grpc.NewServer(grpc.ChainUnaryInterceptor(
		grpcAuth.NewAuthInterceptor(a.diContainer.AuthClient(ctx)).Unary(),
		interceptor.UnaryValidate(),
	))

	closer.AddNamed("gRPC server", func(_ context.Context) error {
		a.grpcServer.GracefulStop()
		return nil
	})

	reflection.Register(a.grpcServer)

	// the health service lets orchestrators probe the server
	health.RegisterService(a.grpcServer)

	paymentV1.RegisterPaymentServiceServer(a.grpcServer, a.diContainer.PaymentV1API(ctx))

	return nil
}

// initHTTPServer builds the HTTP gateway (REST -> gRPC) with the Swagger UI.
func (a *App) initHTTPServer(ctx context.Context) error {
	gwMux := runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(grpcAuth.GatewayHeaderMatcher(httpAuth.SessionUUIDHeader)),
	)

	err := paymentV1.RegisterPaymentServiceHandlerFromEndpoint(
		ctx,
		gwMux,
		config.AppConfig().GRPC.Address(),
		[]grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())},
	)
	if err != nil {
		return err
	}

	httpMux := http.NewServeMux()
	httpMux.Handle("/api/", gwMux)
	httpMux.Handle("/", swagger.Handler("payment/v1/payment.swagger.json"))

	a.httpServer = &http.Server{
		Addr:              config.AppConfig().HTTP.Address(),
		Handler:           httpMux,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	closer.AddNamed("HTTP server", func(ctx context.Context) error {
		return a.httpServer.Shutdown(ctx)
	})

	return nil
}

func (a *App) runGRPCServer(ctx context.Context) error {
	logger.Info(ctx, "gRPC server listening", "address", config.AppConfig().GRPC.Address())

	return a.grpcServer.Serve(a.listener)
}

func (a *App) runHTTPServer(ctx context.Context) error {
	logger.Info(ctx, "HTTP gateway listening", "address", config.AppConfig().HTTP.Address(), "swagger_ui", "/swagger-ui.html")

	err := a.httpServer.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}
