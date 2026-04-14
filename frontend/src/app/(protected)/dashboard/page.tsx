'use client';

import { useEffect, useState, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import {
  TrendingUp,
  ShoppingBag,
  Package,
  AlertTriangle,
  ArrowUpRight,
  Loader2,
  Users,
  ArrowUp,
  ArrowDown,
  Wallet,
  LayoutDashboard,
} from 'lucide-react';
import { useAuth } from '@/lib/auth/AuthContext';
import { reportsApi, stockApi, purchaseOrdersApi } from '@/lib/api/store-apis';
import { transactionsApi } from '@/lib/api/transactions';
import { formatRp, formatDateTime, thirtyDaysAgoStr, sevenDaysAgoStr, todayStr } from '@/lib/utils';
import type {
  SalesSummaryResponse,
  Transaction,
  StockLevel,
  SalesByProductRow,
  ProfitSummaryResponse,
  PaginatedData,
} from '@/types';
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  CartesianGrid,
  BarChart,
  Bar,
  Cell,
} from 'recharts';

export default function DashboardPage() {
  type PayablesSummary = {
    overdue_debt?: number;
    due_soon_debt?: number;
    future_debt?: number;
  };
  type CashierRevenueRow = {
    cashier_name: string;
    total_sales: number;
    transaction_count: number;
  };

  const router = useRouter();
  const { selectedStore } = useAuth();
  const [summary, setSummary] = useState<SalesSummaryResponse | null>(null);
  const [recentTxns, setRecentTxns] = useState<Transaction[]>([]);
  const [lowStock, setLowStock] = useState<StockLevel[]>([]);
  const [payables, setPayables] = useState<PayablesSummary | null>(null);
  const [profitData, setProfitData] = useState<ProfitSummaryResponse | null>(null);
  const [loading, setLoading] = useState(true);

  type TimeFilter = 'today' | '7days' | '30days';
  const [cashierFilter, setCashierFilter] = useState<TimeFilter>('30days');
  const [cashierRevenue, setCashierRevenue] = useState<CashierRevenueRow[]>([]);

  const [productFilter, setProductFilter] = useState<TimeFilter>('30days');
  const [productData, setProductData] = useState<SalesByProductRow[]>([]);

  const loadCashierData = useCallback(() => {
    if (!selectedStore) return;
    const sid = selectedStore.store_id;
    let dFrom;
    if (cashierFilter === 'today') dFrom = todayStr();
    else if (cashierFilter === '7days') dFrom = sevenDaysAgoStr();
    else dFrom = thirtyDaysAgoStr();

    reportsApi
      .byCashier(sid, dFrom, todayStr())
      .then(res => setCashierRevenue(res.data || []))
      .catch(console.error);
  }, [selectedStore, cashierFilter]);

  useEffect(() => {
    loadCashierData();
  }, [loadCashierData]);

  const loadProductData = useCallback(() => {
    if (!selectedStore) return;
    const sid = selectedStore.store_id;
    let dFrom;
    if (productFilter === 'today') dFrom = todayStr();
    else if (productFilter === '7days') dFrom = sevenDaysAgoStr();
    else dFrom = thirtyDaysAgoStr();

    reportsApi
      .byProduct(sid, dFrom, todayStr())
      .then(res => setProductData(res.data || []))
      .catch(console.error);
  }, [selectedStore, productFilter]);

  useEffect(() => {
    loadProductData();
  }, [loadProductData]);

  const loadData = useCallback(() => {
    if (!selectedStore) return;
    const sid = selectedStore.store_id;
    setLoading(true);
    Promise.all([
      reportsApi.salesSummary(sid, thirtyDaysAgoStr(), todayStr()),
      transactionsApi.list(sid, { per_page: 5 }),
      stockApi.levels(sid, true),
      purchaseOrdersApi.payableSummary(sid),
      reportsApi.profit(sid, thirtyDaysAgoStr(), todayStr(), 'day'),
    ])
      .then(([s, t, st, p, pr]) => {
        setSummary(s.data as SalesSummaryResponse);
        setRecentTxns((t.data as PaginatedData<Transaction>).data ?? []);
        setLowStock(st.data as StockLevel[]);
        setPayables(p.data as PayablesSummary);
        setProfitData(pr.data as ProfitSummaryResponse);
      })
      .catch(console.error)
      .finally(() => setLoading(false));
  }, [selectedStore]);

  const handleCreatePO = () => {
    if (!selectedStore || lowStock.length === 0) return;
    const items = lowStock.map(s => ({
      product_id: s.product_id,
      product_name: s.product_name,
      product_sku: s.product_sku,
      unit: s.unit,
      quantity: s.min_quantity > 0 ? s.min_quantity : 1,
      unit_cost: s.cost_price || 0,
    }));
    sessionStorage.setItem('openCreatePOWithItems', JSON.stringify(items));
    router.push('/purchase-orders');
  };

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    loadData();
  }, [loadData]);

  /* ── Empty / loading states ─────────────────────────────────────────────── */
  if (!selectedStore) {
    return (
      <div className="p-8">
        <div className="card empty-state" style={{ padding: 48 }}>
          <Package size={40} style={{ color: 'var(--text-3)' }} />
          <p className="type-subheading" style={{ marginTop: 8 }}>
            Pilih toko terlebih dahulu
          </p>
          <p className="type-body-sm">Gunakan selector toko di sidebar untuk memilih toko.</p>
        </div>
      </div>
    );
  }

  if (loading) {
    return (
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          minHeight: '60vh',
        }}
      >
        <Loader2 size={28} className="loading-spin" style={{ color: 'var(--brand)' }} />
      </div>
    );
  }

  /* ── Derived data ─────────────────────────────────────────────────────────── */
  const chartData = (summary?.rows ?? [])
    .slice()
    .reverse()
    .map(r => ({
      date: r.date.slice(5), // MM-DD
      sales: r.total_sales,
      transactions: r.transaction_count,
    }));

  const sortedProducts = [...productData].sort((a, b) => b.total_quantity - a.total_quantity);
  const topProducts = sortedProducts.slice(0, 3);
  const bottomProducts = sortedProducts
    .filter(p => p.total_quantity > 0)
    .reverse()
    .slice(0, 3);

  const statCards = [
    {
      label: 'Total Penjualan (30 hari)',
      value: formatRp(summary?.total_sales ?? 0),
      icon: TrendingUp,
      color: '#10b981',
      bg: 'rgba(16,185,129,0.10)',
    },
    {
      label: 'Total Transaksi (30 hari)',
      value: (summary?.total_transactions ?? 0).toLocaleString('id-ID'),
      icon: ShoppingBag,
      color: '#6366f1',
      bg: 'rgba(99,102,241,0.10)',
    },
    {
      label: 'Stok Menipis',
      value: lowStock.length.toString(),
      icon: AlertTriangle,
      color: '#f59e0b',
      bg: 'rgba(245,158,11,0.10)',
      alert: lowStock.length > 0,
    },
    {
      label: 'Penjualan Hari Ini',
      value: formatRp(summary?.rows?.[0]?.total_sales ?? 0),
      icon: ArrowUpRight,
      color: '#0884f6',
      bg: 'rgba(8,132,246,0.10)',
    },
    {
      label: 'Laba Bersih (30 hari)',
      value: formatRp(profitData?.net_profit ?? 0),
      icon: Wallet,
      color: '#8b5cf6',
      bg: 'rgba(139,92,246,0.10)',
    },
  ];

  /* ── Shared recharts style ─────────────────────────────────────────────── */
  const tooltipStyle = {
    background: 'var(--bg-card)',
    border: '1px solid var(--border-md)',
    borderRadius: 10,
    color: 'var(--text-1)',
    fontSize: 12,
    boxShadow: 'var(--shadow-md)',
  };
  const axisTickStyle = { fill: 'var(--text-3)', fontSize: 11 };
  /* ── Render ────────────────────────────────────────────────────────────── */
  return (
    <div className="w-full" style={{ padding: '24px 28px 40px', maxWidth: 1400, margin: '0 auto' }}>
      {/* ── Page Header ── */}
      <div style={{ marginBottom: 28 }} className="reveal-animate">
        <h1 className="page-title">
          <LayoutDashboard size={20} style={{ color: 'var(--brand)' }} />
          Dashboard
        </h1>
        <p className="page-subtitle">
          {selectedStore.store_name} &middot;{' '}
          {new Date().toLocaleDateString('id-ID', {
            weekday: 'long',
            day: 'numeric',
            month: 'long',
            year: 'numeric',
          })}
        </p>
      </div>

      {/* ── Stat Cards ── */}
      <div
        style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))',
          gap: 14,
          marginBottom: 24,
        }}
      >
        {statCards.map(({ label, value, icon: Icon, color, bg, alert }, i) => {
          const isCurrency = value.startsWith('Rp');
          const displayValue = isCurrency ? value.replace('Rp', '').trim() : value;

          return (
            <div
              key={label}
              className="stat-card reveal-animate"
              style={{ animationDelay: `${0.1 + i * 0.05}s` }}
            >
              <div className="stat-icon" style={{ background: bg }}>
                <Icon size={22} style={{ color }} />
              </div>
              <div style={{ minWidth: 0, flex: 1 }}>
                <div className="stat-label" title={label}>
                  {label}
                </div>
                <div className="stat-val" style={{ color: alert ? '#f59e0b' : 'var(--text-1)' }}>
                  {isCurrency && <span className="stat-currency">Rp</span>}
                  <span className="stat-number">{displayValue}</span>
                </div>
              </div>
            </div>
          );
        })}
      </div>

      {/* ── Charts + Right Sidebar ── */}
      <div
        style={{
          display: 'grid',
          gridTemplateColumns: '1fr 320px',
          gap: 16,
          alignItems: 'start',
        }}
      >
        {/* Left column */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
          {/* Sales trend chart */}
          <div className="card reveal-animate" style={{ padding: 22, animationDelay: '0.35s' }}>
            <div style={{ marginBottom: 18 }}>
              <div className="type-subheading">Tren Penjualan (30 hari)</div>
              <div className="type-caption" style={{ marginTop: 2 }}>
                Total dalam Rupiah
              </div>
            </div>
            {chartData.length > 0 ? (
              <ResponsiveContainer debounce={300} width="100%" height={220}>
                <LineChart data={chartData}>
                  <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" vertical={false} />
                  <XAxis dataKey="date" tick={axisTickStyle} axisLine={false} tickLine={false} />
                  <YAxis
                    tick={axisTickStyle}
                    axisLine={false}
                    tickLine={false}
                    tickFormatter={v =>
                      v >= 1000000 ? `${(v / 1000000).toFixed(1)}M` : `${(v / 1000).toFixed(0)}K`
                    }
                  />
                  <Tooltip
                    contentStyle={tooltipStyle}
                    formatter={(v: unknown) => [formatRp(Number(v)), 'Penjualan']}
                  />
                  <Line
                    type="monotone"
                    dataKey="sales"
                    stroke="#0884f6"
                    strokeWidth={2.5}
                    dot={false}
                    activeDot={{ r: 5, fill: '#0884f6' }}
                  />
                </LineChart>
              </ResponsiveContainer>
            ) : (
              <div className="empty-state" style={{ height: 220 }}>
                <TrendingUp size={32} />
                <p>Belum ada data penjualan</p>
              </div>
            )}
          </div>

          {/* Cashier revenue bar chart */}
          <div className="card reveal-animate" style={{ padding: 22, animationDelay: '0.45s' }}>
            <div
              style={{
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'flex-start',
                marginBottom: 18,
              }}
            >
              <div>
                <div className="type-subheading">Pendapatan per Kasir</div>
                <div className="type-caption" style={{ marginTop: 2 }}>
                  Berdasarkan Transaksi Selesai
                </div>
              </div>
              <select
                className="input"
                style={{ width: 'auto', padding: '5px 10px', fontSize: '0.8rem' }}
                value={cashierFilter}
                onChange={e => setCashierFilter(e.target.value as TimeFilter)}
              >
                <option value="today">Hari Ini</option>
                <option value="7days">7 Hari Terakhir</option>
                <option value="30days">30 Hari Terakhir</option>
              </select>
            </div>
            {cashierRevenue.length > 0 ? (
              <ResponsiveContainer debounce={300} width="100%" height={240}>
                <BarChart
                  data={cashierRevenue}
                  margin={{ left: -20, right: 10, top: 8, bottom: 0 }}
                >
                  <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" vertical={false} />
                  <XAxis
                    dataKey="cashier_name"
                    tick={axisTickStyle}
                    axisLine={false}
                    tickLine={false}
                  />
                  <YAxis
                    tick={axisTickStyle}
                    axisLine={false}
                    tickLine={false}
                    tickFormatter={v =>
                      v >= 1000000 ? `${(v / 1000000).toFixed(1)}M` : `${(v / 1000).toFixed(0)}K`
                    }
                  />
                  <Tooltip
                    cursor={{ fill: 'var(--bg-elevated)' }}
                    contentStyle={tooltipStyle}
                    formatter={(value, _name, item) => {
                      const row = item?.payload as CashierRevenueRow | undefined;
                      return [
                        formatRp(Number(value)),
                        `Pendapatan (${row?.transaction_count ?? 0} Trx)`,
                      ];
                    }}
                  />
                  <Bar dataKey="total_sales" radius={[6, 6, 0, 0]} maxBarSize={52}>
                    {cashierRevenue.map((_, index) => {
                      const colors = [
                        '#0884f6',
                        '#10b981',
                        '#f59e0b',
                        '#8b5cf6',
                        '#f43f5e',
                        '#06b6d4',
                      ];
                      return <Cell key={`cell-${index}`} fill={colors[index % colors.length]} />;
                    })}
                  </Bar>
                </BarChart>
              </ResponsiveContainer>
            ) : (
              <div className="empty-state" style={{ height: 240 }}>
                <Users size={32} style={{ color: 'var(--text-3)' }} />
                <p>Belum ada data kasir untuk periode ini</p>
              </div>
            )}
          </div>

          {/* Product performance */}
          <div className="card reveal-animate" style={{ padding: 22, animationDelay: '0.55s' }}>
            <div
              style={{
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'flex-start',
                marginBottom: 18,
              }}
            >
              <div>
                <div className="type-subheading">Performa Produk</div>
                <div className="type-caption" style={{ marginTop: 2 }}>
                  Produk paling laku &amp; kurang laku
                </div>
              </div>
              <select
                className="input"
                style={{ width: 'auto', padding: '5px 10px', fontSize: '0.8rem' }}
                value={productFilter}
                onChange={e => setProductFilter(e.target.value as TimeFilter)}
              >
                <option value="today">Hari Ini</option>
                <option value="7days">7 Hari Terakhir</option>
                <option value="30days">30 Hari Terakhir</option>
              </select>
            </div>

            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 24 }}>
              {/* Top products */}
              <div>
                <div className="flex items-center gap-1.5" style={{ marginBottom: 12 }}>
                  <ArrowUp size={15} style={{ color: '#10b981' }} />
                  <span className="type-body-sm" style={{ fontWeight: 600 }}>
                    Top 3 Terlaris
                  </span>
                </div>
                <div className="flex flex-col gap-2">
                  {topProducts.length > 0 ? (
                    topProducts.map((p, i) => (
                      <div
                        key={`${p.product_id}-${i}`}
                        className="flex justify-between items-center"
                        style={{
                          padding: '8px 0',
                          borderBottom: '1px solid var(--border)',
                        }}
                      >
                        <div>
                          <div
                            className="type-body-sm"
                            style={{ fontWeight: 600, color: 'var(--text-1)' }}
                          >
                            {p.product_name}
                          </div>
                          <div className="type-caption">{p.total_quantity} terjual</div>
                        </div>
                        <div
                          style={{
                            fontSize: '0.82rem',
                            fontWeight: 700,
                            color: 'var(--brand)',
                            fontVariantNumeric: 'tabular-nums',
                          }}
                        >
                          {formatRp(p.total_revenue)}
                        </div>
                      </div>
                    ))
                  ) : (
                    <div className="type-caption" style={{ fontStyle: 'italic' }}>
                      Belum ada data...
                    </div>
                  )}
                </div>
              </div>

              {/* Bottom products */}
              <div>
                <div className="flex items-center gap-1.5" style={{ marginBottom: 12 }}>
                  <ArrowDown size={15} style={{ color: '#f43f5e' }} />
                  <span className="type-body-sm" style={{ fontWeight: 600 }}>
                    Top 3 Kurang Laku
                  </span>
                </div>
                <div className="flex flex-col gap-2">
                  {bottomProducts.length > 0 ? (
                    bottomProducts.map((p, i) => (
                      <div
                        key={`${p.product_id}-${i}`}
                        className="flex justify-between items-center"
                        style={{
                          padding: '8px 0',
                          borderBottom: '1px solid var(--border)',
                        }}
                      >
                        <div>
                          <div
                            className="type-body-sm"
                            style={{ fontWeight: 600, color: 'var(--text-1)' }}
                          >
                            {p.product_name}
                          </div>
                          <div className="type-caption">{p.total_quantity} terjual</div>
                        </div>
                        <div
                          style={{
                            fontSize: '0.82rem',
                            fontWeight: 700,
                            color: 'var(--text-2)',
                            fontVariantNumeric: 'tabular-nums',
                          }}
                        >
                          {formatRp(p.total_revenue)}
                        </div>
                      </div>
                    ))
                  ) : (
                    <div className="type-caption" style={{ fontStyle: 'italic' }}>
                      Belum ada data...
                    </div>
                  )}
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* ── Right column ── */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
          {/* Low stock alert */}
          {lowStock.length > 0 && (
            <div
              className="card reveal-animate"
              style={{ padding: 18, borderColor: 'rgba(245,158,11,0.25)', animationDelay: '0.65s' }}
            >
              <div className="flex justify-between items-center" style={{ marginBottom: 14 }}>
                <div className="flex items-center gap-2">
                  <AlertTriangle size={15} style={{ color: '#f59e0b' }} />
                  <span style={{ fontWeight: 600, fontSize: '0.85rem', color: '#d97706' }}>
                    Stok Menipis ({lowStock.length})
                  </span>
                </div>
                <button
                  onClick={handleCreatePO}
                  className="btn btn-primary btn-sm"
                  style={{ fontSize: '0.72rem', padding: '4px 10px' }}
                >
                  + Buat PO
                </button>
              </div>
              <div className="flex flex-col gap-2.5">
                {lowStock.slice(0, 5).map(s => (
                  <div key={s.product_id} className="flex justify-between items-center">
                    <span
                      className="type-body-sm"
                      style={{
                        overflow: 'hidden',
                        textOverflow: 'ellipsis',
                        whiteSpace: 'nowrap',
                        maxWidth: '60%',
                      }}
                    >
                      {s.product_name}
                    </span>
                    <span className="badge badge-amber">
                      {s.quantity} {s.unit}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Purchase payables summary */}
          <div className="card reveal-animate" style={{ padding: 18, animationDelay: '0.75s' }}>
            <div className="type-subheading" style={{ marginBottom: 14 }}>
              Hutang Pembelian
            </div>
            <div className="flex flex-col gap-3">
              <div className="flex justify-between items-center">
                <span className="type-body-sm">Jatuh Tempo (Lewat)</span>
                <span className="badge badge-red">{formatRp(payables?.overdue_debt || 0)}</span>
              </div>
              <div className="flex justify-between items-center">
                <span className="type-body-sm">Segera Tempo (7 Hari)</span>
                <span className="badge badge-amber">{formatRp(payables?.due_soon_debt || 0)}</span>
              </div>
              <div className="flex justify-between items-center">
                <span className="type-body-sm">Akan Datang</span>
                <span className="badge badge-gray">{formatRp(payables?.future_debt || 0)}</span>
              </div>
            </div>
          </div>

          {/* Recent transactions */}
          <div
            className="card reveal-animate"
            style={{ padding: 18, flex: 1, animationDelay: '0.85s' }}
          >
            <div className="type-subheading" style={{ marginBottom: 14 }}>
              Transaksi Terakhir
            </div>
            {recentTxns.length === 0 ? (
              <div className="empty-state" style={{ padding: '20px 0' }}>
                <ShoppingBag size={24} />
                <p className="type-body-sm">Belum ada transaksi</p>
              </div>
            ) : (
              <div className="flex flex-col">
                {recentTxns.map(t => (
                  <div
                    key={t.id}
                    className="flex justify-between items-center"
                    style={{
                      padding: '10px 0',
                      borderBottom: '1px solid var(--border)',
                    }}
                  >
                    <div>
                      <div className="type-body-sm" style={{ fontWeight: 600 }}>
                        {t.customer_name || 'Pelanggan'}
                      </div>
                      <div className="type-caption">{formatDateTime(t.created_at)}</div>
                    </div>
                    <div style={{ textAlign: 'right' }}>
                      <div
                        style={{
                          fontSize: '0.82rem',
                          fontWeight: 700,
                          color: 'var(--brand)',
                          fontVariantNumeric: 'tabular-nums',
                        }}
                      >
                        {formatRp(t.total)}
                      </div>
                      <span
                        className={`badge ${t.status === 'completed' ? 'badge-green' : 'badge-red'}`}
                      >
                        {t.status}
                      </span>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
