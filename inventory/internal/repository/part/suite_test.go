package part

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	repoModel "inventory-go/inventory/internal/repository/model"
)

const (
	gpuUuid     = "uuid-gpu"
	cpuUuid     = "uuid-cpu"
	ramUuid     = "uuid-ram"
	noManufUuid = "uuid-no-manufacturer"

	// mongoURIEnv holds the connection string of a MongoDB used by these integration tests,
	// e.g. mongodb://inventory-service-user:inventory-service-password@localhost:27017/?authSource=admin.
	// Start one with `task up-core up-inventory`. The tests are skipped when it is not set.
	mongoURIEnv = "MONGO_TEST_URI"
)

type RepositorySuite struct {
	suite.Suite

	client     *mongo.Client
	db         *mongo.Database
	repository *repository
}

func (s *RepositorySuite) SetupSuite() {
	uri := os.Getenv(mongoURIEnv)
	if uri == "" {
		s.T().Skipf("%s is not set, skipping MongoDB repository tests", mongoURIEnv)
	}

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	s.Require().NoError(err)

	s.client = client
}

func (s *RepositorySuite) TearDownSuite() {
	if s.client != nil {
		s.Require().NoError(s.client.Disconnect(context.Background()))
	}
}

// SetupTest runs before every test: each test gets its own empty database filled with a known
// set of parts instead of the production seed, so no state leaks between tests.
func (s *RepositorySuite) SetupTest() {
	ctx := context.Background()

	s.db = s.client.Database(fmt.Sprintf("inventory-test-%d", time.Now().UnixNano()))
	s.repository = NewRepository(s.db)

	_, err := s.repository.collection.InsertMany(ctx, testParts())
	s.Require().NoError(err)
}

func (s *RepositorySuite) TearDownTest() {
	s.Require().NoError(s.db.Drop(context.Background()))
}

func testParts() []repoModel.Part {
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

	return []repoModel.Part{
		{
			Uuid:          gpuUuid,
			Name:          "RTX 4090",
			Description:   "Flagship GPU",
			Price:         1999.99,
			StockQuantity: 5,
			Category:      repoModel.CategoryGPU,
			Dimensions:    &repoModel.Dimensions{Length: 30, Width: 14, Height: 6, Weight: 2.2},
			Manufacturer:  &repoModel.Manufacturer{Name: "NVIDIA", Country: "USA", Website: "https://www.nvidia.com"},
			Tags:          []string{"gpu", "nvidia"},
			Metadata:      map[string]any{"series": "40"},
			CreatedAt:     &now,
			UpdatedAt:     &now,
		},
		{
			Uuid:         cpuUuid,
			Name:         "Ryzen 7 7800X3D",
			Category:     repoModel.CategoryCPU,
			Manufacturer: &repoModel.Manufacturer{Name: "AMD", Country: "USA"},
			Tags:         []string{"cpu", "amd"},
		},
		{
			Uuid:         ramUuid,
			Name:         "Vengeance DDR5 32GB",
			Category:     repoModel.CategoryRAM,
			Manufacturer: &repoModel.Manufacturer{Name: "Samsung", Country: "South Korea"},
			Tags:         []string{"ram", "ddr5"},
		},
		{
			Uuid:     noManufUuid,
			Name:     "Generic SSD",
			Category: repoModel.CategoryStorage,
			Tags:     []string{"ssd"},
		},
	}
}

func TestRepositorySuite(t *testing.T) {
	suite.Run(t, new(RepositorySuite))
}
