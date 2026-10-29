package config

type InventoryGRPCConfig interface {
	Address() string
}

type InventoryHTTPConfig interface {
	Address() string
}

type MongoConfig interface {
	URI() string
	DatabaseName() string
}
