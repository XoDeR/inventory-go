module inventory-go/inventory

go 1.27.2

require (
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.26.3
	github.com/stretchr/testify v1.12.1
	go.mongodb.org/mongo-driver/v2 v2.9.2
	google.golang.org/grpc v1.84.0
	google.golang.org/protobuf v1.36.12
	inventory-go/shared v0.0.0
)

require (
	github.com/envoyproxy/protoc-gen-validate v1.3.3 // indirect
	github.com/klauspost/compress v1.19.2 // indirect
	github.com/stretchr/objx v0.5.3 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.2.0 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/youmark/pkcs8 v0.0.0-20240726163527-a2c0da244d78 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/crypto v0.55.0 // indirect
	golang.org/x/net v0.58.0 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/text v0.41.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20261019180522-ce908d29299b // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20261019180522-ce908d29299b // indirect
)

replace inventory-go/shared => ../shared
