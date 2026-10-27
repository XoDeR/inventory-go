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

// Part is the MongoDB document of the "parts" collection. The part UUID is used as the document _id.
type Part struct {
	Uuid          string        `bson:"_id"`
	Name          string        `bson:"name"`
	Description   string        `bson:"description,omitempty"`
	Price         float64       `bson:"price"`
	StockQuantity int64         `bson:"stock_quantity"`
	Category      Category      `bson:"category"`
	Dimensions    *Dimensions   `bson:"dimensions,omitempty"`
	Manufacturer  *Manufacturer `bson:"manufacturer,omitempty"`
	Tags          []string      `bson:"tags,omitempty"`
	// Metadata values are one of: string, int64, float64, bool.
	Metadata  map[string]any `bson:"metadata,omitempty"`
	CreatedAt *time.Time     `bson:"created_at,omitempty"`
	UpdatedAt *time.Time     `bson:"updated_at,omitempty"`
}

type Dimensions struct {
	Length float64 `bson:"length"`
	Width  float64 `bson:"width"`
	Height float64 `bson:"height"`
	Weight float64 `bson:"weight"`
}

type Manufacturer struct {
	Name    string `bson:"name"`
	Country string `bson:"country"`
	Website string `bson:"website,omitempty"`
}
