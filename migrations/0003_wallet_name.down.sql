-- 0003_wallet_name.down.sql
-- Remove wallet name column.

ALTER TABLE wallets DROP COLUMN IF EXISTS name;
