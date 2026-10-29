Generate the .env files (once, and after any change to deploy/env/.env):

task env:generate   # deploy/env/.env.template -> deploy/env/.env -> deploy/compose/<service>/.env

Start the infrastructure (the up-* tasks run env:generate themselves):

task up-all

Run all services from the repository root:

go run ./inventory/cmd/grpc_server   # gRPC :50051, HTTP gateway + Swagger :8081
go run ./payment/cmd/grpc_server     # gRPC :50052, HTTP gateway + Swagger :8082
go run ./order/cmd/http_server       # HTTP :8080


Test:

task test-api
