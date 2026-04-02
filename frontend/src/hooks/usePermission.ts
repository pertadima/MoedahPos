/**
 * usePermission — lightweight RBAC hook for the frontend.
 *
 * Mirrors the backend role-permission assignments.
 * Menu visibility is cosmetic; the backend still enforces per-request checks.
 */
'use client';

import { useAuth } from '@/lib/auth/AuthContext';

type Permission =
  | 'products.read'  | 'products.create' | 'products.update' | 'products.delete'
  | 'stock.read'     | 'stock.update'
  | 'reports.read'
  | 'suppliers.read' | 'suppliers.create'| 'suppliers.update'| 'suppliers.delete'
  | 'sales.create'   | 'sales.read'
  | 'users.read'     | 'users.create'    | 'users.update'    | 'users.delete'
  | 'stores.read'    | 'stores.create'   | 'stores.update'   | 'stores.delete';

type RoleName = 'superadmin' | 'admin' | 'manager' | 'cashier' | 'staff';

const ROLE_PERMISSIONS: Record<RoleName, Permission[]> = {
  superadmin: [
    'products.read', 'products.create', 'products.update', 'products.delete',
    'stock.read', 'stock.update',
    'reports.read',
    'suppliers.read', 'suppliers.create', 'suppliers.update', 'suppliers.delete',
    'sales.create', 'sales.read',
    'users.read', 'users.create', 'users.update', 'users.delete',
    'stores.read', 'stores.create', 'stores.update', 'stores.delete',
  ],
  admin: [
    'products.read', 'products.create', 'products.update', 'products.delete',
    'stock.read', 'stock.update',
    'reports.read',
    'suppliers.read', 'suppliers.create', 'suppliers.update', 'suppliers.delete',
    'sales.create', 'sales.read',
    'users.read', 'users.create', 'users.update', 'users.delete',
    'stores.read', 'stores.create', 'stores.update', 'stores.delete',
  ],
  manager: [
    'products.read', 'products.create', 'products.update',
    'stock.read', 'stock.update',
    'reports.read',
    'suppliers.read', 'suppliers.create',
    'sales.create', 'sales.read',
  ],
  cashier: [
    'products.read', 'products.update',
    'stock.read',
    'sales.create', 'sales.read',
  ],
  staff: [
    'products.read',
    'stock.read',
    'sales.read',
  ],
};

const PERMISSION_SET: Record<RoleName, Set<Permission>> = Object.fromEntries(
  Object.entries(ROLE_PERMISSIONS).map(([role, perms]) => [role, new Set(perms)])
) as Record<RoleName, Set<Permission>>;

export function usePermission() {
  const { selectedStore } = useAuth();
  const role = (selectedStore?.role ?? '') as RoleName;

  const can = (perm: Permission): boolean => {
    if (!role) return false;
    if (role === 'superadmin') return true;
    return PERMISSION_SET[role]?.has(perm) ?? false;
  };

  const isAdminOrAbove = role === 'superadmin' || role === 'admin';
  const isManagerOrAbove = isAdminOrAbove || role === 'manager';

  return { can, role, isAdminOrAbove, isManagerOrAbove };
}
