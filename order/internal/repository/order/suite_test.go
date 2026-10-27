package order

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/suite"

	"inventory-go/order/internal/migrator"
	"inventory-go/order/migrations"
)

// postgresURIEnv holds the connection string of a PostgreSQL used by these integration tests,
// e.g. postgres://order-service-user:order-service-password@localhost:5432/order-service.
// The user must be allowed to create databases. Start one with `task up-core up-order`.
// The tests are skipped when it is not set.
const postgresURIEnv = "POSTGRES_TEST_URI"

type RepositorySuite struct {
	suite.Suite

	adminPool  *pgxpool.Pool
	pool       *pgxpool.Pool
	dbName     string
	repository *repository
}

func (s *RepositorySuite) SetupSuite() {
	uri := os.Getenv(postgresURIEnv)
	if uri == "" {
		s.T().Skipf("%s is not set, skipping PostgreSQL repository tests", postgresURIEnv)
	}

	pool, err := pgxpool.New(context.Background(), uri)
	s.Require().NoError(err)

	s.adminPool = pool
}

func (s *RepositorySuite) TearDownSuite() {
	if s.adminPool != nil {
		s.adminPool.Close()
	}
}

// SetupTest runs before every test: each test gets its own empty database with all migrations
// applied, so no state leaks between tests and the migrations are verified along the way.
func (s *RepositorySuite) SetupTest() {
	ctx := context.Background()

	s.dbName = fmt.Sprintf("order_test_%d", time.Now().UnixNano())
	_, err := s.adminPool.Exec(ctx, fmt.Sprintf("CREATE DATABASE %q", s.dbName))
	s.Require().NoError(err)

	cfg := s.adminPool.Config()
	cfg.ConnConfig.Database = s.dbName

	s.pool, err = pgxpool.NewWithConfig(ctx, cfg)
	s.Require().NoError(err)

	db := stdlib.OpenDBFromPool(s.pool)
	s.Require().NoError(migrator.NewMigrator(db).Up(ctx, migrations.FS))
	s.Require().NoError(db.Close())

	s.repository = NewRepository(s.pool)
}

func (s *RepositorySuite) TearDownTest() {
	ctx := context.Background()

	s.pool.Close()
	_, err := s.adminPool.Exec(ctx, fmt.Sprintf("DROP DATABASE %q WITH (FORCE)", s.dbName))
	s.Require().NoError(err)
}

func TestRepositorySuite(t *testing.T) {
	suite.Run(t, new(RepositorySuite))
}

func stdlibDB(s *RepositorySuite) *sql.DB {
	return stdlib.OpenDBFromPool(s.pool)
}
