# draftea-challenge

## Quickstart
- `make up`
- `make migrate`
- `make logs`

## Migrations
Migrations use `golang-migrate` with versioned SQL files under `migrations/`:
- Apply: `make migrate` (uses `docker compose run --rm migrate up`)
- Rollback: `docker compose run --rm migrate down 1`

## Docker Compose Flow
- `make up` starts Postgres, RabbitMQ, mock gateway, and the API.
- `make migrate` runs the migration container against the Postgres service.

## OpenAPI
- Spec: `docs/openapi.yaml`

## Configuration
Configuration precedence is:
1) YAML config at the path specified by `CONFIG_FILE` (optional).
2) Environment variables override YAML values.
3) Built-in defaults.

Included files:
- `config/config.local.yaml` for running locally.
- `config/config.docker.yaml` for docker-compose.
- Docker Compose sets `CONFIG_FILE=/app/config/config.docker.yaml` directly.
