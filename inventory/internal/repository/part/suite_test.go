package part

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	repoModel "inventory-go/inventory/internal/repository/model"
)

const (
	gpuUuid     = "uuid-gpu"
	cpuUuid     = "uuid-cpu"
	ramUuid     = "uuid-ram"
	noManufUuid = "uuid-no-manufacturer"
)

type RepositorySuite struct {
	suite.Suite

	repository *repository
}

// SetupTest runs before every test, so each test gets a fresh repository with a known
// set of parts instead of the production seed, and no state leaks between tests.
func (s *RepositorySuite) SetupTest() {
	s.repository = NewRepository()
	s.repository.data = make(map[string]repoModel.Part)

	for _, part := range testParts() {
		s.repository.data[part.Uuid] = part
	}
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
