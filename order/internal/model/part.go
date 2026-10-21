package model

// Part is the part of an inventory item that an order needs: its identity and price.
type Part struct {
	Uuid  string
	Price float64
}

type PartsFilter struct {
	Uuids []string
}
