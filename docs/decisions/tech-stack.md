# Tech Stack
↩️ [Return to README](../../README.md)
## Language
- Go 1.24.11: simple concurrency, strong tooling, and performance.

## Web Framework
- Gin: minimal, fast routing with middleware support.

## Database
- PostgreSQL: transactional integrity and row-level locks.
- GORM + pgx driver: pragmatic data access with transactions and SQL support.

## Messaging
- RabbitMQ: topic exchange for payment/refund events.
- Outbox pattern to avoid dual-write and enable retries.

## Config
- YAML config via Viper + env overrides with envconfig.

## Logging
- Zap structured logger with levels.

## Testing
- Go test with mocks and builders.

## Containers
- Docker + docker-compose for local orchestration.
