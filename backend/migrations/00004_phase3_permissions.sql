-- +goose Up
-- +goose StatementBegin

-- ─── Add missing columns ───────────────────────────────────────────────────────
ALTER TABLE suppliers ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
ALTER TABLE categories ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_suppliers_active ON suppliers(is_active) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_categories_store ON categories(store_id) WHERE deleted_at IS NULL;

-- ─── Phase 3 permissions (matches router permission strings exactly) ──────────
-- Remove old purchase_orders.* and reports.view names, insert correct names
DELETE FROM role_permissions
  WHERE permission_id IN (
    SELECT id FROM permissions
    WHERE name IN (
      'purchase_orders.create','purchase_orders.read',
      'purchase_orders.update','purchase_orders.receive',
      'reports.view','reports.export'
    )
  );

DELETE FROM permissions
  WHERE name IN (
    'purchase_orders.create','purchase_orders.read',
    'purchase_orders.update','purchase_orders.receive',
    'reports.view','reports.export'
  );

-- Insert correct permission names (matching router withPerm() calls)
INSERT INTO permissions (id, name, description) VALUES
  (uuid_generate_v4(), 'purchases.create',  'Create purchase orders'),
  (uuid_generate_v4(), 'purchases.read',    'View purchase orders'),
  (uuid_generate_v4(), 'purchases.update',  'Update purchase orders'),
  (uuid_generate_v4(), 'purchases.receive', 'Receive purchase orders into stock'),
  (uuid_generate_v4(), 'purchases.delete',  'Cancel purchase orders'),
  (uuid_generate_v4(), 'suppliers.create',  'Create suppliers'),
  (uuid_generate_v4(), 'suppliers.update',  'Update suppliers'),
  (uuid_generate_v4(), 'suppliers.delete',  'Soft-delete suppliers'),
  (uuid_generate_v4(), 'reports.read',      'View reports and analytics');

-- superadmin & admin: all permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name IN ('superadmin', 'admin')
  AND p.name IN (
    'purchases.create','purchases.read','purchases.update',
    'purchases.receive','purchases.delete',
    'suppliers.create','suppliers.update','suppliers.delete',
    'reports.read'
  )
ON CONFLICT DO NOTHING;

-- manager: purchases + reports, no supplier delete
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'manager'
  AND p.name IN (
    'purchases.create','purchases.read','purchases.update','purchases.receive',
    'suppliers.create','suppliers.update',
    'reports.read'
  )
ON CONFLICT DO NOTHING;

-- cashier: create + read transactions only (no purchases/reports)
-- (already has transactions.create, transactions.read from seed 2)

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM role_permissions
  WHERE permission_id IN (
    SELECT id FROM permissions WHERE name IN (
      'purchases.create','purchases.read','purchases.update',
      'purchases.receive','purchases.delete',
      'suppliers.create','suppliers.update','suppliers.delete',
      'reports.read'
    )
  );
DELETE FROM permissions WHERE name IN (
  'purchases.create','purchases.read','purchases.update',
  'purchases.receive','purchases.delete',
  'suppliers.create','suppliers.update','suppliers.delete',
  'reports.read'
);
ALTER TABLE suppliers DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE categories DROP COLUMN IF EXISTS deleted_at;
-- +goose StatementEnd
