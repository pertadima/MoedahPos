-- +goose Up
-- +goose StatementBegin

-- ─── Fix: manager can create transactions (POS/cashier access) ────────────────
-- The transactions.create permission allows POST /transactions (checkout).
-- Manager role needs this to operate the cashier / POS page.
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'manager'
  AND p.name = 'transactions.create'
ON CONFLICT DO NOTHING;

-- ─── Fix: cashier can update table availability status ────────────────────────
-- PUT /tables/{id}/status is gated by "products.update".
-- Cashier role needs products.update so they can set tables to
-- available / occupied / reserved during service.
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'cashier'
  AND p.name = 'products.update'
ON CONFLICT DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DELETE FROM role_permissions
WHERE role_id  = (SELECT id FROM roles WHERE name = 'manager')
  AND permission_id = (SELECT id FROM permissions WHERE name = 'transactions.create');

DELETE FROM role_permissions
WHERE role_id  = (SELECT id FROM roles WHERE name = 'cashier')
  AND permission_id = (SELECT id FROM permissions WHERE name = 'products.update');

-- +goose StatementEnd
