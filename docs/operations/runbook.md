# Runbook

## Local
- `make up`
- `make migrate`
- `go run cmd/api/main.go`
- `go run cmd/relay/main.go`
- `go run cmd/consumer/main.go`

## Migrations
- `make migrate-up`
- `make migrate-down N=1`
- `make migrate-version`
- `make migrate-force V=2`

## Docker Compose
- `make up`
- `make migrate`
- Logs: `make logs`

## Testing
- `go test ./...`

## OpenAPI
- Spec: `docs/openapi.yaml`

## API Key
If `app.api_key` is set, include the header:
```
X-API-Key: <your-api-key>
```

## CORS
The API allows cross-origin requests (CORS) with:
- Origins: `*`
- Methods: `GET`, `POST`, `OPTIONS`
- Headers: `Content-Type`, `Idempotency-Key`, `X-API-Key`, `X-Request-ID`
- Exposed headers: `X-Request-ID`

## Test-Only Endpoints
- `GET /healthz`
- `POST /wallets/{user_id}/top-up`
- `GET /wallets`
- `POST /wallets`

## Outbox Retention and Retry
- The relay retries publish failures with exponential backoff.
- The outbox table is append-only; in production you should clean or archive sent events.
  - Example: delete `sent_at IS NOT NULL` rows older than N days.
  - Or archive to a cold table for audit purposes.
  - For high volume, partition by time.
