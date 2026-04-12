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
} from 'lucide-react';
import { useAuth } from '@/lib/auth/AuthContext';
import { activityLogsApi, ActivityLog } from '@/lib/api/activity-logs';
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
};

// ── Components ───────────────────────────────────────────────────────────────

function MetadataViewer({ data }: { data: any }) {
  if (!data) return <span className="text-gray-400 italic">No metadata</span>;
  return (
    <pre className="p-4 bg-gray-900 text-emerald-400 rounded-lg text-xs overflow-x-auto font-mono leading-relaxed border border-gray-800">
      {JSON.stringify(data, null, 2)}
    </pre>
  );
}

function LogRow({ log }: { log: ActivityLog }) {
  const [expanded, setExpanded] = useState(false);
  const action = ACTION_LABELS[log.action_type] || { label: log.action_type, color: 'gray' };

  return (
    <>
      <tr
        className={`hover:bg-gray-50 dark:hover:bg-gray-800/50 cursor-pointer transition-colors ${expanded ? 'bg-gray-50/80 dark:bg-gray-800/30' : ''}`}
        onClick={() => setExpanded(!expanded)}
      >
        <td className="p-4 py-5">
          <div className="flex items-center gap-3">
            <div className="bg-white dark:bg-gray-800 p-2 rounded-lg border border-gray-100 dark:border-gray-700 shadow-sm">
              <Clock size={16} className="text-gray-400" />
            </div>
            <div>
              <div className="font-semibold text-sm">
                {new Date(log.created_at).toLocaleTimeString('id-ID', {
                  hour: '2-digit',
                  minute: '2-digit',
                })}
              </div>
              <div className="text-[10px] text-gray-400 font-medium uppercase tracking-wider">
                {new Date(log.created_at).toLocaleDateString('id-ID', {
                  day: 'numeric',
                  month: 'short',
                })}
              </div>
            </div>
          </div>
        </td>
        <td className="p-4 py-5">
          <div className="flex items-center gap-2">
            <div className="w-8 h-8 rounded-full bg-indigo-50 dark:bg-indigo-900/20 flex items-center justify-center text-indigo-600 dark:text-indigo-400 font-bold text-xs uppercase">
              {log.user_name.charAt(0)}
            </div>
            <span className="text-sm font-medium">{log.user_name}</span>
          </div>
        </td>
        <td className="p-4 py-5">
          <div className="flex items-center gap-2">
            <div className="p-1.5 rounded-md bg-gray-100 dark:bg-gray-800">
              {MODULE_ICONS[log.module] || <Info size={14} />}
            </div>
            <span className="text-xs font-semibold uppercase tracking-wide text-gray-500 dark:text-gray-400">
              {log.module}
            </span>
          </div>
        </td>
        <td className="p-4 py-5">
          <span
            className="px-2.5 py-1 rounded-full text-[11px] font-bold uppercase tracking-tight"
            style={{
              backgroundColor: `${action.color}15`,
              color: action.color,
              border: `1px solid ${action.color}25`,
            }}
          >
            {action.label}
          </span>
        </td>
        <td className="p-4 py-5">
          <div className="flex items-center justify-end">
            {expanded ? (
              <ChevronUp size={18} className="text-gray-400" />
            ) : (
              <ChevronDown size={18} className="text-gray-400" />
            )}
          </div>
        </td>
      </tr>
      {expanded && (
        <tr>
          <td colSpan={5} className="p-0 border-b border-gray-100 dark:border-gray-800">
            <div className="p-6 bg-gray-50/50 dark:bg-gray-900/20 animate-in fade-in slide-in-from-top-2 duration-200">
              <div className="flex flex-col gap-4">
                <div className="flex items-center gap-2 text-xs font-bold text-gray-400 uppercase tracking-widest px-1">
                  <Info size={12} /> Metadata Details
                </div>
                <MetadataViewer data={log.metadata} />
                {log.reference_id && (
                  <div className="flex items-center gap-2 text-[11px] text-gray-500 font-mono bg-white dark:bg-gray-800 px-3 py-1.5 rounded-md border border-gray-100 dark:border-gray-700 w-fit">
                    REFERENCE ID:{' '}
                    <span className="font-bold text-gray-900 dark:text-gray-100">
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
  const [moduleFilter, setModuleFilter] = useState('');
  const [userFilter, setUserFilter] = useState('');
  const [actionFilter, setActionFilter] = useState('');

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
          module: moduleFilter || undefined,
          user_id: userFilter || undefined,
          action_type: actionFilter || undefined,
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
    [storeId, dateFrom, dateTo, moduleFilter, userFilter, actionFilter]
  );

  useEffect(() => {
    fetchUsers();
  }, [fetchUsers]);

  useEffect(() => {
    loadLogs(1);
  }, [loadLogs]);

  return (
    <div className="p-6 max-w-[1400px] w-full mx-auto">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 mb-8">
        <div className="flex items-center gap-4">
          <div className="bg-indigo-600 p-3 rounded-2xl shadow-lg shadow-indigo-200 dark:shadow-none">
            <ShieldCheck size={28} className="text-white" />
          </div>
          <div>
            <h1 className="text-2xl font-black tracking-tight text-gray-900 dark:text-white">
              Activity Log
            </h1>
            <p className="text-gray-500 text-sm font-medium">Audit trail & traceability system</p>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <button
            onClick={() => loadLogs(1)}
            className="btn btn-primary shadow-md"
            disabled={loading}
          >
            {loading ? <Clock size={16} className="animate-spin" /> : <Search size={16} />}
            <span>Perbarui Data</span>
          </button>
        </div>
      </div>

      {/* Filters Card */}
      <div className="bg-white dark:bg-gray-900 border border-gray-100 dark:border-gray-800 rounded-3xl p-6 shadow-sm mb-8">
        <div className="flex items-center gap-2 mb-6 text-indigo-600 dark:text-indigo-400">
          <Filter size={18} />
          <span className="font-bold text-xs uppercase tracking-widest">Filter Audit</span>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-6">
          <div className="space-y-2">
            <label className="text-[11px] font-black uppercase text-gray-400 tracking-wider flex items-center gap-1.5">
              <CalendarDays size={12} /> Dari Tanggal
            </label>
            <input
              type="date"
              className="w-full bg-gray-50 dark:bg-gray-800 border-none rounded-xl text-sm font-semibold p-3 focus:ring-2 focus:ring-indigo-500 transition-all"
              value={dateFrom}
              onChange={e => setDateFrom(e.target.value)}
            />
          </div>
          <div className="space-y-2">
            <label className="text-[11px] font-black uppercase text-gray-400 tracking-wider flex items-center gap-1.5">
              <CalendarDays size={12} /> Sampai Tanggal
            </label>
            <input
              type="date"
              className="w-full bg-gray-50 dark:bg-gray-800 border-none rounded-xl text-sm font-semibold p-3 focus:ring-2 focus:ring-indigo-500 transition-all"
              value={dateTo}
              onChange={e => setDateTo(e.target.value)}
            />
          </div>
          <div className="space-y-2">
            <label className="text-[11px] font-black uppercase text-gray-400 tracking-wider flex items-center gap-1.5">
              <Package size={12} /> Modul
            </label>
            <select
              className="w-full bg-gray-50 dark:bg-gray-800 border-none rounded-xl text-sm font-semibold p-3 focus:ring-2 focus:ring-indigo-500 transition-all"
              value={moduleFilter}
              onChange={e => setModuleFilter(e.target.value)}
            >
              <option value="">Semua Modul</option>
              <option value="AUTH">Authentication</option>
              <option value="TRANSACTION">Transaction</option>
              <option value="DISCOUNT">Discount</option>
              <option value="INVENTORY">Inventory</option>
            </select>
          </div>
          <div className="space-y-2">
            <label className="text-[11px] font-black uppercase text-gray-400 tracking-wider flex items-center gap-1.5">
              <UserIcon size={12} /> Pengguna
            </label>
            <select
              className="w-full bg-gray-50 dark:bg-gray-800 border-none rounded-xl text-sm font-semibold p-3 focus:ring-2 focus:ring-indigo-500 transition-all"
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
          <div className="space-y-2">
            <label className="text-[11px] font-black uppercase text-gray-400 tracking-wider flex items-center gap-1.5">
              <AlertCircle size={12} /> Jenis Aksi
            </label>
            <select
              className="w-full bg-gray-50 dark:bg-gray-800 border-none rounded-xl text-sm font-semibold p-3 focus:ring-2 focus:ring-indigo-500 transition-all"
              value={actionFilter}
              onChange={e => setActionFilter(e.target.value)}
            >
              <option value="">Semua Aksi</option>
              {Object.entries(ACTION_LABELS).map(([k, v]) => (
                <option key={k} value={k}>
                  {v.label}
                </option>
              ))}
            </select>
          </div>
        </div>
      </div>

      {/* Table Section */}
      <div className="bg-white dark:bg-gray-900 border border-gray-100 dark:border-gray-800 rounded-3xl shadow-sm overflow-hidden min-h-[400px] flex flex-col">
        {loading && logs.length === 0 ? (
          <div className="flex-1 flex flex-col items-center justify-center py-20">
            <div className="w-12 h-12 border-4 border-indigo-500/20 border-t-indigo-500 rounded-full animate-spin mb-4" />
            <p className="text-gray-400 font-bold text-xs uppercase tracking-widest">
              Memuat Log...
            </p>
          </div>
        ) : error ? (
          <div className="flex-1 flex flex-col items-center justify-center py-20 text-center px-6">
            <div className="p-4 bg-red-50 dark:bg-red-900/20 rounded-2xl mb-4 text-red-500">
              <AlertCircle size={32} />
            </div>
            <h3 className="text-lg font-bold text-gray-900 dark:text-white mb-1">
              Gagal Memuat Data
            </h3>
            <p className="text-gray-500 text-sm mb-6">{error}</p>
            <button onClick={() => loadLogs(1)} className="btn btn-secondary">
              Coba Lagi
            </button>
          </div>
        ) : logs.length === 0 ? (
          <div className="flex-1 flex flex-col items-center justify-center py-20 text-center px-6">
            <div className="p-4 bg-gray-50 dark:bg-gray-800 rounded-2xl mb-4 text-gray-400">
              <Search size={32} />
            </div>
            <h3 className="text-lg font-bold text-gray-900 dark:text-white mb-1">
              Log Tidak Ditemukan
            </h3>
            <p className="text-gray-500 text-sm">
              Tidak ada aktivitas yang tercatat untuk filter ini.
            </p>
          </div>
        ) : (
          <>
            <div className="overflow-x-auto">
              <table className="w-full text-left border-collapse">
                <thead>
                  <tr className="bg-gray-50/50 dark:bg-gray-800/50 border-b border-gray-100 dark:border-gray-800">
                    <th className="p-4 text-[10px] font-black uppercase text-gray-400 tracking-widest">
                      Waktu
                    </th>
                    <th className="p-4 text-[10px] font-black uppercase text-gray-400 tracking-widest">
                      Pengguna
                    </th>
                    <th className="p-4 text-[10px] font-black uppercase text-gray-400 tracking-widest">
                      Modul
                    </th>
                    <th className="p-4 text-[10px] font-black uppercase text-gray-400 tracking-widest">
                      Aksi
                    </th>
                    <th className="p-4 text-[10px] font-black uppercase text-gray-400 tracking-widest text-right">
                      Detail
                    </th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-50 dark:divide-gray-800">
                  {logs.map(log => (
                    <LogRow key={log.id} log={log} />
                  ))}
                </tbody>
              </table>
            </div>

            {/* Pagination */}
            {meta.total_pages > 1 && (
              <div className="p-6 border-t border-gray-100 dark:border-gray-800 bg-gray-50/30 dark:bg-gray-800/20 flex items-center justify-between">
                <div className="text-sm font-medium text-gray-500">
                  Menampilkan <span className="text-gray-900 dark:text-white">{logs.length}</span>{' '}
                  dari <span className="text-gray-900 dark:text-white">{meta.total}</span> log
                </div>
                <div className="flex items-center gap-2">
                  <button
                    className="p-2.5 rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed transition-all"
                    onClick={() => loadLogs(page - 1)}
                    disabled={page === 1 || loading}
                  >
                    <ChevronLeft size={18} />
                  </button>
                  <div className="px-4 text-sm font-bold">
                    Halaman {page} dari {meta.total_pages}
                  </div>
                  <button
                    className="p-2.5 rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed transition-all"
                    onClick={() => loadLogs(page + 1)}
                    disabled={page === meta.total_pages || loading}
                  >
                    <ChevronRight size={18} />
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
