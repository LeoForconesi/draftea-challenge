# draftea-challenge

![Coverage](coverage-badge.svg)

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
- If `CONFIG_FILE` is not set, the app selects a default based on `APP_ENV`:
  - `local` -> `config/config.local.yaml`
  - `docker` -> `config/config.docker.yaml`
  - `stage` -> `config/config.stage.yaml`
  - `prod` -> `config/config.prod.yaml`

## Test-Only Endpoints
- `GET /healthz`
- `POST /wallets/{user_id}/top-up`
- `GET /wallets`

## API Key (Test Only)
If `app.api_key` is set, include the header `X-API-Key: <your-api-key>` on requests.
