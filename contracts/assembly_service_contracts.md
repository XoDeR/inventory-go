# AssemblyService contracts

## Consumed events

### OrderPaid

See [order_service_contracts.md](order_service_contracts.md).
On receiving the event AssemblyService waits 10 seconds (the "assembly") and then publishes `ComputerAssembled`.
Consumer group: `assembly-service` (env `ORDER_PAID_CONSUMER_GROUP_ID`).

## Produced events

### ComputerAssembled

Published by AssemblyService after the delay.

| | |
|---|---|
| Topic | `assembly.computer-assembled` (env `COMPUTER_ASSEMBLED_TOPIC_NAME`) |
| Key | `order_uuid` |
| Encoding | JSON |
| Consumers | OrderService, NotificationService |

| Field | Type | Description |
|---|---|---|
| `event_uuid` | string (uuid) | Unique event id, for deduplication |
| `order_uuid` | string (uuid) | Assembled order |
| `user_uuid` | string (uuid) | Owner of the order |
| `build_time_sec` | int64 | Time spent on the assembly, seconds |

```json
{
  "event_uuid": "9a7e6a52-1b5c-4c0e-bb24-0f4f9e0d6a31",
  "order_uuid": "0b9a0a0e-5d2c-4e53-8d75-0c7c5a1d9e10",
  "user_uuid": "a3c1c7de-6d58-4a0e-8f70-5f1c3f0b2d44",
  "build_time_sec": 10
}
```
