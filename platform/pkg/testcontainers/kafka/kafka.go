// Package kafka starts a single-node Kafka (KRaft, no ZooKeeper) for e2e tests.
// Other containers in the same network reach it at <ContainerName>:<Port>.
package kafka

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
	// Port is the port of the broker inside the Docker network.
	Port = "9092"

	controllerPort = "9093"
	startupTimeout = 2 * time.Minute
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
}

type Option func(*Config)

func WithNetworkName(network string) Option { return func(c *Config) { c.NetworkName = network } }

func WithContainerName(name string) Option { return func(c *Config) { c.ContainerName = name } }

func WithImageName(image string) Option { return func(c *Config) { c.ImageName = image } }

func WithLogger(l Logger) Option { return func(c *Config) { c.Logger = l } }

// Container is a running Kafka broker.
type Container struct {
	container testcontainers.Container
	cfg       *Config
}

func NewContainer(ctx context.Context, opts ...Option) (*Container, error) {
	cfg := &Config{
		NetworkName:   "test-network",
		ContainerName: "kafka-container",
		ImageName:     "confluentinc/cp-kafka:7.9.0",
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
			"KAFKA_NODE_ID":                                  "1",
			"KAFKA_PROCESS_ROLES":                            "broker,controller",
			"KAFKA_CONTROLLER_QUORUM_VOTERS":                 "1@" + cfg.ContainerName + ":" + controllerPort,
			"KAFKA_LISTENERS":                                "PLAINTEXT://0.0.0.0:" + Port + ",CONTROLLER://" + cfg.ContainerName + ":" + controllerPort,
			"KAFKA_ADVERTISED_LISTENERS":                     "PLAINTEXT://" + cfg.ContainerName + ":" + Port + ",CONTROLLER://" + cfg.ContainerName + ":" + controllerPort,
			"KAFKA_LISTENER_SECURITY_PROTOCOL_MAP":           "PLAINTEXT:PLAINTEXT,CONTROLLER:PLAINTEXT",
			"KAFKA_INTER_BROKER_LISTENER_NAME":               "PLAINTEXT",
			"KAFKA_CONTROLLER_LISTENER_NAMES":                "CONTROLLER",
			"KAFKA_AUTO_CREATE_TOPICS_ENABLE":                "true",
			"KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR":         "1",
			"KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR": "1",
			"KAFKA_TRANSACTION_STATE_LOG_MIN_ISR":            "1",
			"KAFKA_GROUP_INITIAL_REBALANCE_DELAY_MS":         "0",
			"CLUSTER_ID":                                     "Mk3OEYBSD34fcwNTJENDM2Qk",
		},
		ExposedPorts: []string{Port + "/tcp"},
		WaitingFor:   wait.ForListeningPort(Port + "/tcp").WithStartupTimeout(startupTimeout),
		HostConfigModifier: func(hc *container.HostConfig) {
			hc.AutoRemove = true
		},
	}

	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start kafka container: %w", err)
	}

	cfg.Logger.Info(ctx, "Kafka container started", "name", cfg.ContainerName)

	return &Container{container: c, cfg: cfg}, nil
}

// Address is the broker address inside the Docker network.
func (c *Container) Address() string {
	return c.cfg.ContainerName + ":" + Port
}

func (c *Container) Terminate(ctx context.Context) error {
	if err := c.container.Terminate(ctx); err != nil {
		c.cfg.Logger.Error(ctx, "failed to terminate kafka container", "error", err)
		return err
	}

	c.cfg.Logger.Info(ctx, "Kafka container terminated")

	return nil
}
