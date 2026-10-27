package part

import (
	"context"
	"time"

	"inventory-go/inventory/internal/model"
)

func (s *RepositorySuite) TestGetSuccess() {
	var (
		ctx = context.Background()
		now = time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

		expectedPart = model.Part{
			Uuid:          gpuUuid,
			Name:          "RTX 4090",
			Description:   "Flagship GPU",
			Price:         1999.99,
			StockQuantity: 5,
			Category:      model.CategoryGPU,
			Dimensions:    &model.Dimensions{Length: 30, Width: 14, Height: 6, Weight: 2.2},
			Manufacturer:  &model.Manufacturer{Name: "NVIDIA", Country: "USA", Website: "https://www.nvidia.com"},
			Tags:          []string{"gpu", "nvidia"},
			Metadata:      map[string]any{"series": "40"},
			CreatedAt:     &now,
			UpdatedAt:     &now,
		}
	)

	part, err := s.repository.Get(ctx, gpuUuid)

	s.Require().NoError(err)
	s.Require().Equal(expectedPart, part)
}

func (s *RepositorySuite) TestGetWithoutOptionalFields() {
	ctx := context.Background()

	part, err := s.repository.Get(ctx, noManufUuid)

	s.Require().NoError(err)
	s.Require().Equal(noManufUuid, part.Uuid)
	s.Require().Nil(part.Dimensions)
	s.Require().Nil(part.Manufacturer)
}

func (s *RepositorySuite) TestGetNotFound() {
	ctx := context.Background()

	part, err := s.repository.Get(ctx, "does-not-exist")

	s.Require().ErrorIs(err, model.ErrPartNotFound)
	s.Require().Equal(model.Part{}, part)
}

func (s *RepositorySuite) TestSeedFillsEmptyCollectionOnce() {
	ctx := context.Background()
	s.Require().NoError(s.repository.collection.Drop(ctx))

	s.Require().NoError(s.repository.Seed(ctx))
	s.Require().NoError(s.repository.Seed(ctx))

	parts, err := s.repository.List(ctx, model.PartsFilter{})

	s.Require().NoError(err)
	s.Require().Len(parts, len(seedParts()))
}

func (s *RepositorySuite) TestSeedKeepsExistingData() {
	ctx := context.Background()

	s.Require().NoError(s.repository.Seed(ctx))

	parts, err := s.repository.List(ctx, model.PartsFilter{})

	s.Require().NoError(err)
	s.Require().Len(parts, len(testParts()))
}
