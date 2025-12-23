-- 0001_init.down.sql
-- Drop initial schema.

DROP TABLE IF EXISTS outbox;
DROP TABLE IF EXISTS idempotency_records;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS wallet_balances;
DROP TABLE IF EXISTS wallets;
