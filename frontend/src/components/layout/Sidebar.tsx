'use client';

import Link from 'next/link';
import Image from 'next/image';
import { usePathname } from 'next/navigation';
import {
  LayoutDashboard,
  ShoppingCart,
  Package,
  Warehouse,
  ClipboardList,
  BarChart3,
  Users,
  LogOut,
  Sun,
  Moon,
  Store,
  Tag,
  UtensilsCrossed,
  Grid3x3,
  Receipt,
  History,
  UserRound,
  UserCog,
  ChefHat,
} from 'lucide-react';
import { useAuth } from '@/lib/auth/AuthContext';
import { useTheme } from '@/lib/theme/ThemeContext';

// ── Types ─────────────────────────────────────────────────────────────────────
interface NavItem {
  href: string;
  icon: React.ElementType;
  label: string;
  /** Backend permission key required to see this item. Undefined = visible to all. */
  permission?: string;
  /** If true, only superadmin sees this item (global admin feature). */
  adminOnly?: boolean;
}
interface NavGroup {
  label: string;
  items: NavItem[];
  /** If set, the whole group is hidden when the user lacks this permission. */
  permission?: string;
}

// ── Role → permission map (mirrors backend role_permissions seed) ─────────────
//
// Rather than fetching permissions from the API on every render, we mirror the
// backend RBAC model here. This is safe because:
//   - The role name lives in the JWT / /me response (`selectedStore.role`).
//   - Menu visibility is *cosmetic*; the backend still enforces per-request checks.
//   - If the seed changes, update this map too.
//
type Permission =
  | 'products.read'
  | 'products.create'
  | 'products.update'
  | 'products.delete'
  | 'stock.read'
  | 'stock.update'
  | 'reports.read'
  | 'suppliers.read'
  | 'suppliers.create'
  | 'sales.create'
  | 'sales.read'
  | 'users.read'
  | 'users.create'
  | 'users.update'
  | 'users.delete';

type RoleName = 'superadmin' | 'admin' | 'manager' | 'cashier' | 'staff';

const ROLE_PERMISSIONS: Record<RoleName, Set<Permission>> = {
  superadmin: new Set<Permission>([
    'products.read',
    'products.create',
    'products.update',
    'products.delete',
    'stock.read',
    'stock.update',
    'reports.read',
    'suppliers.read',
    'suppliers.create',
    'sales.create',
    'sales.read',
    'users.read',
    'users.create',
    'users.update',
    'users.delete',
  ]),
  admin: new Set<Permission>([
    'products.read',
    'products.create',
    'products.update',
    'products.delete',
    'stock.read',
    'stock.update',
    'reports.read',
    'suppliers.read',
    'suppliers.create',
    'sales.create',
    'sales.read',
    'users.read',
    'users.create',
    'users.update',
    'users.delete',
  ]),
  manager: new Set<Permission>([
    'products.read',
    'products.create',
    'products.update',
    'stock.read',
    'stock.update',
    'reports.read',
    'suppliers.read',
    'sales.create',
    'sales.read',
  ]),
  cashier: new Set<Permission>(['products.read', 'stock.read', 'sales.create', 'sales.read']),
  staff: new Set<Permission>(['products.read', 'stock.read', 'sales.read']),
};

function hasPermission(role: string | undefined, perm: string): boolean {
  if (!role) return false;
  if (role === 'superadmin') return true;
  const perms = ROLE_PERMISSIONS[role as RoleName];
  return perms ? perms.has(perm as Permission) : false;
}

// ── Nav group definitions ─────────────────────────────────────────────────────
const baseGroups: NavGroup[] = [
  {
    label: 'Umum',
    items: [{ href: '/dashboard', icon: LayoutDashboard, label: 'Dashboard' }],
  },
  {
    label: 'Penjualan',
    permission: 'sales.create',
    items: [
      { href: '/pos', icon: ShoppingCart, label: 'Kasir / POS', permission: 'sales.create' },
      {
        href: '/transactions',
        icon: Receipt,
        label: 'Riwayat Transaksi',
        permission: 'sales.read',
      },
      { href: '/customers', icon: UserRound, label: 'Customer', permission: 'sales.read' },
    ],
  },
  {
    label: 'Inventori',
    permission: 'products.read',
    items: [
      { href: '/products', icon: Package, label: 'Produk', permission: 'products.read' },
      { href: '/categories', icon: Tag, label: 'Kategori', permission: 'products.read' },
      {
        href: '/price-history',
        icon: History,
        label: 'Riwayat Harga',
        permission: 'products.read',
      },
      { href: '/stock', icon: Warehouse, label: 'Stok', permission: 'stock.read' },
    ],
  },
  {
    label: 'Pembelian',
    permission: 'suppliers.read',
    items: [
      {
        href: '/purchase-orders',
        icon: ClipboardList,
        label: 'Purchase Order',
        permission: 'suppliers.read',
      },
      { href: '/suppliers', icon: Users, label: 'Supplier', permission: 'suppliers.read' },
    ],
  },
  {
    label: 'Analitik',
    permission: 'reports.read',
    items: [{ href: '/reports', icon: BarChart3, label: 'Laporan', permission: 'reports.read' }],
  },
  {
    label: 'Pengaturan',
    items: [
      { href: '/stores', icon: Store, label: 'Toko', adminOnly: true },
      { href: '/users', icon: UserCog, label: 'Pengguna', adminOnly: true },
    ],
  },
];

const restaurantGroup: NavGroup = {
  label: 'Restoran',
  permission: 'products.read',
  items: [
    { href: '/kds', icon: ChefHat, label: 'Kitchen Display', permission: 'products.read' },
    { href: '/tables', icon: Grid3x3, label: 'Meja', permission: 'products.read' },
    { href: '/menu-items', icon: UtensilsCrossed, label: 'Menu', permission: 'products.read' },
  ],
};

// ── Sidebar ───────────────────────────────────────────────────────────────────
export default function Sidebar() {
  const pathname = usePathname();
  const { user, selectedStore, stores, selectStore, logout } = useAuth();
  const { isDark, toggleTheme } = useTheme();
  const isRestaurant = selectedStore?.store_type === 'restaurant';
  const role = selectedStore?.role;
  const isSuperOrAdmin = role === 'superadmin' || role === 'admin';

  // Inject restaurant group after "Penjualan" when in restaurant mode
  const allGroups: NavGroup[] = isRestaurant
    ? [baseGroups[0], baseGroups[1], restaurantGroup, ...baseGroups.slice(2)]
    : baseGroups;

  // Build the final visible groups based on the current store role
  const finalGroups = allGroups
    .map(group => {
      const isSettings = group.label === 'Pengaturan';

      const visibleItems = group.items.filter(item => {
        if (item.adminOnly) return isSuperOrAdmin;
        if (item.permission) return hasPermission(role, item.permission);
        return true;
      });

      if (visibleItems.length === 0) return null;

      // For non-settings groups, check group-level permission gate
      if (!isSettings && group.permission && !hasPermission(role, group.permission)) return null;

      return { ...group, items: visibleItems };
    })
    .filter(Boolean) as NavGroup[];

  return (
    <aside className="sidebar">
      {/* Logo */}
      <div className="sidebar-logo">
        <Image
          src={isDark ? '/logo-icon-dark.svg' : '/logo-icon-light.svg'}
          alt="Moedah"
          width={32}
          height={36}
          style={{ objectFit: 'contain', flexShrink: 0 }}
          priority
        />
        <div
          style={{
            fontWeight: 800,
            fontSize: '1rem',
            color: 'var(--brand)',
            letterSpacing: '-0.3px',
          }}
        >
          Moedah
        </div>
      </div>

      {/* Store Selector */}
      {stores.length > 0 && (
        <div style={{ padding: '10px 12px 8px', borderBottom: '1px solid var(--border)' }}>
          <div
            style={{
              fontSize: '0.68rem',
              color: 'var(--text-3)',
              marginBottom: 5,
              textTransform: 'uppercase',
              letterSpacing: '0.06em',
            }}
          >
            Toko Aktif
          </div>
          <select
            value={selectedStore?.store_id ?? ''}
            onChange={e => selectStore(e.target.value)}
            style={{
              width: '100%',
              background: 'var(--bg-elevated)',
              border: '1px solid var(--border-md)',
              borderRadius: 7,
              padding: '6px 10px',
              color: 'var(--text-1)',
              fontSize: '0.8rem',
              cursor: 'pointer',
              outline: 'none',
            }}
          >
            {!selectedStore && <option value="">— Pilih Toko —</option>}
            {stores.map(s => (
              <option key={s.store_id} value={s.store_id}>
                {s.store_name}
              </option>
            ))}
          </select>
          {selectedStore && (
            <div style={{ display: 'flex', gap: 6, marginTop: 5, alignItems: 'center' }}>
              <div style={{ fontSize: '0.7rem', color: 'var(--accent-em)' }}>
                {selectedStore.role}
              </div>
              <div
                style={{
                  fontSize: '0.62rem',
                  padding: '1px 5px',
                  borderRadius: 4,
                  fontWeight: 600,
                  background: isRestaurant ? 'rgba(251,146,60,0.15)' : 'rgba(16,185,129,0.12)',
                  color: isRestaurant ? '#fb923c' : 'var(--accent-em)',
                }}
              >
                {isRestaurant ? '🍽️ Restaurant' : '🏪 Retail'}
              </div>
            </div>
          )}
        </div>
      )}

      {/* Grouped Navigation */}
      <nav className="sidebar-nav" style={{ flex: 1, overflowY: 'auto' }}>
        {finalGroups.map(group => (
          <div key={group.label} style={{ marginBottom: 4 }}>
            {/* Group label */}
            <div
              style={{
                fontSize: '0.62rem',
                fontWeight: 700,
                letterSpacing: '0.09em',
                textTransform: 'uppercase',
                color: 'var(--text-3)',
                padding: '10px 14px 4px',
              }}
            >
              {group.label}
            </div>

            {/* Group items */}
            {group.items.map(({ href, icon: Icon, label }) => {
              const active =
                pathname === href || (href !== '/dashboard' && pathname.startsWith(href));
              return (
                <Link key={href} href={href} className={`nav-item ${active ? 'active' : ''}`}>
                  <Icon size={15} />
                  {label}
                </Link>
              );
            })}
          </div>
        ))}
      </nav>

      {/* User Profile + Logout */}
      <div style={{ padding: '12px', borderTop: '1px solid var(--border)' }}>
        {user && (
          <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 8 }}>
            <div
              style={{
                width: 32,
                height: 32,
                borderRadius: '50%',
                flexShrink: 0,
                background: 'linear-gradient(135deg, var(--accent-in), var(--accent-em))',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                fontSize: '0.8rem',
                fontWeight: 700,
                color: '#fff',
              }}
            >
              {user.name.charAt(0).toUpperCase()}
            </div>
            <div style={{ overflow: 'hidden' }}>
              <div
                style={{
                  fontSize: '0.8rem',
                  fontWeight: 600,
                  color: 'var(--text-1)',
                  whiteSpace: 'nowrap',
                  overflow: 'hidden',
                  textOverflow: 'ellipsis',
                }}
              >
                {user.name}
              </div>
              <div
                style={{
                  fontSize: '0.7rem',
                  color: 'var(--text-3)',
                  whiteSpace: 'nowrap',
                  overflow: 'hidden',
                  textOverflow: 'ellipsis',
                }}
              >
                {user.email}
              </div>
            </div>
          </div>
        )}
        {/* Theme toggle + Logout row */}
        <div style={{ display: 'flex', gap: 6 }}>
          <button
            onClick={toggleTheme}
            className="btn btn-ghost btn-sm"
            title={isDark ? 'Ganti ke mode terang' : 'Ganti ke mode gelap'}
            style={{ padding: '6px 8px', borderRadius: 8, border: '1px solid var(--border-md)' }}
          >
            {isDark ? <Sun size={14} /> : <Moon size={14} />}
          </button>
          <button
            onClick={logout}
            className="btn btn-ghost btn-sm"
            style={{ flex: 1, justifyContent: 'flex-start', gap: 8 }}
          >
            <LogOut size={14} /> Keluar
          </button>
        </div>
      </div>
    </aside>
  );
}
