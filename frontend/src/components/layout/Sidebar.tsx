'use client';

import { useMemo } from 'react';
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
  Store,
  Tag,
  UtensilsCrossed,
  Grid3x3,
  Receipt,
  History,
  UserRound,
  UserCog,
  ChefHat,
  Wallet,
  ChevronLeft,
  ChevronRight,
  ArrowDownToLine,
  ShieldCheck,
} from 'lucide-react';
import { useAuth } from '@/lib/auth/AuthContext';
import { useTheme } from '@/lib/theme/ThemeContext';
import { useSidebar } from '@/lib/context/SidebarContext';

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
  | 'users.delete'
  | 'reports.audit';

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
    'reports.audit',
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
    'reports.audit',
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
    'reports.audit',
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
      { href: '/stock', icon: Warehouse, label: 'Stok Terkini', permission: 'stock.read' },
    ],
  },
  {
    label: 'Pembelian',
    permission: 'suppliers.read',
    items: [
      {
        href: '/purchase-orders',
        icon: ClipboardList,
        label: 'Pembelian',
        permission: 'suppliers.read',
      },
      { href: '/suppliers', icon: Users, label: 'Supplier', permission: 'suppliers.read' },
    ],
  },
  {
    label: 'Keuangan & Analitik',
    permission: 'reports.read',
    items: [
      { href: '/incomes', icon: ArrowDownToLine, label: 'Pemasukan', permission: 'reports.read' },
      { href: '/expenses', icon: Wallet, label: 'Pengeluaran', permission: 'reports.read' },
      { href: '/reports', icon: BarChart3, label: 'Laporan', permission: 'reports.read' },
      {
        href: '/activity-logs',
        icon: ShieldCheck,
        label: 'Activity Log',
        permission: 'reports.audit',
      },
    ],
  },
  {
    label: 'Pengaturan',
    items: [
      { href: '/stores', icon: Store, label: 'Toko', adminOnly: true },
      { href: '/users', icon: UserCog, label: 'Pengguna', adminOnly: true },
      {
        href: '/settings/income-categories',
        icon: Tag,
        label: 'Kategori Pemasukan',
        adminOnly: true,
      },
      {
        href: '/settings/expense-categories',
        icon: Tag,
        label: 'Kategori Pengeluaran',
        adminOnly: true,
      },
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
  const { selectedStore, stores, selectStore } = useAuth();
  const { isDark } = useTheme();
  const { isCollapsed, toggleCollapsed } = useSidebar();
  const isRestaurant = selectedStore?.store_type === 'restaurant';
  const role = selectedStore?.role;
  const isSuperOrAdmin = role === 'superadmin' || role === 'admin';

  // Inject restaurant group after "Penjualan" when in restaurant mode
  const allGroups: NavGroup[] = isRestaurant
    ? [baseGroups[0], baseGroups[1], restaurantGroup, ...baseGroups.slice(2)]
    : baseGroups;

  // Build the final visible groups based on the current store role
  const finalGroups = useMemo(() => {
    return allGroups
      .map(group => {
        const isSettings = group.label === 'Pengaturan';

        const visibleItems = group.items
          .map(item => {
            if (item.href === '/products' && isRestaurant) {
              return { ...item, label: 'Bahan Baku' };
            }
            return item;
          })
          .filter(item => {
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
  }, [allGroups, isRestaurant, isSuperOrAdmin, role]);

  return (
    <>
      {/* Mobile overlay */}
      <div
        className={`sidebar-overlay-backdrop ${!isCollapsed ? 'open' : ''}`}
        onClick={toggleCollapsed}
      />

      <aside className={`sidebar ${isCollapsed ? 'sidebar-collapsed' : 'sidebar-expanded open'}`}>
        {/* ── Logo / Header ── */}
        <div className="sidebar-logo">
          {isCollapsed ? (
            <button
              onClick={toggleCollapsed}
              className="btn btn-ghost btn-sm"
              title="Perluas sidebar"
              aria-label="Expand sidebar"
              style={{ padding: '6px 8px', margin: 'auto' }}
            >
              <ChevronRight size={16} />
            </button>
          ) : (
            <div className="flex items-center justify-between w-full">
              <div className="flex items-center">
                <Image
                  src={isDark ? '/logo-dashboard-dark.svg' : '/logo-dashboard-light.svg'}
                  alt="Moedah"
                  width={90}
                  height={24}
                  style={{ objectFit: 'contain', flexShrink: 0 }}
                  priority
                />
              </div>
              <div className="flex items-center">
                <button
                  onClick={toggleCollapsed}
                  className="btn btn-ghost btn-sm"
                  title="Ciutkan sidebar"
                  aria-label="Collapse sidebar"
                  style={{ padding: '6px' }}
                >
                  <ChevronLeft size={14} />
                </button>
              </div>
            </div>
          )}
        </div>

        {/* ── Store Selector ── */}
        {stores.length > 0 && !isCollapsed && (
          <div
            style={{
              padding: '10px 12px 10px',
              borderBottom: '1px solid var(--border)',
            }}
          >
            <p className="type-overline" style={{ marginBottom: 6, paddingLeft: 2 }}>
              Toko Aktif
            </p>
            <select
              value={selectedStore?.store_id ?? ''}
              onChange={e => selectStore(e.target.value)}
              className="input"
              style={{ fontSize: '0.8rem', padding: '6px 10px' }}
            >
              {!selectedStore && <option value="">— Pilih Toko —</option>}
              {stores.map(s => (
                <option key={s.store_id} value={s.store_id}>
                  {s.store_name}
                </option>
              ))}
            </select>
            {selectedStore && (
              <div className="flex items-center gap-2 mt-1.5">
                <span className="badge badge-blue" style={{ textTransform: 'capitalize' }}>
                  {selectedStore.role}
                </span>
                <span
                  className="badge"
                  style={{
                    background: isRestaurant ? 'rgba(251,146,60,0.10)' : 'rgba(16,185,129,0.10)',
                    color: isRestaurant ? '#fb923c' : '#10b981',
                  }}
                >
                  {isRestaurant ? '🍽️ Restoran' : '🏪 Retail'}
                </span>
              </div>
            )}
          </div>
        )}

        {/* ── Navigation ── */}
        <nav className="sidebar-nav" style={{ flex: 1, overflowY: 'auto' }}>
          {finalGroups.map(group => (
            <div key={group.label}>
              {/* Group section label */}
              {!isCollapsed && <div className="nav-section">{group.label}</div>}

              {/* Group nav items */}
              {group.items.map(({ href, icon: Icon, label }) => {
                const active =
                  pathname === href || (href !== '/dashboard' && pathname.startsWith(href + '/'));
                return (
                  <Link
                    key={href}
                    href={href}
                    className={`nav-item ${active ? 'active' : ''}`}
                    title={isCollapsed ? label : undefined}
                    style={{
                      justifyContent: isCollapsed ? 'center' : 'flex-start',
                      padding: isCollapsed ? '12px 0' : undefined,
                    }}
                  >
                    <Icon size={14} strokeWidth={active ? 2.5 : 2} />
                    {!isCollapsed && (
                      <span
                        style={{
                          opacity: isCollapsed ? 0 : 1,
                          transition: 'opacity 0.2s',
                          whiteSpace: 'nowrap',
                        }}
                      >
                        {label}
                      </span>
                    )}
                  </Link>
                );
              })}
            </div>
          ))}
        </nav>
      </aside>
    </>
  );
}
