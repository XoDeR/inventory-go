package part

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	def "inventory-go/inventory/internal/repository"
	repoModel "inventory-go/inventory/internal/repository/model"
)

const collectionName = "parts"

var _ def.InventoryRepository = (*repository)(nil)

type repository struct {
	collection *mongo.Collection
}

func NewRepository(db *mongo.Database) *repository {
	return &repository{
		collection: db.Collection(collectionName),
	}
}

// Seed fills an empty collection with the initial catalog. The service has no create RPC,
// so without it the inventory would always be empty. A non-empty collection is left untouched.
func (r *repository) Seed(ctx context.Context) error {
	count, err := r.collection.CountDocuments(ctx, struct{}{})
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	_, err = r.collection.InsertMany(ctx, seedParts())

	return err
}

func seedParts() []repoModel.Part {
	now := time.Now().UTC()

	return []repoModel.Part{
		{
			Uuid:          "6f1c8a4e-3b52-4d7a-9c10-1a2b3c4d5e01",
			Name:          "Ryzen 7 7800X3D",
			Description:   "8-core desktop processor with 3D V-Cache",
			Price:         399.99,
			StockQuantity: 25,
			Category:      repoModel.CategoryCPU,
			Dimensions:    &repoModel.Dimensions{Length: 4.0, Width: 4.0, Height: 0.5, Weight: 0.04},
			Manufacturer:  &repoModel.Manufacturer{Name: "AMD", Country: "USA", Website: "https://www.amd.com"},
			Tags:          []string{"cpu", "amd", "gaming"},
			CreatedAt:     &now,
			UpdatedAt:     &now,
		},
		{
			Uuid:          "6f1c8a4e-3b52-4d7a-9c10-1a2b3c4d5e02",
			Name:          "GeForce RTX 4070",
			Description:   "12 GB GDDR6X graphics card",
			Price:         549.0,
			StockQuantity: 12,
			Category:      repoModel.CategoryGPU,
			Dimensions:    &repoModel.Dimensions{Length: 24.0, Width: 11.0, Height: 4.0, Weight: 1.1},
			Manufacturer:  &repoModel.Manufacturer{Name: "NVIDIA", Country: "USA", Website: "https://www.nvidia.com"},
			Tags:          []string{"gpu", "nvidia", "gaming"},
			CreatedAt:     &now,
			UpdatedAt:     &now,
		},
		{
			Uuid:          "6f1c8a4e-3b52-4d7a-9c10-1a2b3c4d5e03",
			Name:          "Vengeance DDR5 32GB",
			Description:   "2x16 GB DDR5-6000 memory kit",
			Price:         119.5,
			StockQuantity: 60,
			Category:      repoModel.CategoryRAM,
			Dimensions:    &repoModel.Dimensions{Length: 13.5, Width: 0.8, Height: 4.5, Weight: 0.08},
			Manufacturer:  &repoModel.Manufacturer{Name: "Corsair", Country: "USA", Website: "https://www.corsair.com"},
			Tags:          []string{"ram", "ddr5"},
			CreatedAt:     &now,
			UpdatedAt:     &now,
		},
		{
			Uuid:          "6f1c8a4e-3b52-4d7a-9c10-1a2b3c4d5e04",
			Name:          "990 PRO 2TB",
			Description:   "NVMe PCIe 4.0 solid-state drive",
			Price:         169.0,
			StockQuantity: 40,
			Category:      repoModel.CategoryStorage,
			Dimensions:    &repoModel.Dimensions{Length: 8.0, Width: 2.2, Height: 0.2, Weight: 0.01},
			Manufacturer:  &repoModel.Manufacturer{Name: "Samsung", Country: "South Korea", Website: "https://www.samsung.com"},
			Tags:          []string{"ssd", "nvme"},
			CreatedAt:     &now,
			UpdatedAt:     &now,
		},
	}
}
