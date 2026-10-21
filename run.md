Run all services:

go run ./inventory/cmd/grpc_server   # gRPC :50051, HTTP gateway + Swagger :8081
go run ./payment/cmd/grpc_server     # gRPC :50052, HTTP gateway + Swagger :8082
go run ./order/cmd/http_server       # HTTP :8080


Test:

task test-api
