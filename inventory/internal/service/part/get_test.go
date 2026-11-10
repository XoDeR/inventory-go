package part

import (
	"context"
	"errors"

	"inventory-go/inventory/internal/model"
)

const testUUID = "3b1f6c1e-6a5e-4b0e-9a54-2f6c1d9f7a10"

func testPart() model.Part {
	return model.Part{
		Uuid:          testUUID,
		Name:          "RTX 4090",
		Description:   "Flagship GPU",
		Price:         1999.99,
		StockQuantity: 5,
		Category:      model.CategoryGPU,
		Tags:          []string{"gpu", "nvidia"},
	}
}

func (s *ServiceSuite) TestGetCacheHit() {
	ctx := context.Background()
	expectedPart := testPart()

	// the repository mock has no expectations: calling it would fail the test
	s.inventoryCacheRepository.EXPECT().Get(ctx, testUUID).Return(expectedPart, nil)

	part, err := s.service.Get(ctx, testUUID)

	s.Require().NoError(err)
	s.Require().Equal(expectedPart, part)
}

func (s *ServiceSuite) TestGetCacheMissFillsCache() {
	ctx := context.Background()
	expectedPart := testPart()

	s.inventoryCacheRepository.EXPECT().Get(ctx, testUUID).Return(model.Part{}, model.ErrPartNotFound)
	s.inventoryRepository.EXPECT().Get(ctx, testUUID).Return(expectedPart, nil)
	s.inventoryCacheRepository.EXPECT().Set(ctx, expectedPart, testCacheTTL).Return(nil)

	part, err := s.service.Get(ctx, testUUID)

	s.Require().NoError(err)
	s.Require().Equal(expectedPart, part)
}

func (s *ServiceSuite) TestGetCacheUnavailableFallsBackToDatabase() {
	ctx := context.Background()
	expectedPart := testPart()

	s.inventoryCacheRepository.EXPECT().Get(ctx, testUUID).Return(model.Part{}, errors.New("redis is down"))
	s.inventoryRepository.EXPECT().Get(ctx, testUUID).Return(expectedPart, nil)
	s.inventoryCacheRepository.EXPECT().Set(ctx, expectedPart, testCacheTTL).Return(errors.New("redis is down"))

	part, err := s.service.Get(ctx, testUUID)

	s.Require().NoError(err)
	s.Require().Equal(expectedPart, part)
}

func (s *ServiceSuite) TestGetNotFound() {
	ctx := context.Background()

	s.inventoryCacheRepository.EXPECT().Get(ctx, testUUID).Return(model.Part{}, model.ErrPartNotFound)
	s.inventoryRepository.EXPECT().Get(ctx, testUUID).Return(model.Part{}, model.ErrPartNotFound)

	part, err := s.service.Get(ctx, testUUID)

	s.Require().ErrorIs(err, model.ErrPartNotFound)
	s.Require().Equal(model.Part{}, part)
}

func (s *ServiceSuite) TestGetRepositoryError() {
	var (
		ctx     = context.Background()
		repoErr = errors.New("database is unavailable")

		// The repository returns a non-empty part together with the error,
		// the service must not leak it to the caller (nor cache it).
		partWithError = model.Part{Uuid: testUUID, Name: "must not be returned"}
	)

	s.inventoryCacheRepository.EXPECT().Get(ctx, testUUID).Return(model.Part{}, model.ErrPartNotFound)
	s.inventoryRepository.EXPECT().Get(ctx, testUUID).Return(partWithError, repoErr)

	part, err := s.service.Get(ctx, testUUID)

	s.Require().ErrorIs(err, repoErr)
	s.Require().Equal(model.Part{}, part)
}
