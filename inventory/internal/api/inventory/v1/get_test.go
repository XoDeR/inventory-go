package v1

import (
	"context"
	"errors"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"inventory-go/inventory/internal/converter"
	"inventory-go/inventory/internal/model"
	inventoryV1 "inventory-go/shared/pkg/proto/inventory/v1"
)

func (s *APISuite) TestGetPartSuccess() {
	var (
		ctx  = context.Background()
		uuid = "3b1f6c1e-6a5e-4b0e-9a54-2f6c1d9f7a10"
		now  = time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

		part = model.Part{
			Uuid:          uuid,
			Name:          "RTX 4090",
			Description:   "Flagship GPU",
			Price:         1999.99,
			StockQuantity: 5,
			Category:      model.CategoryGPU,
			Dimensions:    &model.Dimensions{Length: 30, Width: 14, Height: 6, Weight: 2.2},
			Manufacturer:  &model.Manufacturer{Name: "NVIDIA", Country: "USA", Website: "https://www.nvidia.com"},
			Tags:          []string{"gpu", "nvidia"},
			Metadata:      map[string]any{"series": "40", "ports": int64(4), "ratio": 1.5, "rgb": true},
			CreatedAt:     &now,
			UpdatedAt:     &now,
		}
	)

	s.inventoryService.EXPECT().Get(ctx, uuid).Return(part, nil)

	resp, err := s.api.GetPart(ctx, &inventoryV1.GetPartRequest{Uuid: uuid})

	s.Require().NoError(err)
	s.Require().Equal(converter.PartToProto(part), resp.GetPart())
	s.Require().Equal(uuid, resp.GetPart().GetUuid())
	s.Require().Equal(inventoryV1.Category_CATEGORY_GPU, resp.GetPart().GetCategory())
	s.Require().Len(resp.GetPart().GetMetadata(), 4)
}

func (s *APISuite) TestGetPartNotFound() {
	var (
		ctx  = context.Background()
		uuid = "3b1f6c1e-6a5e-4b0e-9a54-2f6c1d9f7a10"
	)

	s.inventoryService.EXPECT().Get(ctx, uuid).Return(model.Part{}, model.ErrPartNotFound)

	resp, err := s.api.GetPart(ctx, &inventoryV1.GetPartRequest{Uuid: uuid})

	s.Require().Nil(resp)
	s.Require().Equal(codes.NotFound, status.Code(err))
	s.Require().Contains(status.Convert(err).Message(), uuid)
}

func (s *APISuite) TestGetPartServiceError() {
	var (
		ctx        = context.Background()
		uuid       = "3b1f6c1e-6a5e-4b0e-9a54-2f6c1d9f7a10"
		serviceErr = errors.New("database is unavailable")
	)

	s.inventoryService.EXPECT().Get(ctx, uuid).Return(model.Part{}, serviceErr)

	resp, err := s.api.GetPart(ctx, &inventoryV1.GetPartRequest{Uuid: uuid})

	s.Require().Nil(resp)
	s.Require().ErrorIs(err, serviceErr)
}
