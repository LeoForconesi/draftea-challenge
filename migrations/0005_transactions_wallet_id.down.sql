-- 0005_transactions_wallet_id.down.sql
-- Remove wallet_id from transactions.

ALTER TABLE transactions DROP CONSTRAINT IF EXISTS fk_transactions_wallet;
DROP INDEX IF EXISTS idx_transactions_wallet_id;
ALTER TABLE transactions DROP COLUMN IF EXISTS wallet_id;
