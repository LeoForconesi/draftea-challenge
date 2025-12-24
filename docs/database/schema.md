# Database Schema

## Relation diagram
![database schema](docs/database/Db-relation-diagram.svg)

## Tables
### wallets
- id (varchar(36), PK)
- user_id (varchar(36), unique)
- name (char(20), nullable)
- created_at (timestamptz)

### wallet_balances
- id (varchar(36), PK)
- wallet_id (varchar(36), FK -> wallets.id)
- user_id (varchar(36))
- currency (varchar(8))
- current_balance (bigint)
- created_at, updated_at (timestamptz)
- unique(user_id, currency)

### transactions
- id (varchar(36), PK)
- wallet_id (varchar(36), FK -> wallets.id)
- user_id (varchar(36))
- type (varchar(32))
- amount (bigint)
- currency (varchar(8))
- status (varchar(32))
- provider_id (varchar(36))
- external_reference (text)
- created_at, updated_at (timestamptz)
- indexes: user_id, created_at

### idempotency_records
- id (varchar(36), PK)
- user_id (varchar(36))
- key (text)
- request_id (varchar(36))
- response (jsonb)
- created_at (timestamptz)
- unique(user_id, key)

### outbox
- id (varchar(36), PK)
- event_type (varchar(128))
- payload (jsonb)
- created_at (timestamptz)
- sent_at (timestamptz, nullable)

## Transactions & Consistency
- Payments use DB transactions with row locks on wallet balances.
- The outbox event is written in the same transaction as business state changes.
- Relay publishes outbox records to RabbitMQ and marks them sent.
