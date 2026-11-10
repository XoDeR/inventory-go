package part_cache

import (
	"context"
	"time"

	"github.com/google/uuid"

	"inventory-go/inventory/internal/model"
)

func (s *RepositorySuite) TestSetGetRoundTrip() {
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	part := model.Part{
		Uuid:          uuid.NewString(),
		Name:          "RTX 4090",
		Price:         1999.99,
		StockQuantity: 5,
		Category:      model.CategoryGPU,
		Dimensions:    &model.Dimensions{Length: 30, Width: 14, Height: 6, Weight: 2},
		Manufacturer:  &model.Manufacturer{Name: "NVIDIA", Country: "USA"},
		Tags:          []string{"gpu"},
		// the types of the metadata values must survive the round trip
		Metadata:  map[string]any{"vram_gb": int64(24), "boost": 2.5, "rgb": true, "series": "40"},
		CreatedAt: &now,
	}

	s.Require().NoError(s.repository.Set(ctx, part, time.Minute))
	s.T().Cleanup(func() { _ = s.repository.Delete(ctx, part.Uuid) })

	got, err := s.repository.Get(ctx, part.Uuid)

	s.Require().NoError(err)
	s.Require().Equal(part.Metadata, got.Metadata)
	s.Require().Equal(part.Dimensions, got.Dimensions)
	s.Require().Equal(part.Manufacturer, got.Manufacturer)
	s.Require().True(part.CreatedAt.Equal(*got.CreatedAt))
	s.Require().Equal(part.Name, got.Name)
}

func (s *RepositorySuite) TestGetMiss() {
	_, err := s.repository.Get(context.Background(), uuid.NewString())

	s.Require().ErrorIs(err, model.ErrPartNotFound)
}

func (s *RepositorySuite) TestDelete() {
	ctx := context.Background()
	part := model.Part{Uuid: uuid.NewString(), Name: "to delete"}

	s.Require().NoError(s.repository.Set(ctx, part, time.Minute))
	s.Require().NoError(s.repository.Delete(ctx, part.Uuid))

	_, err := s.repository.Get(ctx, part.Uuid)
	s.Require().ErrorIs(err, model.ErrPartNotFound)
}

func (s *RepositorySuite) TestExpiresAfterTTL() {
	ctx := context.Background()
	part := model.Part{Uuid: uuid.NewString(), Name: "short lived"}

	s.Require().NoError(s.repository.Set(ctx, part, 200*time.Millisecond))

	s.Require().Eventually(func() bool {
		_, err := s.repository.Get(ctx, part.Uuid)
		return err == model.ErrPartNotFound
	}, 3*time.Second, 100*time.Millisecond)
}
