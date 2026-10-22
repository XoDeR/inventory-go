package part

import (
	"context"
	"errors"

	"inventory-go/inventory/internal/model"
)

func (s *ServiceSuite) TestGetSuccess() {
	var (
		ctx  = context.Background()
		uuid = "3b1f6c1e-6a5e-4b0e-9a54-2f6c1d9f7a10"

		expectedPart = model.Part{
			Uuid:          uuid,
			Name:          "RTX 4090",
			Description:   "Flagship GPU",
			Price:         1999.99,
			StockQuantity: 5,
			Category:      model.CategoryGPU,
			Tags:          []string{"gpu", "nvidia"},
		}
	)

	s.inventoryRepository.EXPECT().Get(ctx, uuid).Return(expectedPart, nil)

	part, err := s.service.Get(ctx, uuid)

	s.Require().NoError(err)
	s.Require().Equal(expectedPart, part)
}

func (s *ServiceSuite) TestGetNotFound() {
	var (
		ctx  = context.Background()
		uuid = "3b1f6c1e-6a5e-4b0e-9a54-2f6c1d9f7a10"
	)

	s.inventoryRepository.EXPECT().Get(ctx, uuid).Return(model.Part{}, model.ErrPartNotFound)

	part, err := s.service.Get(ctx, uuid)

	s.Require().ErrorIs(err, model.ErrPartNotFound)
	s.Require().Equal(model.Part{}, part)
}

func (s *ServiceSuite) TestGetRepositoryError() {
	var (
		ctx     = context.Background()
		uuid    = "3b1f6c1e-6a5e-4b0e-9a54-2f6c1d9f7a10"
		repoErr = errors.New("database is unavailable")

		// The repository returns a non-empty part together with the error,
		// the service must not leak it to the caller.
		partWithError = model.Part{Uuid: uuid, Name: "must not be returned"}
	)

	s.inventoryRepository.EXPECT().Get(ctx, uuid).Return(partWithError, repoErr)

	part, err := s.service.Get(ctx, uuid)

	s.Require().ErrorIs(err, repoErr)
	s.Require().Equal(model.Part{}, part)
}
