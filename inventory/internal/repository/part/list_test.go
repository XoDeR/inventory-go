package part

import (
	"context"

	"inventory-go/inventory/internal/model"
)

func uuidsOf(parts []model.Part) []string {
	uuids := make([]string, 0, len(parts))
	for _, part := range parts {
		uuids = append(uuids, part.Uuid)
	}

	return uuids
}

func (s *RepositorySuite) TestListEmptyFilter() {
	ctx := context.Background()

	parts, err := s.repository.List(ctx, model.PartsFilter{})

	s.Require().NoError(err)
	s.Require().ElementsMatch([]string{gpuUuid, cpuUuid, ramUuid, noManufUuid}, uuidsOf(parts))
}

func (s *RepositorySuite) TestListByUuids() {
	ctx := context.Background()
	filter := model.PartsFilter{Uuids: []string{gpuUuid, ramUuid, "unknown"}}

	parts, err := s.repository.List(ctx, filter)

	s.Require().NoError(err)
	s.Require().ElementsMatch([]string{gpuUuid, ramUuid}, uuidsOf(parts))
}

func (s *RepositorySuite) TestListByNames() {
	ctx := context.Background()
	filter := model.PartsFilter{Names: []string{"RTX 4090", "Generic SSD"}}

	parts, err := s.repository.List(ctx, filter)

	s.Require().NoError(err)
	s.Require().ElementsMatch([]string{gpuUuid, noManufUuid}, uuidsOf(parts))
}

func (s *RepositorySuite) TestListByCategories() {
	ctx := context.Background()
	filter := model.PartsFilter{Categories: []model.Category{model.CategoryCPU, model.CategoryRAM}}

	parts, err := s.repository.List(ctx, filter)

	s.Require().NoError(err)
	s.Require().ElementsMatch([]string{cpuUuid, ramUuid}, uuidsOf(parts))
}

func (s *RepositorySuite) TestListByManufacturerCountries() {
	ctx := context.Background()
	filter := model.PartsFilter{ManufacturerCountries: []string{"USA"}}

	parts, err := s.repository.List(ctx, filter)

	s.Require().NoError(err)
	// The part without a manufacturer has no country, so it never matches.
	s.Require().ElementsMatch([]string{gpuUuid, cpuUuid}, uuidsOf(parts))
}

func (s *RepositorySuite) TestListByTags() {
	ctx := context.Background()
	filter := model.PartsFilter{Tags: []string{"nvidia", "ddr5"}}

	parts, err := s.repository.List(ctx, filter)

	s.Require().NoError(err)
	// OR within a field: it is enough for a part to have any of the tags.
	s.Require().ElementsMatch([]string{gpuUuid, ramUuid}, uuidsOf(parts))
}

func (s *RepositorySuite) TestListCombinedFiltersUseAnd() {
	ctx := context.Background()
	filter := model.PartsFilter{
		ManufacturerCountries: []string{"USA"},
		Categories:            []model.Category{model.CategoryGPU, model.CategoryRAM},
	}

	parts, err := s.repository.List(ctx, filter)

	s.Require().NoError(err)
	// RAM is in the categories but made in South Korea, so only the GPU satisfies both fields.
	s.Require().ElementsMatch([]string{gpuUuid}, uuidsOf(parts))
}

func (s *RepositorySuite) TestListNothingFound() {
	ctx := context.Background()
	filter := model.PartsFilter{Names: []string{"does not exist"}}

	parts, err := s.repository.List(ctx, filter)

	s.Require().NoError(err)
	s.Require().NotNil(parts)
	s.Require().Empty(parts)
}
