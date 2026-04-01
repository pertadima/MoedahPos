'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import {
  LayoutDashboard, ShoppingCart, Package, Warehouse,
  ClipboardList, BarChart3, Users, LogOut,
  Store, CreditCard, Tag, UtensilsCrossed, Grid3x3, Receipt, History, UserRound, UserCog,
} from 'lucide-react';
import { useAuth } from '@/lib/auth/AuthContext';

// ── Types ─────────────────────────────────────────────────────────────────────
interface NavItem {
  href: string;
  icon: React.ElementType;
  label: string;
}
interface NavGroup {
  label: string;
  items: NavItem[];
}

// ── Nav group definitions ─────────────────────────────────────────────────────
const baseGroups: NavGroup[] = [
  {
    label: 'Umum',
    items: [
      { href: '/dashboard', icon: LayoutDashboard, label: 'Dashboard' },
    ],
  },
  {
    label: 'Penjualan',
    items: [
      { href: '/pos',          icon: ShoppingCart, label: 'Kasir / POS' },
      { href: '/transactions', icon: Receipt,      label: 'Riwayat Transaksi' },
      { href: '/customers',   icon: UserRound,    label: 'Customer' },
    ],
  },
  {
    label: 'Inventori',
    items: [
      { href: '/products',      icon: Package,   label: 'Produk' },
      { href: '/categories',    icon: Tag,        label: 'Kategori' },
      { href: '/price-history', icon: History,    label: 'Riwayat Harga' },
      { href: '/stock',         icon: Warehouse,  label: 'Stok' },
    ],
  },
  {
    label: 'Pembelian',
    items: [
      { href: '/purchase-orders', icon: ClipboardList, label: 'Purchase Order' },
      { href: '/suppliers',       icon: Users,         label: 'Supplier' },
    ],
  },
  {
    label: 'Analitik',
    items: [
      { href: '/reports', icon: BarChart3, label: 'Laporan' },
    ],
  },
  {
    label: 'Pengaturan',
    items: [
      { href: '/stores', icon: Store,   label: 'Toko' },
      { href: '/users',  icon: UserCog, label: 'Pengguna' },
    ],
  },
];

const restaurantGroup: NavGroup = {
  label: 'Restoran',
  items: [
    { href: '/tables',     icon: Grid3x3,         label: 'Meja' },
    { href: '/menu-items', icon: UtensilsCrossed, label: 'Menu' },
  ],
};

// ── Sidebar ───────────────────────────────────────────────────────────────────
export default function Sidebar() {
  const pathname = usePathname();
  const { user, selectedStore, stores, selectStore, logout } = useAuth();
  const isRestaurant = selectedStore?.store_type === 'restaurant';

  // Inject restaurant group after "Penjualan" when in restaurant mode
  const groups: NavGroup[] = isRestaurant
    ? [
        baseGroups[0],                        // Umum
        baseGroups[1],                        // Penjualan
        restaurantGroup,                      // Restoran (tables + menu)
        ...baseGroups.slice(2),               // Inventori … Pengaturan
      ]
    : baseGroups;

  return (
    <aside className="sidebar">
      {/* Logo */}
      <div className="sidebar-logo">
        <div className="sidebar-logo-icon">
          <CreditCard size={18} color="#fff" />
        </div>
        <div>
          <div style={{ fontWeight: 700, fontSize: '0.95rem', color: 'var(--text-1)' }}>MoedahPOS</div>
          <div style={{ fontSize: '0.7rem', color: 'var(--text-3)' }}>v3.0</div>
        </div>
      </div>

      {/* Store Selector */}
      {stores.length > 0 && (
        <div style={{ padding: '10px 12px 8px', borderBottom: '1px solid var(--border)' }}>
          <div style={{ fontSize: '0.68rem', color: 'var(--text-3)', marginBottom: 5, textTransform: 'uppercase', letterSpacing: '0.06em' }}>
            Toko Aktif
          </div>
          <select
            value={selectedStore?.store_id ?? ''}
            onChange={e => selectStore(e.target.value)}
            style={{
              width: '100%', background: 'var(--bg-elevated)', border: '1px solid var(--border-md)',
              borderRadius: 7, padding: '6px 10px', color: 'var(--text-1)', fontSize: '0.8rem',
              cursor: 'pointer', outline: 'none',
            }}
          >
            {!selectedStore && <option value="">— Pilih Toko —</option>}
            {stores.map(s => (
              <option key={s.store_id} value={s.store_id}>{s.store_name}</option>
            ))}
          </select>
          {selectedStore && (
            <div style={{ display: 'flex', gap: 6, marginTop: 5, alignItems: 'center' }}>
              <div style={{ fontSize: '0.7rem', color: 'var(--accent-em)' }}>{selectedStore.role}</div>
              <div style={{
                fontSize: '0.62rem', padding: '1px 5px', borderRadius: 4, fontWeight: 600,
                background: isRestaurant ? 'rgba(251,146,60,0.15)' : 'rgba(16,185,129,0.12)',
                color: isRestaurant ? '#fb923c' : 'var(--accent-em)',
              }}>
                {isRestaurant ? '🍽️ Restaurant' : '🏪 Retail'}
              </div>
            </div>
          )}
        </div>
      )}

      {/* Grouped Navigation */}
      <nav className="sidebar-nav" style={{ flex: 1, overflowY: 'auto' }}>
        {groups.map(group => (
          <div key={group.label} style={{ marginBottom: 4 }}>
            {/* Group label */}
            <div style={{
              fontSize: '0.62rem', fontWeight: 700, letterSpacing: '0.09em',
              textTransform: 'uppercase', color: 'var(--text-3)',
              padding: '10px 14px 4px',
            }}>
              {group.label}
            </div>

            {/* Group items */}
            {group.items.map(({ href, icon: Icon, label }) => {
              const active = pathname === href || (href !== '/dashboard' && pathname.startsWith(href));
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
            <div style={{
              width: 32, height: 32, borderRadius: '50%', flexShrink: 0,
              background: 'linear-gradient(135deg, var(--accent-in), var(--accent-em))',
              display: 'flex', alignItems: 'center', justifyContent: 'center',
              fontSize: '0.8rem', fontWeight: 700, color: '#fff',
            }}>
              {user.name.charAt(0).toUpperCase()}
            </div>
            <div style={{ overflow: 'hidden' }}>
              <div style={{ fontSize: '0.8rem', fontWeight: 600, color: 'var(--text-1)', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                {user.name}
              </div>
              <div style={{ fontSize: '0.7rem', color: 'var(--text-3)', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                {user.email}
              </div>
            </div>
          </div>
        )}
        <button onClick={logout} className="btn btn-ghost btn-sm"
          style={{ width: '100%', justifyContent: 'flex-start', gap: 8 }}>
          <LogOut size={14} /> Keluar
        </button>
      </div>
    </aside>
  );
}
