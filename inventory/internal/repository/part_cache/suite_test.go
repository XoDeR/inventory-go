package part_cache

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	cacheRedis "inventory-go/platform/pkg/cache/redis"
	"inventory-go/platform/pkg/logger"
)

// redisAddrEnv holds the address of a Redis used by these integration tests, e.g. localhost:6379.
// Start one with `task up-core up-inventory`. The tests are skipped when it is not set.
const redisAddrEnv = "REDIS_TEST_ADDR"

type RepositorySuite struct {
	suite.Suite

	repository *repository
}

func (s *RepositorySuite) SetupSuite() {
	addr := os.Getenv(redisAddrEnv)
	if addr == "" {
		s.T().Skipf("%s is not set, skipping Redis repository tests", redisAddrEnv)
	}

	pool := cacheRedis.NewPool(addr, 2, time.Minute, 5*time.Second)
	client := cacheRedis.NewClient(pool, logger.NoopLogger{}, 5*time.Second)

	s.Require().NoError(client.Ping(context.Background()))

	s.repository = NewRepository(client)
}

func TestRepositorySuite(t *testing.T) {
	suite.Run(t, new(RepositorySuite))
}
