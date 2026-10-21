module inventory-go/inventory

go 1.27.2

require (
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.26.3
	google.golang.org/grpc v1.84.0
	google.golang.org/protobuf v1.36.12
	inventory-go/shared v0.0.0
)

require (
	github.com/envoyproxy/protoc-gen-validate v1.3.3 // indirect
	golang.org/x/net v0.58.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/text v0.41.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20261019180522-ce908d29299b // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20261019180522-ce908d29299b // indirect
)

replace inventory-go/shared => ../shared
