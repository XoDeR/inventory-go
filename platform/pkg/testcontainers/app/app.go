package app

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"inventory-go/platform/pkg/logger"
)

const (
	defaultAppName        = "app"
	defaultAppPort        = "50051"
	defaultStartupTimeout = 1 * time.Minute
)

type Logger interface {
	Info(ctx context.Context, msg string, args ...any)
	Error(ctx context.Context, msg string, args ...any)
}

type Config struct {
	Name          string
	DockerfileDir string
	Dockerfile    string
	BuildArgs     map[string]*string
	Port          string
	Env           map[string]string
	Networks      []string
	LogOutput     io.Writer
	StartupWait   wait.Strategy
	Logger        Logger
}

// Container is a service built from a Dockerfile and started in Docker.
type Container struct {
	container    testcontainers.Container
	externalHost string
	externalPort string
	cfg          *Config
}

func NewContainer(ctx context.Context, opts ...Option) (*Container, error) {
	cfg := &Config{
		Name:          defaultAppName,
		Port:          defaultAppPort,
		Dockerfile:    "Dockerfile",
		DockerfileDir: ".",
		BuildArgs:     make(map[string]*string),
		LogOutput:     io.Discard,
		StartupWait:   wait.ForListeningPort(defaultAppPort + "/tcp").WithStartupTimeout(defaultStartupTimeout),
		Env:           make(map[string]string),
		Logger:        logger.NoopLogger{},
	}
	for _, opt := range opts {
		opt(cfg)
	}

	req := testcontainers.ContainerRequest{
		Name: cfg.Name,
		FromDockerfile: testcontainers.FromDockerfile{
			Context:        cfg.DockerfileDir,
			Dockerfile:     cfg.Dockerfile,
			BuildArgs:      cfg.BuildArgs,
			BuildLogWriter: cfg.LogOutput,
		},
		Networks:           cfg.Networks,
		Env:                cfg.Env,
		WaitingFor:         cfg.StartupWait,
		ExposedPorts:       []string{cfg.Port + "/tcp"},
		HostConfigModifier: DefaultHostConfig(),
	}

	genericContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start the app container %q: %w", cfg.Name, err)
	}

	mappedPort, err := genericContainer.MappedPort(ctx, cfg.Port+"/tcp")
	if err != nil {
		return nil, fmt.Errorf("failed to get the mapped port: %w", err)
	}

	host, err := genericContainer.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get the container host: %w", err)
	}

	go streamContainerLogs(ctx, genericContainer, cfg.LogOutput, cfg.Logger)

	cfg.Logger.Info(ctx, "App container started", "name", cfg.Name, "uri", net.JoinHostPort(host, mappedPort.Port()))

	return &Container{
		container:    genericContainer,
		externalHost: host,
		externalPort: mappedPort.Port(),
		cfg:          cfg,
	}, nil
}

// Address returns host:port of the app, reachable from the test process.
func (a *Container) Address() string {
	return net.JoinHostPort(a.externalHost, a.externalPort)
}

func (a *Container) Terminate(ctx context.Context) error {
	return a.container.Terminate(ctx)
}

func streamContainerLogs(ctx context.Context, container testcontainers.Container, out io.Writer, log Logger) {
	logs, err := container.Logs(ctx)
	if err != nil {
		log.Error(ctx, "failed to get container logs", "error", err)
		return
	}

	go func() {
		defer func() {
			if cerr := logs.Close(); cerr != nil {
				log.Error(ctx, "failed to close container logs", "error", cerr)
			}
		}()

		_, cerr := io.Copy(out, logs)
		if cerr != nil && cerr != io.EOF {
			log.Error(ctx, "error copying container logs", "error", cerr)
		}
	}()
}

func DefaultHostConfig() func(hc *container.HostConfig) {
	return func(hc *container.HostConfig) {
		hc.AutoRemove = true
	}
}
