-- 0003_wallet_name.up.sql
-- Add optional wallet name for testing visibility.

ALTER TABLE wallets ADD COLUMN IF NOT EXISTS name CHAR(20);
