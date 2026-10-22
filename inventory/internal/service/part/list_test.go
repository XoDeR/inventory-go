package part

import (
	"context"
	"errors"

	"inventory-go/inventory/internal/model"
)

func (s *ServiceSuite) TestListSuccess() {
	var (
		ctx = context.Background()

		filter = model.PartsFilter{
			Categories: []model.Category{model.CategoryGPU},
			Tags:       []string{"nvidia"},
		}

		expectedParts = []model.Part{
			{Uuid: "uuid-1", Name: "RTX 4090", Category: model.CategoryGPU},
			{Uuid: "uuid-2", Name: "RTX 4080", Category: model.CategoryGPU},
		}
	)

	// The filter is passed to the repository unchanged: the service must not modify it.
	s.inventoryRepository.EXPECT().List(ctx, filter).Return(expectedParts, nil)

	parts, err := s.service.List(ctx, filter)

	s.Require().NoError(err)
	s.Require().Equal(expectedParts, parts)
}

func (s *ServiceSuite) TestListEmptyFilter() {
	var (
		ctx    = context.Background()
		filter = model.PartsFilter{}

		expectedParts = []model.Part{{Uuid: "uuid-1"}, {Uuid: "uuid-2"}, {Uuid: "uuid-3"}}
	)

	s.inventoryRepository.EXPECT().List(ctx, filter).Return(expectedParts, nil)

	parts, err := s.service.List(ctx, filter)

	s.Require().NoError(err)
	s.Require().Len(parts, len(expectedParts))
}

func (s *ServiceSuite) TestListNothingFound() {
	var (
		ctx    = context.Background()
		filter = model.PartsFilter{Names: []string{"does not exist"}}
	)

	s.inventoryRepository.EXPECT().List(ctx, filter).Return([]model.Part{}, nil)

	parts, err := s.service.List(ctx, filter)

	s.Require().NoError(err)
	s.Require().Empty(parts)
}

func (s *ServiceSuite) TestListRepositoryError() {
	var (
		ctx     = context.Background()
		filter  = model.PartsFilter{}
		repoErr = errors.New("database is unavailable")
	)

	s.inventoryRepository.EXPECT().List(ctx, filter).Return(nil, repoErr)

	parts, err := s.service.List(ctx, filter)

	s.Require().ErrorIs(err, repoErr)
	s.Require().Nil(parts)
}
