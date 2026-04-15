'use client';

import { useEffect, useState, useCallback } from 'react';
import {
  CalendarDays,
  User as UserIcon,
  Search,
  ChevronLeft,
  ChevronRight,
  Filter,
  Info,
  ShieldCheck,
  Package,
  CreditCard,
  LogIn,
  AlertCircle,
  Clock,
  ChevronDown,
  ChevronUp,
  Loader2,
} from 'lucide-react';
import { useAuth } from '@/lib/auth/AuthContext';
import { activityLogsApi, type ActivityLog } from '@/lib/api/activity-logs';
import { storesApi } from '@/lib/api/store-apis';
import type { PaginatedData, User } from '@/types';
import { ApiError } from '@/lib/api/client';

// ── Helpers ──────────────────────────────────────────────────────────────────
const today = () => new Date().toISOString().slice(0, 10);
const daysAgo = (n: number) => {
  const d = new Date();
  d.setDate(d.getDate() - n);
  return d.toISOString().slice(0, 10);
};

const MODULE_ICONS: Record<string, React.ReactNode> = {
  AUTH: <ShieldCheck size={16} className="text-blue-500" />,
  TRANSACTION: <CreditCard size={16} className="text-emerald-500" />,
  DISCOUNT: <AlertCircle size={16} className="text-amber-500" />,
  INVENTORY: <Package size={16} className="text-indigo-500" />,
  PURCHASE: <Package size={16} className="text-purple-500" />,
  INCOME: <LogIn size={16} className="text-emerald-500" />,
  EXPENSE: <CreditCard size={16} className="text-rose-500" />,
};

const ACTION_LABELS: Record<string, { label: string; color: string }> = {
  AUTH_LOGIN: { label: 'Login', color: 'var(--accent-em)' },
  AUTH_LOGOUT: { label: 'Logout', color: 'var(--text-3)' },
  TRANSACTION_CREATE: { label: 'Transaksi Baru', color: 'var(--accent-em)' },
  TRANSACTION_CANCEL: { label: 'Transaksi Dibatalkan', color: 'var(--accent-rd)' },
  DISCOUNT_ITEM: { label: 'Diskon Item', color: 'var(--accent-am)' },
  DISCOUNT_CART: { label: 'Diskon Keranjang', color: 'var(--accent-am)' },
  PRICE_OVERRIDE: { label: 'Override Harga', color: 'var(--accent-rd)' },
  STOCK_ADJUSTMENT: { label: 'Penyesuaian Stok', color: 'var(--accent-in)' },
  PURCHASE_ORDER_CREATE: { label: 'PO Baru', color: '#8b5cf6' },
  PURCHASE_ORDER_UPDATE: { label: 'PO Diperbarui', color: '#a78bfa' },
  PURCHASE_ORDER_PAYMENT: { label: 'Pembayaran PO', color: '#10b981' },
  INCOME_CREATE: { label: 'Pemasukan Baru', color: '#10b981' },
  INCOME_UPDATE: { label: 'Update Pemasukan', color: '#34d399' },
  INCOME_DELETE: { label: 'Hapus Pemasukan', color: '#ef4444' },
  EXPENSE_CREATE: { label: 'Pengeluaran Baru', color: '#ef4444' },
  EXPENSE_UPDATE: { label: 'Update Pengeluaran', color: '#f87171' },
  EXPENSE_DELETE: { label: 'Hapus Pengeluaran', color: '#ef4444' },
};

function formatRp(val: number) {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0,
  }).format(val);
}

function getReadableDescription(log: ActivityLog): string | null {
  const meta: any = log.metadata || {};
  switch (log.action_type) {
    case 'PURCHASE_ORDER_CREATE':
      return `Membuat Purchase Order ${meta.po_number || ''} supplier ${meta.supplier_name || '-'} (${meta.total_amount ? formatRp(meta.total_amount) : ''})`;
    case 'PURCHASE_ORDER_PAYMENT':
      return `Mencatat pembayaran ${meta.payment_amount ? formatRp(meta.payment_amount) : ''} untuk PO ${meta.po_number || ''}`;
    case 'INCOME_CREATE':
      return `Menambah pemasukan ${meta.amount ? formatRp(meta.amount) : ''} (${meta.category || ''})`;
    case 'EXPENSE_CREATE':
      return `Mencatat pengeluaran ${meta.amount ? formatRp(meta.amount) : ''} (${meta.category || ''})`;
    case 'INCOME_DELETE':
      return `Menghapus pemasukan ${meta.amount ? formatRp(meta.amount) : ''} (${meta.category || ''})`;
    case 'EXPENSE_DELETE':
      return `Menghapus pengeluaran ${meta.amount ? formatRp(meta.amount) : ''} (${meta.category || ''})`;
  }
  return null;
}

// ── Components ───────────────────────────────────────────────────────────────

function MetadataViewer({ data }: { data: any }) {
  if (!data)
    return <span style={{ color: 'var(--text-3)', fontStyle: 'italic' }}>No metadata</span>;
  return (
    <pre
      style={{
        padding: 16,
        background: 'var(--bg-elevated)',
        color: 'var(--accent-in)',
        borderRadius: 8,
        fontSize: '0.75rem',
        overflowX: 'auto',
        fontFamily: 'monospace',
        border: '1px solid var(--border)',
      }}
    >
      {JSON.stringify(data, null, 2)}
    </pre>
  );
}

function LogRow({ log, index = 0 }: { log: ActivityLog; index?: number }) {
  const [expanded, setExpanded] = useState(false);
  const action = ACTION_LABELS[log.action_type] || { label: log.action_type, color: 'gray' };

  return (
    <>
      <tr
        onClick={() => setExpanded(!expanded)}
        className="reveal-animate"
        style={{ cursor: 'pointer', animationDelay: `${0.2 + index * 0.02}s` }}
      >
        <td>
          <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
            <div
              style={{
                background: 'var(--bg-card)',
                padding: 8,
                borderRadius: 8,
                border: '1px solid var(--border)',
                boxShadow: '0 1px 2px rgba(0,0,0,0.05)',
              }}
            >
              <Clock size={16} style={{ color: 'var(--text-3)' }} />
            </div>
            <div>
              <div style={{ fontWeight: 600, fontSize: '0.875rem' }}>
                {new Date(log.created_at).toLocaleTimeString('id-ID', {
                  hour: '2-digit',
                  minute: '2-digit',
                })}
              </div>
              <div
                style={{
                  fontSize: '0.625rem',
                  color: 'var(--text-3)',
                  fontWeight: 500,
                  textTransform: 'uppercase',
                  letterSpacing: '0.05em',
                }}
              >
                {new Date(log.created_at).toLocaleDateString('id-ID', {
                  day: 'numeric',
                  month: 'short',
                })}
              </div>
            </div>
          </div>
        </td>
        <td>
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <div
              style={{
                width: 32,
                height: 32,
                borderRadius: '50%',
                background: 'rgba(8,132,246,0.1)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                color: 'var(--brand)',
                fontWeight: 700,
                fontSize: '0.75rem',
                textTransform: 'uppercase',
              }}
            >
              {log.user_name.charAt(0)}
            </div>
            <span style={{ fontSize: '0.875rem', fontWeight: 500 }}>{log.user_name}</span>
          </div>
        </td>
        <td>
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <div style={{ padding: 6, borderRadius: 6, background: 'var(--bg-elevated)' }}>
              {MODULE_ICONS[log.module] || <Info size={14} />}
            </div>
            <span
              style={{
                fontSize: '0.75rem',
                fontWeight: 600,
                textTransform: 'uppercase',
                letterSpacing: '0.05em',
                color: 'var(--text-2)',
              }}
            >
              {log.module}
            </span>
          </div>
        </td>
        <td>
          <span
            style={{
              padding: '4px 10px',
              borderRadius: 100,
              fontSize: '0.6875rem',
              fontWeight: 700,
              textTransform: 'uppercase',
              letterSpacing: '-0.025em',
              backgroundColor: `${action.color}15`,
              color: action.color,
              border: `1px solid ${action.color}25`,
            }}
          >
            {action.label}
          </span>
          {getReadableDescription(log) && (
            <div style={{ fontSize: '0.75rem', color: 'var(--text-3)', marginTop: 8 }}>
              {getReadableDescription(log)}
            </div>
          )}
        </td>
        <td style={{ textAlign: 'right' }}>
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'flex-end' }}>
            {expanded ? (
              <ChevronUp size={18} style={{ color: 'var(--text-3)' }} />
            ) : (
              <ChevronDown size={18} style={{ color: 'var(--text-3)' }} />
            )}
          </div>
        </td>
      </tr>
      {expanded && (
        <tr>
          <td colSpan={5} style={{ padding: 0, borderBottom: '1px solid var(--border)' }}>
            <div style={{ padding: 24, background: 'var(--bg-elevated)' }}>
              <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
                <div
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: 8,
                    fontSize: '0.75rem',
                    fontWeight: 700,
                    color: 'var(--text-2)',
                    textTransform: 'uppercase',
                    letterSpacing: '0.1em',
                  }}
                >
                  <Info size={12} /> Metadata Details
                </div>
                <MetadataViewer data={log.metadata} />
                {log.reference_id && (
                  <div
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: 8,
                      fontSize: '0.6875rem',
                      color: 'var(--text-2)',
                      fontFamily: 'monospace',
                      background: 'var(--bg-card)',
                      padding: '6px 12px',
                      borderRadius: 6,
                      border: '1px solid var(--border)',
                      width: 'fit-content',
                    }}
                  >
                    REFERENCE ID:{' '}
                    <span style={{ fontWeight: 700, color: 'var(--text-1)' }}>
                      {log.reference_id}
                    </span>
                  </div>
                )}
              </div>
            </div>
          </td>
        </tr>
      )}
    </>
  );
}

// ── Page ─────────────────────────────────────────────────────────────────────

export default function ActivityLogPage() {
  const { selectedStore } = useAuth();
  const storeId = selectedStore?.store_id;

  // Filters
  const [dateFrom, setDateFrom] = useState(() => daysAgo(7));
  const [dateTo, setDateTo] = useState(() => today());
  const [userFilter, setUserFilter] = useState('');

  // Data
  const [logs, setLogs] = useState<ActivityLog[]>([]);
  const [users, setUsers] = useState<User[]>([]);
  const [meta, setMeta] = useState({ page: 1, per_page: 20, total: 0, total_pages: 1 });
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const fetchUsers = useCallback(async () => {
    if (!storeId) return;
    try {
      const res = await storesApi.listMembers(storeId);
      setUsers(res.data?.data || []);
    } catch (e) {
      console.error('Failed to fetch users', e);
    }
  }, [storeId]);

  const loadLogs = useCallback(
    async (p = 1) => {
      if (!storeId) return;
      setLoading(true);
      setError('');
      try {
        const res = await activityLogsApi.list(storeId, {
          page: p,
          per_page: 20,
          start_date: dateFrom ? `${dateFrom}T00:00:00Z` : undefined,
          end_date: dateTo ? `${dateTo}T23:59:59Z` : undefined,
          user_id: userFilter || undefined,
        });
        const body = res.data as PaginatedData<ActivityLog>;
        setLogs(body.data ?? []);
        setMeta(body.meta ?? { page: p, per_page: 20, total: 0, total_pages: 1 });
        setPage(p);
      } catch (err) {
        if (err instanceof ApiError) setError(err.message);
        else setError('Gagal memuat log aktivitas');
      } finally {
        setLoading(false);
      }
    },
    [storeId, dateFrom, dateTo, userFilter]
  );

  useEffect(() => {
    fetchUsers();
  }, [fetchUsers]);

  useEffect(() => {
    loadLogs(1);
  }, [loadLogs]);

  return (
    <div className="w-full p-6">
      {/* Header */}
      <div
        className="reveal-animate"
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          marginBottom: 32,
        }}
      >
        <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
          <div style={{ background: 'var(--accent-em)', padding: 12, borderRadius: 16 }}>
            <ShieldCheck size={28} color="#fff" />
          </div>
          <div>
            <h1 className="page-title">Activity Log</h1>
            <p className="page-subtitle">Audit trail & traceability system</p>
          </div>
        </div>

        <button className="btn btn-primary" onClick={() => loadLogs(1)} disabled={loading}>
          {loading ? <Clock size={16} className="loading-spin" /> : <Search size={16} />}
          <span>Perbarui Data</span>
        </button>
      </div>

      {/* Filters Card */}
      <div
        className="card reveal-animate"
        style={{ padding: 24, marginBottom: 32, animationDelay: '0.1s' }}
      >
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: 8,
            marginBottom: 24,
            color: 'var(--brand)',
          }}
        >
          <Filter size={18} />
          <span
            style={{
              fontWeight: 700,
              fontSize: '0.75rem',
              textTransform: 'uppercase',
              letterSpacing: '0.1em',
            }}
          >
            Filter Audit
          </span>
        </div>

        <div
          style={{
            display: 'grid',
            gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))',
            gap: 24,
            marginBottom: 16,
          }}
        >
          <div className="input-group">
            <label
              className="input-label"
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: 6,
                fontSize: '0.6875rem',
                fontWeight: 800,
                textTransform: 'uppercase',
                letterSpacing: '0.05em',
              }}
            >
              <CalendarDays size={12} /> Dari Tanggal
            </label>
            <input
              type="date"
              className="input"
              style={{ fontWeight: 600 }}
              value={dateFrom}
              onChange={e => setDateFrom(e.target.value)}
            />
          </div>
          <div className="input-group">
            <label
              className="input-label"
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: 6,
                fontSize: '0.6875rem',
                fontWeight: 800,
                textTransform: 'uppercase',
                letterSpacing: '0.05em',
              }}
            >
              <CalendarDays size={12} /> Sampai Tanggal
            </label>
            <input
              type="date"
              className="input"
              style={{ fontWeight: 600 }}
              value={dateTo}
              onChange={e => setDateTo(e.target.value)}
            />
          </div>
          <div className="input-group">
            <label
              className="input-label"
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: 6,
                fontSize: '0.6875rem',
                fontWeight: 800,
                textTransform: 'uppercase',
                letterSpacing: '0.05em',
              }}
            >
              <UserIcon size={12} /> Pengguna
            </label>
            <select
              className="input"
              style={{ fontWeight: 600 }}
              value={userFilter}
              onChange={e => setUserFilter(e.target.value)}
            >
              <option value="">Semua Pengguna</option>
              {users.map(u => (
                <option key={u.id} value={u.id}>
                  {u.name}
                </option>
              ))}
            </select>
          </div>
        </div>

        {/* Fast Filters */}
        <div style={{ display: 'flex', flexWrap: 'wrap', alignItems: 'center', gap: 12 }}>
          <button
            onClick={() => {
              setDateFrom(today());
              setDateTo(today());
            }}
            className="btn btn-secondary btn-sm"
          >
            Hari Ini
          </button>
          <button
            onClick={() => {
              setDateFrom(daysAgo(7));
              setDateTo(today());
            }}
            className="btn btn-secondary btn-sm"
          >
            Minggu Ini
          </button>
          <button
            onClick={() => {
              setDateFrom(daysAgo(30));
              setDateTo(today());
            }}
            className="btn btn-secondary btn-sm"
          >
            Bulan Ini
          </button>
        </div>
      </div>

      {/* Table Section */}
      <div
        className="card reveal-animate"
        style={{
          overflow: 'hidden',
          minHeight: 400,
          display: 'flex',
          flexDirection: 'column',
          animationDelay: '0.15s',
        }}
      >
        {loading && logs.length === 0 ? (
          <div
            style={{
              flex: 1,
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center',
              justifyContent: 'center',
              padding: 80,
            }}
          >
            <Loader2
              size={32}
              className="loading-spin"
              style={{ color: 'var(--brand)', marginBottom: 16 }}
            />
            <p
              style={{
                color: 'var(--text-3)',
                fontWeight: 700,
                fontSize: '0.75rem',
                textTransform: 'uppercase',
                letterSpacing: '0.1em',
              }}
            >
              Memuat Log...
            </p>
          </div>
        ) : error ? (
          <div
            style={{
              flex: 1,
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center',
              justifyContent: 'center',
              padding: 80,
              textAlign: 'center',
            }}
          >
            <div
              style={{
                padding: 16,
                background: 'rgba(239, 68, 68, 0.1)',
                borderRadius: 16,
                marginBottom: 16,
                color: 'var(--accent-rd)',
              }}
            >
              <AlertCircle size={32} />
            </div>
            <h3
              style={{
                fontSize: '1.125rem',
                fontWeight: 700,
                color: 'var(--text-1)',
                marginBottom: 4,
              }}
            >
              Gagal Memuat Data
            </h3>
            <p style={{ color: 'var(--text-2)', fontSize: '0.875rem', marginBottom: 24 }}>
              {error}
            </p>
            <button onClick={() => loadLogs(1)} className="btn btn-secondary">
              Coba Lagi
            </button>
          </div>
        ) : logs.length === 0 ? (
          <div
            style={{
              flex: 1,
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center',
              justifyContent: 'center',
              padding: 80,
              textAlign: 'center',
            }}
          >
            <div
              style={{
                padding: 16,
                background: 'var(--bg-elevated)',
                borderRadius: 16,
                marginBottom: 16,
                color: 'var(--text-3)',
              }}
            >
              <Search size={32} />
            </div>
            <h3
              style={{
                fontSize: '1.125rem',
                fontWeight: 700,
                color: 'var(--text-1)',
                marginBottom: 4,
              }}
            >
              Log Tidak Ditemukan
            </h3>
            <p style={{ color: 'var(--text-2)', fontSize: '0.875rem' }}>
              Tidak ada aktivitas yang tercatat untuk filter ini.
            </p>
          </div>
        ) : (
          <>
            <div style={{ overflowX: 'auto' }}>
              <table className="tbl">
                <thead>
                  <tr>
                    <th>Waktu</th>
                    <th>Pengguna</th>
                    <th>Modul</th>
                    <th>Aksi</th>
                    <th style={{ textAlign: 'right' }}>Detail</th>
                  </tr>
                </thead>
                <tbody>
                  {logs.map((log, i) => (
                    <LogRow key={log.id} log={log} index={i} />
                  ))}
                </tbody>
              </table>
            </div>

            {/* Pagination */}
            {meta.total_pages > 1 && (
              <div
                style={{
                  padding: 24,
                  borderTop: '1px solid var(--border)',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'space-between',
                  background: 'var(--bg-card)',
                }}
              >
                <div style={{ fontSize: '0.875rem', fontWeight: 500, color: 'var(--text-2)' }}>
                  Menampilkan{' '}
                  <span style={{ color: 'var(--text-1)', fontWeight: 600 }}>{logs.length}</span>{' '}
                  dari <span style={{ color: 'var(--text-1)', fontWeight: 600 }}>{meta.total}</span>{' '}
                  log
                </div>
                <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                  <button
                    className="btn btn-secondary btn-sm"
                    onClick={() => loadLogs(page - 1)}
                    disabled={page === 1 || loading}
                  >
                    <ChevronLeft size={16} />
                  </button>
                  <div
                    style={{
                      padding: '0 16px',
                      fontSize: '0.875rem',
                      fontWeight: 700,
                      color: 'var(--text-1)',
                    }}
                  >
                    Halaman {page} dari {meta.total_pages}
                  </div>
                  <button
                    className="btn btn-secondary btn-sm"
                    onClick={() => loadLogs(page + 1)}
                    disabled={page === meta.total_pages || loading}
                  >
                    <ChevronRight size={16} />
                  </button>
                </div>
              </div>
            )}
          </>
        )}
      </div>
    </div>
  );
}
