# Project Context — Draftea Backend Challenge (Go)

## Goal
Build a production-like wallet & payments API in Go (latest stable), dockerized, with Makefile workflows.
Focus: clean architecture/hexagonal + basic DDD, SOLID, clear boundaries, tests, idempotency, concurrency handling, and good error handling.

## Non-Goals (Explicitly Out of Scope)
- No real authentication/authorization system (only fake API key)
- No real external payment provider
- No cloud deployment
- No schema migrations framework beyond simple SQL
- No eventual consistency for wallet balance (balance updates are synchronous)

## Database Migrations (MANDATORY)

### Tooling
We use **golang-migrate/migrate** for PostgreSQL migrations.
- Migrations are **SQL files** only.
- We do **NOT** run migrations via custom Go scripts (no `migrate.go`, no GORM auto-migrate).
- The application must **not** apply schema changes at runtime.

### Folder structure
All migrations live in:
- `./migrations/`

File naming convention is mandatory:
- `NNNN_<description>.up.sql`
- `NNNN_<description>.down.sql`

Examples:
- `0001_init_schema.up.sql`
- `0001_init_schema.down.sql`
- `0002_add_transactions_table.up.sql`
- `0002_add_transactions_table.down.sql`

### Version tracking
`migrate` manages schema versioning using the DB table:
- `schema_migrations` (version + dirty)

Never manually edit this table unless doing a recovery (`force`) and documenting it.

### Execution model
Migrations are executed **outside** the app process:
- Locally via Docker Compose service `migrate`
- In CI/CD as a dedicated step before running the app / tests

The expected flow is:
1. Start DB
2. Run migrations
3. Start app

### Docker Compose
A dedicated service named `migrate` must exist and use the official image:
- `migrate/migrate`

It must mount the migrations folder and point to Postgres via `DATABASE_URL` (or equivalent).

### Makefile commands (required)
The Makefile must provide these targets:

- `make migrate-up`  
  Runs all pending migrations.

- `make migrate-down`  
  Rolls back one migration (or supports `N=` param if implemented).

- `make migrate-version`  
  Prints current schema version.

- `make migrate-force`  
  Used only to recover from a dirty state (must require explicit `V=` version).

### Prohibited patterns
- No “run all SQL files” loops.
- No filesystem-order-based execution without versioning.
- No migrations mixed with app code under `main` package.
- No schema changes on app startup (except a *health check* or *read-only* version check).

---

## Deliverables
- REST API:
    - POST /wallets/{user_id}/payments (idempotent with Idempotency-Key)
    - GET  /wallets/{user_id}/balance
    - GET  /wallets/{user_id}/transactions (paginated)
- PostgreSQL schema + SQL migrations (run by init container)
- External mock payment gateway (HTTP server in docker-compose) with failure modes
- Unit tests for critical usecases/services (aim ~80% where reasonable)
- Dockerfile + docker-compose (app + postgres + mock-gateway)
- Makefile targets: up/down/reset/migrate/test/lint/fmt/clean/logs
- README:
    - architecture diagrams (Mermaid)
    - decisions and tradeoffs
    - how to run, test, and examples of requests/responses

---

## Tech Stack (Decided)
- Language: Go (latest stable)
- HTTP Framework: Gin
- DB: PostgreSQL
- ORM: GORM using pgx driver
- Config:
    - Env vars (prod/dev)
    - Local YAML config support via Viper
    - envconfig for env binding
    - App can run locally (connecting to docker services) or inside docker
- Validation: manual (explicit checks) for clarity
- Logging: structured logger (zap) with levels debug/info/warn/error
- CI (optional but valued): GitHub Actions with golangci-lint + tests + build
- Messaging: RabbitMQ (docker-compose)
    - Publish domain events via Outbox pattern (DB) and relay to RabbitMQ
    - Consumer(s) for metrics/audit (separate containers/processes)
---

## Architecture Principles (Clean / Hexagonal + Basic DDD)
Layers:
- `domain`:
    - entities/value objects: Wallet, Balance, Transaction, Payment
    - domain rules and domain errors
- `application` (usecases):
    - orchestrates payment flow, idempotency, DB transactions, gateway call
- `adapters`:
    - `http` (Gin handlers + middlewares)
    - `persistence` (Postgres/GORM repositories)
    - `gateway` (HTTP client to mock-gateway)
- `platform`:
    - config, logger, db, server wiring (composition root)

Rules:
- Domain must not import infrastructure packages.
- Use interfaces at boundaries: repositories, gateway, clock, id generator.
- Handlers are thin; business logic is in usecases/domain.
- Always pass `context.Context` end-to-end.
- Use typed error codes and consistent error responses.

---

## Domain Model (Target)
### Multi-currency balances
- Represent money as BIGINT minor units (int64).
- Wallet balances are per user_id + currency:
    - table `wallet_balances` with `current_balance` per currency
- Ledger of transactions as immutable-ish history:
    - table `transactions`

### Service Provider modeling (Payments)
Payments reference:
- `provider_id` (internal UUID) -> `service_providers`
- `external_reference` (string; e.g., bill/customer reference)
- `amount`, `currency`

---

## Data Consistency & Concurrency
### Concurrency requirement
Two payments can happen simultaneously for same wallet/currency.
Approach:
- Use DB transaction + row lock:
    - `SELECT ... FOR UPDATE` on `wallet_balances` for (user_id, currency)
    - validate sufficient funds
    - update `current_balance`
    - insert transaction record
    - commit

### Idempotency
POST /payments must accept `Idempotency-Key` header.
- Store idempotency record keyed by (user_id, idempotency_key)
- If repeated:
    - return same response as original (same payment/transaction id + status)
- Ensure uniqueness constraint in DB.
---

## Patterns: Factory + Builder (for testability)
We will use patterns intentionally and sparingly:
Factories = composition root (runtime wiring)
Builders = solo tests

### Factories (composition root)
- Factories build runtime components with real dependencies:
    - DB connection, repositories, gateway client, outbox relay, Rabbit publisher, HTTP server
- Factory package must be the only place wiring concrete implementations.
- Expose constructors returning interfaces when possible.

### Builders (test helpers + complex object creation)
- Builders are primarily for tests to create valid domain objects and requests:
    - PaymentRequestBuilder
    - WalletBuilder
    - TransactionBuilder
- Builders must live in `internal/testsupport` or `test/` and not leak into production domain logic.

Goal:
- Production uses explicit constructors + validation.
- Tests use builders to keep them readable and to reduce setup boilerplate.

---

## Payment Flow (Synchronous)
1) Parse request, validate.
2) Start DB tx:
    - lock wallet_balance row FOR UPDATE
    - check funds
    - create a transaction record with status=PENDING (or RESERVED)
    - update wallet_balance (debit or "hold" strategy; chosen approach: debit in tx)
3) Call external mock-gateway HTTP:
    - modes: happy / timeout / 500 / declined / latency / random
4) Start DB tx to finalize:
    - update transaction status: APPROVED/DECLINED/FAILED
    - if gateway failed and funds were debited, perform internal refund/compensation:
        - create REFUND transaction and restore balance (internal operation)
5) Return API response.

Note: For simplicity, keep synchronous. Mention in README that production may use Outbox + async processing, queues, retries, DLQ, etc.

---

## Async Events (RabbitMQ + Outbox)
We will avoid dual-write (DB + Rabbit) using Transactional Outbox:

- Within the same DB transaction that finalizes a payment:
    - write business data (transactions, balances)
    - write an outbox event record

- An Outbox Relay worker reads pending outbox rows and publishes to RabbitMQ.
- After successful publish, marks outbox row as SENT (or deletes it).
- If RabbitMQ is down, events remain in DB and will be retried later.

RabbitMQ exchanges/queues:
- Exchange: payments.events (topic)
- Routing keys:
    - payment.created
    - payment.completed
    - payment.failed
    - refund.created
- Queues:
    - metrics.queue (bind payment.*)
    - audit.queue (bind payment.* and refund.*)

Consumers:
- metrics-consumer: increments counters/logs (for interview + future observability)
- audit-consumer: stores audit logs or just logs structured entries

Note: Core HTTP flow remains synchronous; messaging is for side-effects and observability.

---

## Error Handling Contract
All errors must be consistent JSON:
{
"error": {
"code": "INSUFFICIENT_FUNDS",
"message": "...",
"details": { ... }
}
}

HTTP mapping (suggested):
- VALIDATION_ERROR -> 400
- UNAUTHORIZED (bad api key) -> 401
- NOT_FOUND -> 404
- INSUFFICIENT_FUNDS -> 409
- GATEWAY_TIMEOUT -> 504
- GATEWAY_ERROR -> 502
- INTERNAL -> 500

Never leak raw internal errors to clients; log with context.

---

## Resilience Patterns (Gateway + Rabbit)
### Payment Gateway Client
- Context timeouts
- Retry with exponential backoff + jitter for transient errors (timeout/5xx)
- No retry on declined/logical failures
- Circuit breaker (fail-fast when gateway unhealthy)
- Bulkhead (limit concurrent gateway calls via semaphore/worker pool)

Bulkheads / worker pools are used:
- to limit concurrent calls to the payment gateway
- to limit concurrent RabbitMQ publishing in outbox relay

### Rabbit Publisher (Outbox Relay)
- Retry with backoff when publish fails
- Idempotent publish strategy via outbox unique event_id
- Publisher confirms (if using AMQP confirms) when feasible

---

## Gin Middlewares (Must)
- RequestID (use header or generate)
- Structured logging (method/path/status/latency + request_id)
- Recovery (panic handling)
- Fake API Key auth (X-API-Key)
- Context timeout (set request deadline; important for gateway calls)

---

## Logging (Structured, Levels)
Use zap/zerolog with levels:
- debug: DB tx boundaries, repository calls timing, idempotency hits
- info: request start/end, payment created, gateway result
- warn: validation errors, insufficient funds, declined
- error: gateway timeouts/errors, unexpected failures

Include fields:
request_id, user_id, payment_id/transaction_id, currency, amount, provider_id, external_reference, latency_ms

---

## Project Structure (Recommended)
- cmd/api/main.go
- internal/
    - platform/
        - config/
        - logger/
        - db/
        - server/
    - domain/
        - wallet/
        - transaction/
        - payment/
        - errors/
    - application/
        - payments/
        - wallets/
        - outbox/
    - adapters/
        - messaging
            - rabbitmq/
        - http/
            - handlers/
            - middleware/
            - presenter/
        - persistence/postgres/
        - gateway/httpclient/
- migrations/
- docker-compose.yml
- Dockerfile
- Makefile

---

## Makefile Contract
- make up            # docker compose up -d
- make down          # docker compose down
- make reset         # down -v && up && migrate
- make migrate       # run migrations (init container or explicit)
- make test          # go test ./... with coverage output
- make lint          # golangci-lint run
- make fmt           # gofmt/goimports
- make logs          # docker compose logs -f
- make clean         # remove artifacts
- make rabbit-ui      # optional: opens management UI info (document only)
- make consume        # optional: run consumers locally

---

## Testing
- Unit tests with mocks (interfaces) for usecases:
    - happy path payment
    - insufficient funds
    - user/wallet not found
    - gateway timeout
    - gateway error 500
    - declined
    - idempotency hit
- Aim for ~80% coverage on critical application layer.

## Constraints for Implementation
- Do not introduce new endpoints not listed in this document
- Do not introduce additional infrastructure components
- Do not change API contracts or domain rules unless explicitly stated
- Follow the described data model and flows strictly