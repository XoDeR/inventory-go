package mongo

import (
	"context"
	"fmt"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

const (
	mongoPort           = "27017"
	mongoStartupTimeout = 1 * time.Minute

	mongoEnvUsernameKey = "MONGO_INITDB_ROOT_USERNAME"
	mongoEnvPasswordKey = "MONGO_INITDB_ROOT_PASSWORD" //nolint:gosec
)

// Container is a MongoDB container together with a client connected to it from the test process.
type Container struct {
	container testcontainers.Container
	client    *mongo.Client
	cfg       *Config
}

func NewContainer(ctx context.Context, opts ...Option) (*Container, error) {
	cfg := buildConfig(opts...)

	container, err := startMongoContainer(ctx, cfg)
	if err != nil {
		return nil, err
	}

	success := false
	defer func() {
		if !success {
			if terr := container.Terminate(ctx); terr != nil {
				cfg.Logger.Error(ctx, "failed to terminate mongo container", "error", terr)
			}
		}
	}()

	cfg.Host, cfg.Port, err = getContainerHostPort(ctx, container)
	if err != nil {
		return nil, err
	}

	client, err := connectMongoClient(ctx, buildMongoURI(cfg))
	if err != nil {
		return nil, err
	}

	cfg.Logger.Info(ctx, "Mongo container started", "host", cfg.Host, "port", cfg.Port)
	success = true

	return &Container{
		container: container,
		client:    client,
		cfg:       cfg,
	}, nil
}

func (c *Container) Client() *mongo.Client {
	return c.client
}

func (c *Container) Config() *Config {
	return c.cfg
}

func (c *Container) Terminate(ctx context.Context) error {
	if err := c.client.Disconnect(ctx); err != nil {
		c.cfg.Logger.Error(ctx, "failed to disconnect mongo client", "error", err)
	}

	if err := c.container.Terminate(ctx); err != nil {
		c.cfg.Logger.Error(ctx, "failed to terminate mongo container", "error", err)
	}

	c.cfg.Logger.Info(ctx, "Mongo container terminated")

	return nil
}

func startMongoContainer(ctx context.Context, cfg *Config) (testcontainers.Container, error) {
	req := testcontainers.ContainerRequest{
		Name:     cfg.ContainerName,
		Image:    cfg.ImageName,
		Networks: []string{cfg.NetworkName},
		Env: map[string]string{
			mongoEnvUsernameKey: cfg.Username,
			mongoEnvPasswordKey: cfg.Password,
		},
		ExposedPorts:       []string{mongoPort + "/tcp"},
		WaitingFor:         wait.ForListeningPort(mongoPort + "/tcp").WithStartupTimeout(mongoStartupTimeout),
		HostConfigModifier: defaultHostConfig(),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start mongo container: %w", err)
	}

	return container, nil
}

func getContainerHostPort(ctx context.Context, container testcontainers.Container) (string, string, error) {
	host, err := container.Host(ctx)
	if err != nil {
		return "", "", fmt.Errorf("failed to get container host: %w", err)
	}

	port, err := container.MappedPort(ctx, mongoPort+"/tcp")
	if err != nil {
		return "", "", fmt.Errorf("failed to get mapped port: %w", err)
	}

	return host, port.Port(), nil
}

func buildMongoURI(cfg *Config) string {
	return fmt.Sprintf(
		"mongodb://%s:%s@%s:%s/%s?authSource=%s",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
		cfg.AuthDB,
	)
}

func connectMongoClient(ctx context.Context, uri string) (*mongo.Client, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mongo: %w", err)
	}

	if err = client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, fmt.Errorf("failed to ping mongo: %w", err)
	}

	return client, nil
}
