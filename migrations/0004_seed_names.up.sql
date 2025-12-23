-- 0004_seed_names.up.sql
-- Add names to seeded wallets after name column exists.

UPDATE wallets
SET name = 'Arthur Schoppenhauer'
WHERE id = '11111111-1111-1111-1111-111111111111';

UPDATE wallets
SET name = 'Albert Camus'
WHERE id = '22222222-2222-2222-2222-222222222222';
