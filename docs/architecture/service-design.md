# Service Design
↩️ [Return to README](../../README.md)

## Modules and Responsibilities
- API (HTTP handlers): parse/validate requests, map errors, call usecases.
- Application/usecases: orchestrate payment flow, idempotency, and transactions.
- Domain: entities (Wallet, Transaction, Payment) and domain rules/errors.
- Persistence (Postgres): repositories with locking and transaction semantics.
- Gateway client: calls mock payment provider with retries/circuit breaker.
- Outbox relay: publishes outbox events to RabbitMQ.
- Consumers: metrics/audit listeners.

## Endpoints
### POST /wallets/{user_id}/payments
Request:
```json
{
  "provider_id": "uuid",
  "external_reference": "string",
  "amount": 1000,
  "currency": "USD"
}
```
Response:
```json
{
  "transaction_id": "uuid",
  "status": "APPROVED"
}
```

### GET /wallets/{user_id}/balance
Response:
```json
{
  "user_id": "uuid",
  "balances": {
    "USD": 1000
  }
}
```

### GET /wallets/{user_id}/transactions
Response:
```json
{
  "transactions": [
    {
      "id": "uuid",
      "user_id": "uuid",
      "type": "PAYMENT",
      "amount": 1000,
      "currency": "USD",
      "status": "APPROVED",
      "provider_id": "uuid",
      "external_reference": "string",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "total": 1
}
```

## Test-Only Endpoints
These endpoints exist to simplify local testing and visibility:

### GET /healthz
Returns `{ "status": "ok" }`.

### POST /wallets/{user_id}/top-up
Adds balance to a wallet for testing.

### GET /wallets
Lists wallets and balances (paginated).

### POST /wallets
Creates a wallet for a user (test-only).

Production note: funding a wallet would be handled by a separate service with proper user identity, KYC/AML, and ledgering. Top-ups here are for test visibility only.

## Outbox Retention and Retry
- Outbox relay retries publish failures with backoff.
- Outbox rows should be cleaned or archived once sent to prevent unbounded growth.

## Domain Models (Summary)
- Wallet: user_id + balances per currency.
- Transaction: immutable ledger entries with status.
- Payment: payment intent with provider and external reference.

## Layering
- Domain has no infrastructure dependencies.
- Usecases depend on ports (interfaces).
- Adapters implement those ports and are wired in factories.

## API Key (Test Only)
- Static `X-API-Key` is supported for local testing.
- Production should use proper authN/authZ (e.g., OAuth2/JWT + user service).
