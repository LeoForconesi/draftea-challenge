# System Overview
↩️ [Return to README](../../README.md)
## High-Level Architecture
![high-level Architecture](docs/architecture/High-level-diagram.svg)
<details>
<summary>Mermaid Code</summary>

```mermaid
flowchart LR
  Client[API Client] -->|HTTP/JSON| API[API Service (Gin)]
  API -->|Usecases| App[Application Layer]
  App -->|Repos| DB[(PostgreSQL)]
  App -->|Gateway Client| Gateway[Mock Payment Gateway]
  App -->|Outbox Writes| DB
  Relay[Outbox Relay] -->|Read Outbox| DB
  Relay -->|Publish| MQ[(RabbitMQ)]
  MQ --> Metrics[metrics-consumer]
  MQ --> Audit[audit-consumer]
```
</details>



## Request Flow (Payment)
![request flow](docs/architecture/Request Flow.svg)

<details>
<summary>Mermaid Code</summary>

```mermaid
sequenceDiagram
  participant C as Client
  participant API as API Service
  participant DB as PostgreSQL
  participant GW as Mock Gateway
  participant R as Outbox Relay
  participant MQ as RabbitMQ

C->>API: POST /wallets/{user_id}/payments (Idempotency-Key)
API->>DB: Tx: lock wallet balance, validate funds, create tx
API->>GW: Process payment
GW-->>API: status
API->>DB: Tx: update tx status, adjust balance if refund
API->>DB: Tx: insert outbox event
API-->>C: 200/4xx/5xx
R->>DB: fetch pending outbox
R->>MQ: publish event
```
</details>


## Layers (Clean/Hexagonal)
![layers](docs/architecture/Layers(clean-hexagonal).svg)

<details>
<summary>Mermaid Code</summary>

```mermaid
flowchart TB
  subgraph Adapters
    HTTP[HTTP Handlers] --> UC[Usecases]
    PG[Postgres Repos] --> UC
    GW[Gateway Client] --> UC
    MQ[Outbox Publisher] --> UC
  end
  subgraph Domain
    Entities[Wallet/Transaction/Payment]
    Errors[Domain Errors]
  end
  UC --> Entities
  UC --> Errors
```
</details>

## External Integrations
- Mock Payment Gateway via HTTP
- RabbitMQ for async domain events (outbox relay)
