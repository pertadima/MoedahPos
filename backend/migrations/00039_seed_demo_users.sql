-- +goose Up
-- Seed demo users for each role
INSERT INTO users (id, name, email, password_hash, is_active, created_at, updated_at)
SELECT
  id,
  name,
  email,
  password_hash,
  is_active,
  created_at,
  updated_at
FROM (VALUES
  ('00000000-0000-0000-0000-000000000001'::uuid, 'Admin Demo', 'admin@demo.com', '$2a$12$A4j4exixN/HXyH1yaMQ95untEXyiQcTs4COOcVf8xWQvX6KGYeS92'::bytea, true, now(), now()),
  ('00000000-0000-0000-0000-000000000002'::uuid, 'Kasir Demo', 'kasir@demo.com', '$2a$12$1z9gL0V0v.Bjnyy7VUl91.TsInMkZH5KsNDrIpDk1QjM3CrL7JpyK'::bytea, true, now(), now()),
  ('00000000-0000-0000-0000-000000000003'::uuid, 'Super Admin Demo', 'superadmin@demo.com', '$2a$12$NO7sBKHHf9u.JGwhV5Gi/.A2atdXDwpGAALjzZz0fbIq3UlhSqFkq'::bytea, true, now(), now())
) AS t(id, name, email, password_hash, is_active, created_at, updated_at)
WHERE NOT EXISTS (SELECT 1 FROM users WHERE email = 'admin@demo.com');

-- Assign roles to demo users
INSERT INTO user_roles (user_id, role_id)
SELECT u.id, r.id
FROM users u
CROSS JOIN roles r
WHERE (u.email = 'admin@demo.com' AND r.name = 'admin')
   OR (u.email = 'kasir@demo.com' AND r.name = 'cashier')
   OR (u.email = 'superadmin@demo.com' AND r.name = 'superadmin')
ON CONFLICT DO NOTHING;