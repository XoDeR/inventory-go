# NotificationService contracts

NotificationService only consumes events and produces none.

| Event | Topic | Consumer group (env) |
|---|---|---|
| OrderPaid | `order.paid` | `notification-service-order-paid` (`ORDER_PAID_CONSUMER_GROUP_ID`) |
| ComputerAssembled | `assembly.computer-assembled` | `notification-service-computer-assembled` (`COMPUTER_ASSEMBLED_CONSUMER_GROUP_ID`) |

Contracts: [order_service_contracts.md](order_service_contracts.md), [assembly_service_contracts.md](assembly_service_contracts.md).
