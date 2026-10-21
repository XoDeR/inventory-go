package main

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	inventoryV1 "inventory-go/shared/pkg/proto/inventory/v1"
)

// seedParts returns the initial in-memory catalog. The service has no create
// RPC, so without it the inventory would always be empty.
func seedParts() []*inventoryV1.Part {
	now := timestamppb.Now()

	return []*inventoryV1.Part{
		{
			Uuid:          "6f1c8a4e-3b52-4d7a-9c10-1a2b3c4d5e01",
			Name:          "Ryzen 7 7800X3D",
			Description:   "8-core desktop processor with 3D V-Cache",
			Price:         399.99,
			StockQuantity: 25,
			Category:      inventoryV1.Category_CATEGORY_CPU,
			Dimensions:    &inventoryV1.Dimensions{Length: 4.0, Width: 4.0, Height: 0.5, Weight: 0.04},
			Manufacturer:  &inventoryV1.Manufacturer{Name: "AMD", Country: "USA", Website: "https://www.amd.com"},
			Tags:          []string{"cpu", "amd", "gaming"},
			CreatedAt:     now,
			UpdatedAt:     now,
		},
		{
			Uuid:          "6f1c8a4e-3b52-4d7a-9c10-1a2b3c4d5e02",
			Name:          "GeForce RTX 4070",
			Description:   "12 GB GDDR6X graphics card",
			Price:         549.0,
			StockQuantity: 12,
			Category:      inventoryV1.Category_CATEGORY_GPU,
			Dimensions:    &inventoryV1.Dimensions{Length: 24.0, Width: 11.0, Height: 4.0, Weight: 1.1},
			Manufacturer:  &inventoryV1.Manufacturer{Name: "NVIDIA", Country: "USA", Website: "https://www.nvidia.com"},
			Tags:          []string{"gpu", "nvidia", "gaming"},
			CreatedAt:     now,
			UpdatedAt:     now,
		},
		{
			Uuid:          "6f1c8a4e-3b52-4d7a-9c10-1a2b3c4d5e03",
			Name:          "Vengeance DDR5 32GB",
			Description:   "2x16 GB DDR5-6000 memory kit",
			Price:         119.5,
			StockQuantity: 60,
			Category:      inventoryV1.Category_CATEGORY_RAM,
			Dimensions:    &inventoryV1.Dimensions{Length: 13.5, Width: 0.8, Height: 4.5, Weight: 0.08},
			Manufacturer:  &inventoryV1.Manufacturer{Name: "Corsair", Country: "USA", Website: "https://www.corsair.com"},
			Tags:          []string{"ram", "ddr5"},
			CreatedAt:     now,
			UpdatedAt:     now,
		},
		{
			Uuid:          "6f1c8a4e-3b52-4d7a-9c10-1a2b3c4d5e04",
			Name:          "990 PRO 2TB",
			Description:   "NVMe PCIe 4.0 solid-state drive",
			Price:         169.0,
			StockQuantity: 40,
			Category:      inventoryV1.Category_CATEGORY_STORAGE,
			Dimensions:    &inventoryV1.Dimensions{Length: 8.0, Width: 2.2, Height: 0.2, Weight: 0.01},
			Manufacturer:  &inventoryV1.Manufacturer{Name: "Samsung", Country: "South Korea", Website: "https://www.samsung.com"},
			Tags:          []string{"ssd", "nvme"},
			CreatedAt:     now,
			UpdatedAt:     now,
		},
	}
}
