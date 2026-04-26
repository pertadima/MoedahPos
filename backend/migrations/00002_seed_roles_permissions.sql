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
    -- Penjualan (Transactions/Sales History)
    (uuid_generate_v4(), 'penjualan:read',   'View sales history'),
    (uuid_generate_v4(), 'penjualan:write',  'Create sales records'),
    (uuid_generate_v4(), 'penjualan:update', 'Update sales records'),
    (uuid_generate_v4(), 'penjualan:delete', 'Delete/Void sales'),
    -- Kasir (Active POS operations)
    (uuid_generate_v4(), 'kasir:read',       'Access POS interface'),
    (uuid_generate_v4(), 'kasir:write',      'Process new transactions'),
    (uuid_generate_v4(), 'kasir:update',     'Modify active cart'),
    (uuid_generate_v4(), 'kasir:delete',     'Clear active cart'),
    -- Inventory (Products, Categories, Stock)
    (uuid_generate_v4(), 'inventory:read',   'View inventory and stock'),
    (uuid_generate_v4(), 'inventory:write',  'Add new products/stock'),
    (uuid_generate_v4(), 'inventory:update', 'Edit product details'),
    (uuid_generate_v4(), 'inventory:delete', 'Delete products'),
    -- Pembelian (Purchase Orders, Suppliers)
    (uuid_generate_v4(), 'pembelian:read',   'View purchase orders'),
    (uuid_generate_v4(), 'pembelian:write',  'Create purchase orders'),
    (uuid_generate_v4(), 'pembelian:update', 'Edit purchase orders'),
    (uuid_generate_v4(), 'pembelian:delete', 'Delete purchase orders'),
    -- Keuangan (Expenses, Incomes)
    (uuid_generate_v4(), 'keuangan:read',    'View financial records'),
    (uuid_generate_v4(), 'keuangan:write',   'Add expenses/incomes'),
    (uuid_generate_v4(), 'keuangan:update',  'Edit financial records'),
    (uuid_generate_v4(), 'keuangan:delete',  'Delete financial records'),
    -- Settings (Users, Stores, Roles, System)
    (uuid_generate_v4(), 'settings:read',    'View settings and users'),
    (uuid_generate_v4(), 'settings:write',   'Add users/roles'),
    (uuid_generate_v4(), 'settings:update',  'Edit settings/users'),
    (uuid_generate_v4(), 'settings:delete',  'Delete users/roles');

-- ─── Assign Permissions to Roles ─────────────────────────────────────────────
-- superadmin: all permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'superadmin';

-- admin: all permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'admin';

-- manager: all except settings write/update/delete
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'manager'
  AND p.name NOT IN ('settings:write', 'settings:update', 'settings:delete');

-- cashier: kasir (all), penjualan (read), inventory (read)
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'cashier'
  AND p.name IN (
    'kasir:read', 'kasir:write', 'kasir:update', 'kasir:delete',
    'penjualan:read',
    'inventory:read'
  );

-- staff: read-only for inventory, penjualan
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'staff'
  AND p.name IN (
    'inventory:read',
    'penjualan:read'
  );

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM role_permissions;
DELETE FROM permissions;
DELETE FROM roles;
-- +goose StatementEnd
