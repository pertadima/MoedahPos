'use client';

import { useEffect, useState, useCallback } from 'react';
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
} from 'lucide-react';
import { useAuth } from '@/lib/auth/AuthContext';
import { reportsApi, stockApi, purchaseOrdersApi } from '@/lib/api/store-apis';
import { transactionsApi } from '@/lib/api/transactions';
import { formatRp, formatDateTime, thirtyDaysAgoStr, sevenDaysAgoStr, todayStr } from '@/lib/utils';
import type { SalesSummaryResponse, Transaction, StockLevel } from '@/types';
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
  const { selectedStore } = useAuth();
  const [summary, setSummary] = useState<SalesSummaryResponse | null>(null);
  const [recentTxns, setRecentTxns] = useState<Transaction[]>([]);
  const [lowStock, setLowStock] = useState<StockLevel[]>([]);
  const [payables, setPayables] = useState<any>(null);
  const [loading, setLoading] = useState(true);

  type TimeFilter = 'today' | '7days' | '30days';
  const [cashierFilter, setCashierFilter] = useState<TimeFilter>('30days');
  const [cashierRevenue, setCashierRevenue] = useState<any[]>([]);

  const [productFilter, setProductFilter] = useState<TimeFilter>('30days');
  const [productData, setProductData] = useState<any[]>([]);

  const loadCashierData = useCallback(() => {
    if (!selectedStore) return;
    const sid = selectedStore.store_id;
    let dFrom;
    if (cashierFilter === 'today') dFrom = todayStr();
    else if (cashierFilter === '7days') dFrom = sevenDaysAgoStr();
    else dFrom = thirtyDaysAgoStr();

    reportsApi.byCashier(sid, dFrom, todayStr())
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

    reportsApi.byProduct(sid, dFrom, todayStr())
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
    ])
      .then(([s, t, st, p]) => {
        setSummary(s.data as SalesSummaryResponse);
        setRecentTxns((t.data as any).data ?? []);
        setLowStock(st.data as StockLevel[]);
        setPayables(p.data);
      })
      .catch(console.error)
      .finally(() => setLoading(false));
  }, [selectedStore]);

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    loadData();
  }, [loadData]);

  if (!selectedStore) {
    return (
      <div style={{ padding: 32 }}>
        <div className="empty-state card" style={{ padding: 40 }}>
          <Package size={40} style={{ color: 'var(--text-3)' }} />
          <p style={{ fontWeight: 600, color: 'var(--text-2)' }}>Pilih toko terlebih dahulu</p>
          <p style={{ fontSize: '0.85rem' }}>
            Gunakan selector toko di sidebar untuk memilih toko.
          </p>
        </div>
      </div>
    );
  }

  if (loading) {
    return (
      <div
        style={{
          padding: 32,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          minHeight: 300,
        }}
      >
        <Loader2 size={28} className="loading-spin" style={{ color: 'var(--accent-em)' }} />
      </div>
    );
  }

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
  const bottomProducts = sortedProducts.filter(p => p.total_quantity > 0).reverse().slice(0, 3);

  const statCards = [
    {
      label: 'Total Penjualan (30 hari)',
      value: formatRp(summary?.total_sales ?? 0),
      icon: TrendingUp,
      color: '#10b981',
      bg: 'rgba(16,185,129,0.12)',
    },
    {
      label: 'Total Transaksi (30 hari)',
      value: (summary?.total_transactions ?? 0).toLocaleString('id-ID'),
      icon: ShoppingBag,
      color: '#6366f1',
      bg: 'rgba(99,102,241,0.12)',
    },
    {
      label: 'Stok Menipis',
      value: lowStock.length.toString(),
      icon: AlertTriangle,
      color: '#f59e0b',
      bg: 'rgba(245,158,11,0.12)',
      alert: lowStock.length > 0,
    },
    {
      label: 'Penjualan Hari Ini',
      value: formatRp(summary?.rows?.[0]?.total_sales ?? 0),
      icon: ArrowUpRight,
      color: '#10b981',
      bg: 'rgba(16,185,129,0.12)',
    },
  ];

  return (
    <div style={{ padding: '24px' }}>
      {/* Header */}
      <div style={{ marginBottom: 24 }}>
        <h1 className="page-title">Dashboard</h1>
        <p className="page-subtitle">
          {selectedStore.store_name} ·{' '}
          {new Date().toLocaleDateString('id-ID', {
            weekday: 'long',
            day: 'numeric',
            month: 'long',
            year: 'numeric',
          })}
        </p>
      </div>

      {/* Stat Cards */}
      <div
        style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))',
          gap: 14,
          marginBottom: 24,
        }}
      >
        {statCards.map(({ label, value, icon: Icon, color, bg, alert }) => (
          <div key={label} className="stat-card">
            <div className="stat-icon" style={{ background: bg }}>
              <Icon size={20} style={{ color }} />
            </div>
            <div>
              <div className="stat-label">{label}</div>
              <div className="stat-val" style={{ color: alert ? '#f59e0b' : undefined }}>
                {value}
              </div>
            </div>
          </div>
        ))}
      </div>

      {/* Charts + Tables */}
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 340px', gap: 16 }}>
        {/* Left column */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
          {/* Sales Chart */}
          <div className="card" style={{ padding: '20px' }}>
            <div style={{ marginBottom: 16 }}>
              <div style={{ fontWeight: 700, fontSize: '0.95rem' }}>Tren Penjualan (30 hari)</div>
              <div className="text-3" style={{ fontSize: '0.8rem' }}>
                Total dalam Rupiah
              </div>
            </div>
            {chartData.length > 0 ? (
              <ResponsiveContainer width="100%" height={220}>
                <LineChart data={chartData}>
                  <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.05)" />
                  <XAxis
                    dataKey="date"
                    tick={{ fill: 'var(--text-3)', fontSize: 11 }}
                    axisLine={false}
                    tickLine={false}
                  />
                  <YAxis
                    tick={{ fill: 'var(--text-3)', fontSize: 11 }}
                    axisLine={false}
                    tickLine={false}
                    tickFormatter={v =>
                      v >= 1000000 ? `${(v / 1000000).toFixed(1)}M` : `${(v / 1000).toFixed(0)}K`
                    }
                  />
                  <Tooltip
                    contentStyle={{
                      background: 'var(--bg-elevated)',
                      border: '1px solid var(--border-md)',
                      borderRadius: 8,
                      color: 'var(--text-1)',
                      fontSize: 12,
                    }}
                    formatter={(v: unknown) => [formatRp(Number(v)), 'Penjualan']}
                  />
                  <Line
                    type="monotone"
                    dataKey="sales"
                    stroke="#10b981"
                    strokeWidth={2}
                    dot={false}
                    activeDot={{ r: 4 }}
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

          {/* Cashier Revenue Chart */}
          <div className="card" style={{ padding: '20px' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: 16 }}>
              <div>
                <div style={{ fontWeight: 700, fontSize: '0.95rem' }}>Pendapatan per Kasir</div>
                <div className="text-3" style={{ fontSize: '0.8rem' }}>Berdasarkan Transaksi Selesai</div>
              </div>
              <select
                className="input"
                style={{ width: 'auto', padding: '4px 8px', fontSize: '0.8rem' }}
                value={cashierFilter}
                onChange={e => setCashierFilter(e.target.value as TimeFilter)}
              >
                <option value="today">Hari Ini</option>
                <option value="7days">7 Hari Terakhir</option>
                <option value="30days">30 Hari Terakhir</option>
              </select>
            </div>
            {cashierRevenue.length > 0 ? (
              <ResponsiveContainer width="100%" height={260}>
                <BarChart data={cashierRevenue} margin={{ left: -20, right: 10, top: 10, bottom: 0 }}>
                  <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.05)" vertical={false} />
                  <XAxis
                    dataKey="cashier_name"
                    tick={{ fill: 'var(--text-3)', fontSize: 11 }}
                    axisLine={false}
                    tickLine={false}
                  />
                  <YAxis
                    tick={{ fill: 'var(--text-3)', fontSize: 11 }}
                    axisLine={false}
                    tickLine={false}
                    tickFormatter={v =>
                      v >= 1000000 ? `${(v / 1000000).toFixed(1)}M` : `${(v / 1000).toFixed(0)}K`
                    }
                  />
                  <Tooltip
                    cursor={{ fill: 'rgba(255,255,255,0.05)' }}
                    contentStyle={{
                      background: 'var(--bg-elevated)',
                      border: '1px solid var(--border-md)',
                      borderRadius: 8,
                      color: 'var(--text-1)',
                      fontSize: 12,
                    }}
                    formatter={(v: any, name: any, props: any) => [
                      formatRp(Number(v)),
                      `Pendapatan (${props.payload.transaction_count} Trx)`
                    ]}
                  />
                  <Bar dataKey="total_sales" radius={[4, 4, 0, 0]} maxBarSize={60}>
                    {cashierRevenue.map((_, index) => {
                      const colors = ['#10b981', '#3b82f6', '#f43f5e', '#f59e0b', '#8b5cf6', '#06b6d4'];
                      return <Cell key={`cell-${index}`} fill={colors[index % colors.length]} />;
                    })}
                  </Bar>
                </BarChart>
              </ResponsiveContainer>
            ) : (
              <div className="empty-state" style={{ height: 260 }}>
                <Users size={32} style={{ color: 'var(--text-3)' }} />
                <p>Belum ada data kasir untuk periode ini</p>
              </div>
            )}
          </div>

          {/* Most & Least Selling Products */}
          <div className="card" style={{ padding: '20px' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: 16 }}>
              <div>
                <div style={{ fontWeight: 700, fontSize: '0.95rem' }}>Performa Produk</div>
                <div className="text-3" style={{ fontSize: '0.8rem' }}>Produk paling laku & kurang laku</div>
              </div>
              <select
                className="input"
                style={{ width: 'auto', padding: '4px 8px', fontSize: '0.8rem' }}
                value={productFilter}
                onChange={e => setProductFilter(e.target.value as TimeFilter)}
              >
                <option value="today">Hari Ini</option>
                <option value="7days">7 Hari Terakhir</option>
                <option value="30days">30 Hari Terakhir</option>
              </select>
            </div>

            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 24 }}>
              {/* Top Products */}
              <div>
                <div style={{ display: 'flex', alignItems: 'center', gap: 6, marginBottom: 12 }}>
                  <ArrowUp size={16} style={{ color: 'var(--accent-em)' }} />
                  <span style={{ fontWeight: 600, fontSize: '0.85rem' }}>Top 3 Terlaris</span>
                </div>
                <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
                  {topProducts.length > 0 ? topProducts.map(p => (
                    <div key={p.product_id} style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '8px 0', borderBottom: '1px solid var(--border)' }}>
                      <div>
                        <div style={{ fontSize: '0.8rem', fontWeight: 600 }}>{p.product_name}</div>
                        <div className="text-3" style={{ fontSize: '0.72rem' }}>{p.total_quantity} terjual</div>
                      </div>
                      <div style={{ fontSize: '0.82rem', fontWeight: 700, color: 'var(--accent-em)' }}>
                        {formatRp(p.total_revenue)}
                      </div>
                    </div>
                  )) : (
                    <div className="text-3" style={{ fontSize: '0.8rem', fontStyle: 'italic' }}>Belum ada data...</div>
                  )}
                </div>
              </div>

              {/* Bottom Products */}
              <div>
                <div style={{ display: 'flex', alignItems: 'center', gap: 6, marginBottom: 12 }}>
                  <ArrowDown size={16} style={{ color: '#f43f5e' }} />
                  <span style={{ fontWeight: 600, fontSize: '0.85rem' }}>Top 3 Kurang Laku</span>
                </div>
                <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
                  {bottomProducts.length > 0 ? bottomProducts.map(p => (
                    <div key={p.product_id} style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '8px 0', borderBottom: '1px solid var(--border)' }}>
                      <div>
                        <div style={{ fontSize: '0.8rem', fontWeight: 600 }}>{p.product_name}</div>
                        <div className="text-3" style={{ fontSize: '0.72rem' }}>{p.total_quantity} terjual</div>
                      </div>
                      <div style={{ fontSize: '0.82rem', fontWeight: 700 }}>
                        {formatRp(p.total_revenue)}
                      </div>
                    </div>
                  )) : (
                    <div className="text-3" style={{ fontSize: '0.8rem', fontStyle: 'italic' }}>Belum ada data...</div>
                  )}
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Right column */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
          {/* Low Stock Alert */}
          {lowStock.length > 0 && (
            <div className="card" style={{ padding: 16, borderColor: 'rgba(245,158,11,0.3)' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 6, marginBottom: 12 }}>
                <AlertTriangle size={15} style={{ color: '#f59e0b' }} />
                <span style={{ fontWeight: 600, fontSize: '0.85rem', color: '#fbbf24' }}>
                  Stok Menipis ({lowStock.length})
                </span>
              </div>
              <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
                {lowStock.slice(0, 4).map(s => (
                  <div
                    key={s.product_id}
                    style={{
                      display: 'flex',
                      justifyContent: 'space-between',
                      alignItems: 'center',
                      fontSize: '0.8rem',
                    }}
                  >
                    <span
                      style={{
                        color: 'var(--text-1)',
                        overflow: 'hidden',
                        textOverflow: 'ellipsis',
                        whiteSpace: 'nowrap',
                        maxWidth: '65%',
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

          {/* PO Debt summary */}
          <div className="card" style={{ padding: 16 }}>
            <div style={{ fontWeight: 700, fontSize: '0.85rem', marginBottom: 12 }}>
              Hutang Purchase Order
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.8rem' }}>
                <span style={{ color: 'var(--text-2)' }}>Jatuh Tempo (Lewat)</span>
                <span className="badge badge-red">{formatRp(payables?.overdue_debt || 0)}</span>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.8rem' }}>
                <span style={{ color: 'var(--text-2)' }}>Segera Tempo (7 Hari)</span>
                <span className="badge badge-amber">{formatRp(payables?.due_soon_debt || 0)}</span>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.8rem' }}>
                <span style={{ color: 'var(--text-2)' }}>Akan Datang</span>
                <span className="badge badge-gray" style={{ color: 'var(--text-2)' }}>{formatRp(payables?.future_debt || 0)}</span>
              </div>
            </div>
          </div>

          {/* Recent Transactions */}
          <div className="card" style={{ padding: 16, flex: 1 }}>
            <div style={{ fontWeight: 700, fontSize: '0.85rem', marginBottom: 12 }}>
              Transaksi Terakhir
            </div>
            {recentTxns.length === 0 ? (
              <div className="empty-state" style={{ padding: '20px 0' }}>
                <ShoppingBag size={24} />
                <p style={{ fontSize: '0.8rem' }}>Belum ada transaksi</p>
              </div>
            ) : (
              <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
                {recentTxns.map(t => (
                  <div
                    key={t.id}
                    style={{
                      display: 'flex',
                      justifyContent: 'space-between',
                      alignItems: 'center',
                      padding: '8px 0',
                      borderBottom: '1px solid var(--border)',
                    }}
                  >
                    <div>
                      <div style={{ fontSize: '0.8rem', fontWeight: 600 }}>
                        {t.customer_name || 'Pelanggan'}
                      </div>
                      <div className="text-3" style={{ fontSize: '0.72rem' }}>
                        {formatDateTime(t.created_at)}
                      </div>
                    </div>
                    <div style={{ textAlign: 'right' }}>
                      <div
                        style={{ fontSize: '0.82rem', fontWeight: 700, color: 'var(--accent-em)' }}
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
