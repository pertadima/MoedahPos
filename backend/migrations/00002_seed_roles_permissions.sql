-- +goose Up
-- +goose StatementBegin

-- ─── Seed Roles ───────────────────────────────────────────────────────────────
INSERT INTO roles (id, name, description) VALUES
    (uuid_generate_v4(), 'superadmin', 'Full system access across all stores'),
    (uuid_generate_v4(), 'admin',      'Full access within assigned stores'),
    (uuid_generate_v4(), 'manager',    'Manage products, stock, and reports'),
    (uuid_generate_v4(), 'cashier',    'Process transactions only'),
    (uuid_generate_v4(), 'staff',      'View-only access');

-- ─── Seed Permissions ─────────────────────────────────────────────────────────
INSERT INTO permissions (id, name, description) VALUES
    -- Users
    (uuid_generate_v4(), 'users.create',          'Create new users'),
    (uuid_generate_v4(), 'users.read',             'View user details'),
    (uuid_generate_v4(), 'users.update',           'Update user details'),
    (uuid_generate_v4(), 'users.delete',           'Delete users'),
    -- Stores
    (uuid_generate_v4(), 'stores.create',          'Create new stores'),
    (uuid_generate_v4(), 'stores.read',            'View store details'),
    (uuid_generate_v4(), 'stores.update',          'Update store details'),
    (uuid_generate_v4(), 'stores.delete',          'Delete stores'),
    -- Products
    (uuid_generate_v4(), 'products.create',        'Create new products'),
    (uuid_generate_v4(), 'products.read',          'View product details'),
    (uuid_generate_v4(), 'products.update',        'Update product details'),
    (uuid_generate_v4(), 'products.delete',        'Delete products'),
    -- Stock
    (uuid_generate_v4(), 'stock.read',             'View stock levels'),
    (uuid_generate_v4(), 'stock.adjust',           'Manually adjust stock'),
    (uuid_generate_v4(), 'stock.transfer',         'Transfer stock between stores'),
    -- Transactions
    (uuid_generate_v4(), 'transactions.create',    'Create new transactions'),
    (uuid_generate_v4(), 'transactions.read',      'View transactions'),
    (uuid_generate_v4(), 'transactions.void',      'Void/cancel transactions'),
    -- Purchase Orders
    (uuid_generate_v4(), 'purchase_orders.create', 'Create purchase orders'),
    (uuid_generate_v4(), 'purchase_orders.read',   'View purchase orders'),
    (uuid_generate_v4(), 'purchase_orders.update', 'Update purchase orders'),
    (uuid_generate_v4(), 'purchase_orders.receive','Mark purchase orders as received'),
    -- Reports
    (uuid_generate_v4(), 'reports.view',           'View reports and analytics'),
    (uuid_generate_v4(), 'reports.export',         'Export reports');

-- ─── Assign Permissions to Roles ─────────────────────────────────────────────
-- superadmin: all permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'superadmin';

-- admin: all permissions (within their store)
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'admin';

-- manager: products, stock, purchase orders, reports (no user management)
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'manager'
  AND p.name IN (
    'products.create','products.read','products.update',
    'stock.read','stock.adjust','stock.transfer',
    'purchase_orders.create','purchase_orders.read','purchase_orders.update','purchase_orders.receive',
    'transactions.read',
    'reports.view','reports.export'
  );

-- cashier: transactions only + product/stock read
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'cashier'
  AND p.name IN (
    'products.read',
    'stock.read',
    'transactions.create','transactions.read'
  );

-- staff: read-only
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'staff'
  AND p.name IN (
    'products.read',
    'stock.read',
    'transactions.read'
  );

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM role_permissions;
DELETE FROM permissions;
DELETE FROM roles;
-- +goose StatementEnd
