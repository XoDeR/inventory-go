package redis

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"inventory-go/platform/pkg/logger"
)

const (
	redisPort           = "6379"
	redisStartupTimeout = 1 * time.Minute
)

type Logger interface {
	Info(ctx context.Context, msg string, args ...any)
	Error(ctx context.Context, msg string, args ...any)
}

type Config struct {
	NetworkName   string
	ContainerName string
	ImageName     string
	Logger        Logger

	// Host and Port are the address of the container as seen from the test process.
	Host string
	Port string
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

func WithLogger(l Logger) Option {
	return func(c *Config) { c.Logger = l }
}

func buildConfig(opts ...Option) *Config {
	cfg := &Config{
		NetworkName:   "test-network",
		ContainerName: "redis-container",
		ImageName:     "redis:7.2.5-alpine3.20",
		Logger:        logger.NoopLogger{},
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return cfg
}

// Container is a Redis container.
type Container struct {
	container testcontainers.Container
	cfg       *Config
}

func NewContainer(ctx context.Context, opts ...Option) (*Container, error) {
	cfg := buildConfig(opts...)

	req := testcontainers.ContainerRequest{
		Name:         cfg.ContainerName,
		Image:        cfg.ImageName,
		Networks:     []string{cfg.NetworkName},
		ExposedPorts: []string{redisPort + "/tcp"},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort(redisPort+"/tcp"),
			wait.ForLog("Ready to accept connections"),
		).WithDeadline(redisStartupTimeout),
		HostConfigModifier: func(hc *container.HostConfig) { hc.AutoRemove = true },
	}

	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start redis container: %w", err)
	}

	host, err := c.Host(ctx)
	if err != nil {
		_ = c.Terminate(ctx)
		return nil, fmt.Errorf("failed to get container host: %w", err)
	}

	port, err := c.MappedPort(ctx, redisPort+"/tcp")
	if err != nil {
		_ = c.Terminate(ctx)
		return nil, fmt.Errorf("failed to get mapped port: %w", err)
	}

	cfg.Host, cfg.Port = host, port.Port()
	cfg.Logger.Info(ctx, "Redis container started", "host", cfg.Host, "port", cfg.Port)

	return &Container{container: c, cfg: cfg}, nil
}

func (c *Container) Config() *Config {
	return c.cfg
}

// Address returns host:port of the container for connections from the test process.
func (c *Container) Address() string {
	return net.JoinHostPort(c.cfg.Host, c.cfg.Port)
}

func (c *Container) Terminate(ctx context.Context) error {
	if err := c.container.Terminate(ctx); err != nil {
		c.cfg.Logger.Error(ctx, "failed to terminate redis container", "error", err)
	}

	c.cfg.Logger.Info(ctx, "Redis container terminated")

	return nil
}
