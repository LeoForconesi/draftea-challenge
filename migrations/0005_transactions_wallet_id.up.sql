-- 0005_transactions_wallet_id.up.sql
-- Add wallet_id to transactions and enforce relationship.

ALTER TABLE transactions ADD COLUMN IF NOT EXISTS wallet_id VARCHAR(36);

UPDATE transactions t
SET wallet_id = w.id
FROM wallets w
WHERE w.user_id = t.user_id;

ALTER TABLE transactions
ALTER COLUMN wallet_id SET NOT NULL;

ALTER TABLE transactions
ADD CONSTRAINT fk_transactions_wallet
FOREIGN KEY (wallet_id) REFERENCES wallets(id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_transactions_wallet_id ON transactions(wallet_id);
