-- 0002_seed.up.sql
-- Seed initial data for demo/local usage.

INSERT INTO wallets (id, user_id)
VALUES
  ('11111111-1111-1111-1111-111111111111', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'),
  ('22222222-2222-2222-2222-222222222222', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb')
ON CONFLICT DO NOTHING;

INSERT INTO wallet_balances (id, wallet_id, user_id, currency, current_balance)
VALUES
  ('33333333-3333-3333-3333-333333333333', '11111111-1111-1111-1111-111111111111', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'USD', 100000),
  ('44444444-4444-4444-4444-444444444444', '11111111-1111-1111-1111-111111111111', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'MXN', 250000),
  ('55555555-5555-5555-5555-555555555555', '22222222-2222-2222-2222-222222222222', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'USD', 50000)
ON CONFLICT DO NOTHING;

INSERT INTO transactions (id, user_id, type, amount, currency, status, provider_id, external_reference)
VALUES
  ('66666666-6666-6666-6666-666666666666', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'PAYMENT', 15000, 'USD', 'APPROVED', '99999999-9999-9999-9999-999999999999', 'seed-payment-1'),
  ('77777777-7777-7777-7777-777777777777', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'REFUND', 15000, 'USD', 'APPROVED', '99999999-9999-9999-9999-999999999999', 'seed-refund-1'),
  ('88888888-8888-8888-8888-888888888888', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'PAYMENT', 20000, 'USD', 'DECLINED', '99999999-9999-9999-9999-999999999999', 'seed-payment-2')
ON CONFLICT DO NOTHING;
