# OrderService contracts

## Produced events

### OrderPaid

Published by OrderService right after an order was successfully paid (status `PAID` is saved).

| | |
|---|---|
| Topic | `order.paid` (env `ORDER_PAID_TOPIC_NAME`) |
| Key | `order_uuid` |
| Encoding | JSON |
| Consumers | AssemblyService, NotificationService |

| Field | Type | Description |
|---|---|---|
| `event_uuid` | string (uuid) | Unique event id, for deduplication |
| `order_uuid` | string (uuid) | Paid order |
| `user_uuid` | string (uuid) | Owner of the order |
| `payment_method` | string | `CARD`, `CREDIT_CARD`, `STORE_CREDIT`, ... |
| `transaction_uuid` | string (uuid) | Payment transaction id |

```json
{
  "event_uuid": "6f1d1a2e-3b0a-4a37-9c66-1b6f6c2f0a11",
  "order_uuid": "0b9a0a0e-5d2c-4e53-8d75-0c7c5a1d9e10",
  "user_uuid": "a3c1c7de-6d58-4a0e-8f70-5f1c3f0b2d44",
  "payment_method": "CARD",
  "transaction_uuid": "d1f9a5f0-8f2e-4d9b-9a53-2a5c7d2b8e77"
}
```

## Consumed events

### ComputerAssembled

See [assembly_service_contracts.md](assembly_service_contracts.md).
OrderService sets the status of the order `order_uuid` to `COMPLETED`.
Consumer group: `order-service` (env `ORDER_ASSEMBLED_CONSUMER_GROUP_ID`).
