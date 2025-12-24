# Performance Optimization Strategies
↩️ [Return to README](../../README.md)

The current architecture uses a monolithic Go service

## Caching
Introduce caching layers to reduce database load and improve response times:
We could use an In-memory caching (Redis), or use AWS ElastiCache for managed Redis. Through this we can cache:
1. Frequently accessed read-heavy data (user balances, transaction summaries). Maybe some VIP users data can be cached for faster access.
2. Session storage if needed.
3. Rate-limiting counters. El rate limiting necesita contar cuántas requests hizo un cliente (por IP, API key, user ID, etc.) en un período de tiempo (ej: 100 requests por minuto). Si lo hicieras solo contra la base de datos, sería lento y generaría mucha carga.
This counter works using Redis as an incremental counter, having a key per user or IP address with an expiration time. Some typical operation are INCR key -> increment the counter. EXPIRE key 60 -> expire in 60 seconds (1 minute window), or even better using INCR + check if it exceeds the limit; if the key does not exist, create it with EXPIRE.

There are other more advanced rate limiting algorithms like sliding window or token bucket that can also be implemented with Redis.

The most used pattern would be the `Cache-aside pattern` in application code, where we check cache first, fall back to database on miss, and update cache on write. A good candidate for caching would be the **idempotency keys** results, to avoid hitting the database for repeated requests, they could have TTL of 24 hours or so.

## Database Optimization
Currently, we have a Single PostgreSQL cluster we could obtain an immediate improvement if we use read replicas for read-heavy operations (separate writer and reader endpoints in application).
Perhaps consider using CQRS pattern to separate read and write models if read load is significantly higher. Even consider adding Elasticsearch for complex read queries and analytics.
Also, we could implement partitioning strategies for very large tables (e.g., transactions) to improve query performance.

## Additional Scalability Enhancements
### Asynchronous processing
Already implemented via Outbox + SQS workers. Scale worker tasks independently based on queue depth.
### Eventual consistency
Accept eventual consistency for non-critical reads (e.g., metrics, audit logs) to reduce write contention.
### Bulkhead pattern and Microservices
Isolate critical paths (e.g., payment processing) in separate ECS services if failure domains need stricter separation. `We try to break the monolith into smaller services over time`, implementing microservices a more event driven architecture, and considering other patterns like Saga for distributed transactions, and eventual consistency.
This would allow independent scaling of components based on load, and would consider to distribute the database layer as well.
### Observability
Instrument key latency and throughput metrics to trigger scaling and detect bottlenecks early. This way we can set SLOs(Service Level Objective) and SLIs(Service Level Indicator) for performance monitoring.
