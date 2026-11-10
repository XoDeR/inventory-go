# Redis caching

The inventory and order services cache entities in Redis (cache-aside). The shared Redis client lives in
`platform/pkg/cache` (interfaces) and `platform/pkg/cache/redis` (redigo-based implementation).

## Environment variables

Every service that uses the cache needs the same set of variables (generated from
`deploy/env/<service>.env.template` by `task env:generate`):

```bash
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_CONNECTION_TIMEOUT=5s   # max wait for a connection from the pool
REDIS_MAX_IDLE=10             # idle connections kept in the pool
REDIS_IDLE_TIMEOUT=240s       # an idle connection older than this is closed
REDIS_CACHE_TTL=1h            # lifetime of a cache entry
```

Deployment: each service has its own container (`redis-inventory` on host port 6379, `redis-order` on 6380),
started together with the service dependencies: `task up-core up-inventory up-order`.

## Layers

```
Service layer (cache-aside logic)
    |                 |
Repository (DB)   Cache repository (Redis)
```

- `repository/part` (MongoDB) and `repository/order` (PostgreSQL) only talk to their database.
- `repository/part_cache` and `repository/order_cache` only talk to Redis.
- The service decides when to read, fill and invalidate the cache, so the logic is visible in one place and
  each repository is easy to mock.

## Behavior

| Service   | Operation                      | Cache                                              |
|-----------|--------------------------------|----------------------------------------------------|
| inventory | `Get` (GetPart)                | read the cache, on a miss read MongoDB and fill it |
| inventory | `List`                         | not cached                                         |
| order     | `Get`                          | read the cache, on a miss read PostgreSQL and fill |
| order     | `Pay`, `Cancel`                | update PostgreSQL, then invalidate the entry       |
| order     | ComputerAssembled consumer     | update PostgreSQL, then invalidate the entry       |
| order     | `Create`                       | not cached (the first `Get` fills the cache)       |

Keys: `inventory:part:{uuid}` (gob-encoded, keeps the types of the metadata values) and `order:{uuid}` (JSON).

## Graceful degradation

The Redis pool connects lazily, so a service starts and works without Redis. Cache errors are logged and
ignored; the data is then served from the database. A failed invalidation is also only logged: the stale
entry disappears by `REDIS_CACHE_TTL`.

## Tests

- Service unit tests use mocks of the cache repository (hit, miss, cache failure).
- Cache repository tests need a real Redis and are skipped without `REDIS_TEST_ADDR`
  (e.g. `REDIS_TEST_ADDR=localhost:6379` after `task up-core up-inventory`).
- The e2e tests start a Redis container via `platform/pkg/testcontainers/redis`.
