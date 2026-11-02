package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"inventory-go/platform/pkg/logger"
)

const (
	postgresPort           = "5432"
	postgresStartupTimeout = 1 * time.Minute
)

type Logger interface {
	Info(ctx context.Context, msg string, args ...any)
	Error(ctx context.Context, msg string, args ...any)
}

type Config struct {
	NetworkName   string
	ContainerName string
	ImageName     string
	Database      string
	Username      string
	Password      string
	Logger        Logger
}

type Option func(*Config)

func WithNetworkName(network string) Option {
	return func(c *Config) { c.NetworkName = network }
}

func WithContainerName(name string) Option {
	return func(c *Config) { c.ContainerName = name }
}

func WithImageName(image string) Option {
	return func(c *Config) { c.ImageName = image }
}

func WithDatabase(database string) Option {
	return func(c *Config) { c.Database = database }
}

func WithAuth(username, password string) Option {
	return func(c *Config) {
		c.Username = username
		c.Password = password
	}
}

func WithLogger(logger Logger) Option {
	return func(c *Config) { c.Logger = logger }
}

// Container is a PostgreSQL container.
type Container struct {
	container testcontainers.Container
	cfg       *Config
}

func NewContainer(ctx context.Context, opts ...Option) (*Container, error) {
	cfg := &Config{
		NetworkName:   "test-network",
		ContainerName: "postgres-container",
		ImageName:     "postgres:17-alpine",
		Database:      "test",
		Username:      "postgres",
		Password:      "postgres",
		Logger:        logger.NoopLogger{},
	}
	for _, opt := range opts {
		opt(cfg)
	}

	req := testcontainers.ContainerRequest{
		Name:     cfg.ContainerName,
		Image:    cfg.ImageName,
		Networks: []string{cfg.NetworkName},
		Env: map[string]string{
			"POSTGRES_DB":       cfg.Database,
			"POSTGRES_USER":     cfg.Username,
			"POSTGRES_PASSWORD": cfg.Password,
		},
		ExposedPorts: []string{postgresPort + "/tcp"},
		// PostgreSQL starts a temporary server during the init, so the "ready" line shows up twice
		WaitingFor: wait.ForAll(
			wait.ForLog("database system is ready to accept connections").AsRegexp().WithOccurrence(2),
			wait.ForListeningPort(postgresPort+"/tcp"),
		).WithStartupTimeout(postgresStartupTimeout),
		HostConfigModifier: func(hc *container.HostConfig) { hc.AutoRemove = true },
	}

	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start postgres container: %w", err)
	}

	cfg.Logger.Info(ctx, "Postgres container started", "name", cfg.ContainerName)

	return &Container{container: c, cfg: cfg}, nil
}

func (c *Container) Config() *Config {
	return c.cfg
}

func (c *Container) Terminate(ctx context.Context) error {
	err := c.container.Terminate(ctx)
	if err != nil {
		c.cfg.Logger.Error(ctx, "failed to terminate postgres container", "error", err)
		return err
	}

	c.cfg.Logger.Info(ctx, "Postgres container terminated")

	return nil
}
