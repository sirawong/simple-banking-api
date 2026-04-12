-- Seed: 5 mock users for testing
-- Password for all users: Password@123

INSERT INTO users (id, name, email, password_hash, created_at, updated_at)
VALUES
  (
    'a1000000-0000-0000-0000-000000000001',
    'Alice Johnson',
    'alice@example.com',
    '$2a$10$3PZryzDgOx2fdC6TBLxZJOFGPePscCuzfVcKcGHFLl.O8rpsVJFa2',
    NOW(), NOW()
  ),
  (
    'a1000000-0000-0000-0000-000000000002',
    'Bob Smith',
    'bob@example.com',
    '$2a$10$HSMN2/GaiGPKzIUxa8HS5.Nkwz93nkqClvYlUHCDh6q9Zt61KiOLi',
    NOW(), NOW()
  ),
  (
    'a1000000-0000-0000-0000-000000000003',
    'Carol White',
    'carol@example.com',
    '$2a$10$Eka2MXvGjAx234PozK./ROy.7IU2.RWy000TFi.PQZnkBpgN5wEWy',
    NOW(), NOW()
  ),
  (
    'a1000000-0000-0000-0000-000000000004',
    'David Brown',
    'david@example.com',
    '$2a$10$HuneZ87.bvk9j81b.7FgSOP/16UUb9MoROBsDVYHE8nLQvyv8wg2C',
    NOW(), NOW()
  ),
  (
    'a1000000-0000-0000-0000-000000000005',
    'Eve Davis',
    'eve@example.com',
    '$2a$10$I7nlLZKm784EuixtNnejF.CI7EHB1ELV8MYfYc1khHgRj3QuNntL6',
    NOW(), NOW()
  )
ON CONFLICT (id) DO NOTHING;
