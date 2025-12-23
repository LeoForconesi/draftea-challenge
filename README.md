# draftea-challenge

Production-like wallet and payments API built with Go, Gin, PostgreSQL, and RabbitMQ.

## Docs
- [Architecture overview](docs/architecture/system-overview.md)
- [Service design](docs/architecture/service-design.md)
- [Database schema](docs/database/schema.md)
- [Tech stack](docs/decisions/tech-stack.md)
- [Error handling](docs/operations/error-handling.md)
- [Runbook](docs/operations/runbook.md)
- [OpenAPI spec](docs/openapi.yaml)

## Quickstart
- `make up`
- `make migrate`
- `make logs`

## Migrations
Migrations use `golang-migrate` with versioned SQL files under `migrations/`:
- Apply: `make migrate` (uses `docker compose run --rm migrate up`)
- Rollback: `docker compose run --rm migrate down 1`

## Docker Compose Flow
- `make up` starts Postgres, RabbitMQ, mock gateway, API, relay, and consumers.
- `make migrate` runs the migration container against the Postgres service.

## Configuration
Configuration precedence is:
1) YAML config at the path specified by `CONFIG_FILE` (optional).
2) Environment variables override YAML values.
3) Built-in defaults.

Included files:
- `config/config.local.yaml` for running locally.
- `config/config.docker.yaml` for docker-compose.
- Docker Compose sets `CONFIG_FILE=/app/config/config.docker.yaml` directly.
