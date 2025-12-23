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
