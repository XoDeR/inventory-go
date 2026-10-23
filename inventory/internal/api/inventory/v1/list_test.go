package v1

import (
	"context"
	"errors"

	"inventory-go/inventory/internal/model"
	inventoryV1 "inventory-go/shared/pkg/proto/inventory/v1"
)

func (s *APISuite) TestListPartsSuccess() {
	var (
		ctx = context.Background()

		req = &inventoryV1.ListPartsRequest{
			Filter: &inventoryV1.PartsFilter{
				Uuids:                 []string{"uuid-1"},
				Names:                 []string{"RTX 4090"},
				Categories:            []inventoryV1.Category{inventoryV1.Category_CATEGORY_GPU},
				ManufacturerCountries: []string{"USA"},
				Tags:                  []string{"nvidia"},
			},
		}

		// The proto filter must be converted to the model filter before it reaches the service.
		expectedFilter = model.PartsFilter{
			Uuids:                 []string{"uuid-1"},
			Names:                 []string{"RTX 4090"},
			Categories:            []model.Category{model.CategoryGPU},
			ManufacturerCountries: []string{"USA"},
			Tags:                  []string{"nvidia"},
		}

		parts = []model.Part{
			{Uuid: "uuid-1", Name: "RTX 4090", Category: model.CategoryGPU},
			{Uuid: "uuid-2", Name: "RTX 4080", Category: model.CategoryGPU},
		}
	)

	s.inventoryService.EXPECT().List(ctx, expectedFilter).Return(parts, nil)

	resp, err := s.api.ListParts(ctx, req)

	s.Require().NoError(err)
	s.Require().Len(resp.GetParts(), len(parts))
	s.Require().Equal("uuid-1", resp.GetParts()[0].GetUuid())
	s.Require().Equal("uuid-2", resp.GetParts()[1].GetUuid())
	s.Require().Equal(inventoryV1.Category_CATEGORY_GPU, resp.GetParts()[0].GetCategory())
}

func (s *APISuite) TestListPartsNilFilter() {
	ctx := context.Background()

	// A request without a filter means "no filtering", not an error.
	s.inventoryService.EXPECT().List(ctx, model.PartsFilter{}).Return([]model.Part{{Uuid: "uuid-1"}}, nil)

	resp, err := s.api.ListParts(ctx, &inventoryV1.ListPartsRequest{})

	s.Require().NoError(err)
	s.Require().Len(resp.GetParts(), 1)
}

func (s *APISuite) TestListPartsNothingFound() {
	ctx := context.Background()

	s.inventoryService.EXPECT().List(ctx, model.PartsFilter{Categories: []model.Category{}}).Return([]model.Part{}, nil)

	resp, err := s.api.ListParts(ctx, &inventoryV1.ListPartsRequest{Filter: &inventoryV1.PartsFilter{}})

	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.Require().Empty(resp.GetParts())
}

func (s *APISuite) TestListPartsServiceError() {
	var (
		ctx        = context.Background()
		serviceErr = errors.New("database is unavailable")
	)

	s.inventoryService.EXPECT().List(ctx, model.PartsFilter{}).Return(nil, serviceErr)

	resp, err := s.api.ListParts(ctx, &inventoryV1.ListPartsRequest{})

	s.Require().Nil(resp)
	s.Require().ErrorIs(err, serviceErr)
}
