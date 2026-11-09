Generate the .env files (once, and after any change to deploy/env/.env):

task env:generate   # deploy/env/.env.template -> deploy/env/.env -> deploy/compose/<service>/.env

Start the infrastructure (the up-* tasks run env:generate themselves; core brings up Kafka in KRaft mode + Kafka UI on :8090):

task up-all

Build and run all services in one terminal (Ctrl+C stops them):

task run-all

Or run a single service (each builds first): task run:inventory, run:payment, run:order, run:assembly, run:notification.

Or run them by hand from the repository root, one terminal each:

go run ./inventory/cmd/grpc_server      # gRPC :50051, HTTP gateway + Swagger :8081
go run ./payment/cmd/grpc_server        # gRPC :50052, HTTP gateway + Swagger :8082
go run ./order/cmd/http_server          # HTTP :8080, publishes OrderPaid, consumes ComputerAssembled
go run ./assembly/cmd/consumer          # consumes OrderPaid, waits ASSEMBLY_DURATION (10s), publishes ComputerAssembled
go run ./notification/cmd/consumer      # consumes OrderPaid + ComputerAssembled, posts to Slack, answers /start (needs the NOTIFICATION_SLACK_* values in deploy/env/.env)

Event flow (contracts in contracts/):

order: pay -> OrderPaid -> assembly: wait 10s -> ComputerAssembled -> order: status COMPLETED
                      \-> notification              \-> notification

Test:

task test-api
