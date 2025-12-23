-- 0004_seed_names.down.sql
-- Remove names from seeded wallets.

UPDATE wallets
SET name = NULL
WHERE id IN (
  '11111111-1111-1111-1111-111111111111',
  '22222222-2222-2222-2222-222222222222'
);
