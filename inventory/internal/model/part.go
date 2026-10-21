package model

import "time"

type Category int32

const (
	CategoryUnspecified Category = iota
	CategoryCPU
	CategoryGPU
	CategoryRAM
	CategoryStorage
)

type Part struct {
	Uuid          string
	Name          string
	Description   string
	Price         float64
	StockQuantity int64
	Category      Category
	Dimensions    *Dimensions
	Manufacturer  *Manufacturer
	Tags          []string
	// Metadata values are one of: string, int64, float64, bool.
	Metadata  map[string]any
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

type Dimensions struct {
	Length float64
	Width  float64
	Height float64
	Weight float64
}

type Manufacturer struct {
	Name    string
	Country string
	Website string
}

// PartsFilter: an empty field means "do not filter by this field".
type PartsFilter struct {
	Uuids                 []string
	Names                 []string
	Categories            []Category
	ManufacturerCountries []string
	Tags                  []string
}
