# Database Schema
↩️ [Return to README](../../README.md)

## Relation diagram
![database schema](docs/resources/Db-relation-diagram.svg)

<details>
<summary>Code DBML of the schema</summary>

Copy and paste the following code into https://dbdiagram.io/d to visualize the database schema or modify it.

```DBML
Table wallets {
id varchar(36) [pk]
user_id varchar(36) [unique, not null]
name char(20)
created_at timestamptz [not null]
}

Table wallet_balances {
id varchar(36) [pk]
wallet_id varchar(36) [not null]
user_id varchar(36) [not null]
currency varchar(8) [not null]
current_balance bigint [not null, default: 0]
created_at timestamptz [not null]
updated_at timestamptz [not null]

Indexes {
(user_id, currency) [unique]
}
}

Table transactions {
id varchar(36) [pk]
wallet_id varchar(36) [not null]
user_id varchar(36) [not null]
type varchar(32) [not null]
amount bigint [not null]
currency varchar(8) [not null]
status varchar(32) [not null]
provider_id varchar(36)
external_reference text
created_at timestamptz [not null]
updated_at timestamptz [not null]

Indexes {
wallet_id
user_id
created_at
}
}

Table idempotency_records {
id varchar(36) [pk]
user_id varchar(36) [not null]
key text [not null]
request_id varchar(36) [not null]
response jsonb [not null]
created_at timestamptz [not null]

Indexes {
(user_id, key) [unique]
}
}

Table outbox {
id varchar(36) [pk]
event_type varchar(128) [not null]
payload jsonb [not null]
created_at timestamptz [not null]
sent_at timestamptz
}

Ref: wallet_balances.wallet_id > wallets.id
Ref: transactions.wallet_id > wallets.id
```

</details>

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
