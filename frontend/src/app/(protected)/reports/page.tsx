'use client';

import { useEffect, useState, useCallback } from 'react';
import { BarChart3, Loader2, TrendingUp, TrendingDown, Activity } from 'lucide-react';
import { useAuth } from '@/lib/auth/AuthContext';
import { reportsApi } from '@/lib/api/store-apis';
import { formatRp, todayStr, thirtyDaysAgoStr } from '@/lib/utils';
import type { SalesSummaryResponse, SalesByProductRow } from '@/types';
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  CartesianGrid,
  ComposedChart,
  Line,
  Area,
} from 'recharts';

type Tab = 'sales' | 'products' | 'profit' | 'valuation';
type GroupBy = 'day' | 'week' | 'month';

// ── Small helpers ─────────────────────────────────────────────────────────────
function ProfitMarginBadge({ margin }: { margin: number }) {
  const color = margin >= 30 ? '#10b981' : margin >= 15 ? '#f59e0b' : '#ef4444';
  return (
    <span
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        gap: 2,
        padding: '2px 8px',
        borderRadius: 6,
        fontSize: '0.75rem',
        fontWeight: 700,
        background: `${color}18`,
        color,
      }}
    >
      {margin.toFixed(1)}%
    </span>
  );
}

const CUSTOM_TOOLTIP_STYLE = {
  background: 'var(--bg-elevated)',
  border: '1px solid var(--border-md)',
  borderRadius: 8,
  color: 'var(--text-1)',
  fontSize: 12,
};

export default function ReportsPage() {
  const { selectedStore } = useAuth();
  const [dateFrom, setDateFrom] = useState(thirtyDaysAgoStr());
  const [dateTo, setDateTo] = useState(todayStr());
  const [tab, setTab] = useState<Tab>('sales');
  const [groupBy, setGroupBy] = useState<GroupBy>('day');

  const [summary, setSummary] = useState<SalesSummaryResponse | null>(null);
  const [byProduct, setByProduct] = useState<SalesByProductRow[]>([]);
  const [valuation, setValuation] = useState<any>(null);
  const [profit, setProfit] = useState<any>(null);
  const [loading, setLoading] = useState(false);

  const storeId = selectedStore?.store_id;

  const load = () => {
    if (!storeId) return;
    setLoading(true);

    const run = async () => {
      const [sRes, bpRes, svRes, prRes] = await Promise.allSettled([
        reportsApi.salesSummary(storeId, dateFrom, dateTo),
        reportsApi.byProduct(storeId, dateFrom, dateTo),
        reportsApi.stockValuation(storeId),
        reportsApi.profit(storeId, dateFrom, dateTo, groupBy),
      ]);
      if (sRes.status === 'fulfilled') setSummary(sRes.value.data as SalesSummaryResponse);
      if (bpRes.status === 'fulfilled')
        setByProduct((bpRes.value.data ?? []) as SalesByProductRow[]);
      if (svRes.status === 'fulfilled') setValuation(svRes.value.data);
      if (prRes.status === 'fulfilled') setProfit(prRes.value.data);
      if (sRes.status === 'rejected') console.error('salesSummary:', sRes.reason);
      if (bpRes.status === 'rejected') console.error('byProduct:', bpRes.reason);
      if (svRes.status === 'rejected') console.error('stockValuation:', svRes.reason);
      if (prRes.status === 'rejected') console.error('profit:', prRes.reason);
      setLoading(false);
    };
    run();
  };

  // Reload profit when groupBy changes while on the profit tab
  const reloadProfit = useCallback(() => {
    if (!storeId || tab !== 'profit') return;
    reportsApi
      .profit(storeId, dateFrom, dateTo, groupBy)
      .then(r => setProfit(r.data))
      .catch(console.error);
  }, [storeId, tab, dateFrom, dateTo, groupBy]);

  // eslint-disable-next-line react-hooks/set-state-in-effect
  useEffect(() => {
    reloadProfit();
  }, [reloadProfit]);

  // eslint-disable-next-line react-hooks/set-state-in-effect
  useEffect(() => {
    load();
  }, [storeId]); // eslint-disable-line react-hooks/exhaustive-deps

  const chartData = (summary?.rows ?? [])
    .slice()
    .reverse()
    .map(r => ({
      date: r.date.slice(5),
      sales: r.total_sales,
      txn: r.transaction_count,
    }));

  const profitChartData = (profit?.rows ?? []).map((r: any) => ({
    period: r.period,
    revenue: r.total_sales,
    cost: r.total_cost,
    profit: r.gross_profit,
    margin: r.profit_margin,
  }));

  if (!selectedStore)
    return (
      <div style={{ padding: 32 }}>
        <div className="empty-state card" style={{ padding: 40 }}>
          <BarChart3 size={40} />
          <p>Pilih toko terlebih dahulu</p>
        </div>
      </div>
    );

  const tabs: { key: Tab; label: string }[] = [
    { key: 'sales', label: 'Penjualan Harian' },
    { key: 'products', label: 'Per Produk' },
    { key: 'profit', label: '💰 Profit' },
    { key: 'valuation', label: 'Valuasi Stok' },
  ];

  return (
    <div style={{ padding: 24 }}>
      {/* ── Header ── */}
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'flex-start',
          marginBottom: 20,
        }}
      >
        <div>
          <h1 className="page-title">Laporan</h1>
          <p className="page-subtitle">{selectedStore.store_name}</p>
        </div>
        <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
          <input
            type="date"
            className="input"
            style={{ width: 150 }}
            value={dateFrom}
            onChange={e => setDateFrom(e.target.value)}
          />
          <span className="text-2">s/d</span>
          <input
            type="date"
            className="input"
            style={{ width: 150 }}
            value={dateTo}
            onChange={e => setDateTo(e.target.value)}
          />
          <button className="btn btn-primary" onClick={load}>
            Tampilkan
          </button>
        </div>
      </div>

      {/* ── Summary Cards ── */}
      <div
        style={{ display: 'grid', gridTemplateColumns: 'repeat(4,1fr)', gap: 12, marginBottom: 20 }}
      >
        <div className="stat-card">
          <div className="stat-icon" style={{ background: 'rgba(16,185,129,0.12)' }}>
            <TrendingUp size={20} style={{ color: '#10b981' }} />
          </div>
          <div>
            <div className="stat-label">Total Penjualan</div>
            <div className="stat-val">{formatRp(summary?.total_sales ?? 0)}</div>
          </div>
        </div>
        <div className="stat-card">
          <div className="stat-icon" style={{ background: 'rgba(99,102,241,0.12)' }}>
            <BarChart3 size={20} style={{ color: '#6366f1' }} />
          </div>
          <div>
            <div className="stat-label">Total Transaksi</div>
            <div className="stat-val">{summary?.total_transactions ?? 0}</div>
          </div>
        </div>
        <div className="stat-card">
          <div className="stat-icon" style={{ background: 'rgba(239,68,68,0.12)' }}>
            <TrendingDown size={20} style={{ color: '#ef4444' }} />
          </div>
          <div>
            <div className="stat-label">Total HPP</div>
            <div className="stat-val" style={{ color: '#ef4444' }}>
              {formatRp((summary as any)?.total_cost ?? 0)}
            </div>
          </div>
        </div>
        <div className="stat-card">
          <div className="stat-icon" style={{ background: 'rgba(245,158,11,0.12)' }}>
            <Activity size={20} style={{ color: '#f59e0b' }} />
          </div>
          <div>
            <div className="stat-label">Gross Profit</div>
            <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginTop: 2 }}>
              <div className="stat-val" style={{ color: '#f59e0b' }}>
                {formatRp((summary as any)?.gross_profit ?? 0)}
              </div>
              {(summary as any)?.profit_margin != null && (
                <ProfitMarginBadge margin={(summary as any).profit_margin} />
              )}
            </div>
          </div>
        </div>
      </div>

      {/* ── Tab Nav ── */}
      <div
        style={{
          display: 'flex',
          gap: 4,
          background: 'var(--bg-card)',
          borderRadius: 10,
          padding: 4,
          marginBottom: 16,
          width: 'fit-content',
          border: '1px solid var(--border)',
        }}
      >
        {tabs.map(({ key, label }) => (
          <button
            key={key}
            onClick={() => setTab(key)}
            style={{
              padding: '7px 16px',
              borderRadius: 7,
              border: 'none',
              cursor: 'pointer',
              fontSize: '0.85rem',
              fontWeight: 500,
              transition: 'all 0.12s',
              background: tab === key ? 'var(--accent-in)' : 'transparent',
              color: tab === key ? '#fff' : 'var(--text-2)',
            }}
          >
            {label}
          </button>
        ))}
      </div>

      {loading ? (
        <div style={{ display: 'flex', justifyContent: 'center', padding: 40 }}>
          <Loader2 size={24} className="loading-spin" style={{ color: 'var(--accent-em)' }} />
        </div>
      ) : tab === 'sales' ? (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
          <div className="card" style={{ padding: 20 }}>
            <div style={{ fontWeight: 700, marginBottom: 16 }}>Penjualan Harian</div>
            {chartData.length > 0 ? (
              <ResponsiveContainer width="100%" height={240}>
                <BarChart data={chartData}>
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
                    contentStyle={CUSTOM_TOOLTIP_STYLE}
                    formatter={(v: unknown) => [formatRp(Number(v)), 'Penjualan']}
                  />
                  <Bar dataKey="sales" fill="#10b981" radius={[4, 4, 0, 0]} />
                </BarChart>
              </ResponsiveContainer>
            ) : (
              <div className="empty-state" style={{ height: 200 }}>
                <BarChart3 size={32} />
                <p>Belum ada data</p>
              </div>
            )}
          </div>
          <div className="card" style={{ overflow: 'hidden' }}>
            <table className="tbl">
              <thead>
                <tr>
                  <th>Tanggal</th>
                  <th>Transaksi</th>
                  <th>Penjualan</th>
                  <th>HPP</th>
                  <th>Profit</th>
                  <th>Margin</th>
                  <th>Pajak</th>
                  <th>Diskon</th>
                </tr>
              </thead>
              <tbody>
                {(summary?.rows ?? []).map(r => (
                  <tr key={r.date}>
                    <td style={{ fontFamily: 'monospace' }}>{r.date}</td>
                    <td>{r.transaction_count}</td>
                    <td style={{ fontWeight: 600, color: 'var(--accent-em)' }}>
                      {formatRp(r.total_sales)}
                    </td>
                    <td style={{ color: '#ef4444' }}>{formatRp((r as any).total_cost ?? 0)}</td>
                    <td style={{ fontWeight: 700, color: '#f59e0b' }}>
                      {formatRp((r as any).gross_profit ?? 0)}
                    </td>
                    <td>
                      <ProfitMarginBadge margin={(r as any).profit_margin ?? 0} />
                    </td>
                    <td style={{ color: 'var(--text-2)' }}>{formatRp(r.total_tax)}</td>
                    <td style={{ color: 'var(--accent-rd)' }}>{formatRp(r.total_discount)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      ) : tab === 'products' ? (
        <div className="card" style={{ overflow: 'hidden' }}>
          <table className="tbl">
            <thead>
              <tr>
                <th>Produk</th>
                <th>SKU</th>
                <th>Qty</th>
                <th>Revenue</th>
                <th>HPP</th>
                <th>Profit</th>
                <th>Margin</th>
                <th>Pajak</th>
              </tr>
            </thead>
            <tbody>
              {byProduct.map((r, i) => (
                <tr key={r.product_id || i}>
                  <td>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                      <span
                        style={{
                          background: 'var(--bg-elevated)',
                          width: 24,
                          height: 24,
                          borderRadius: 6,
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'center',
                          fontSize: '0.7rem',
                          fontWeight: 700,
                          color: 'var(--text-2)',
                        }}
                      >
                        {i + 1}
                      </span>
                      {r.product_name}
                    </div>
                  </td>
                  <td
                    style={{ fontFamily: 'monospace', fontSize: '0.82rem', color: 'var(--text-2)' }}
                  >
                    {r.sku}
                  </td>
                  <td style={{ fontWeight: 600 }}>{r.total_quantity}</td>
                  <td style={{ fontWeight: 700, color: 'var(--accent-em)' }}>
                    {formatRp(r.total_revenue)}
                  </td>
                  <td style={{ color: '#ef4444' }}>{formatRp((r as any).total_cost ?? 0)}</td>
                  <td style={{ fontWeight: 700, color: '#f59e0b' }}>
                    {formatRp((r as any).gross_profit ?? 0)}
                  </td>
                  <td>
                    <ProfitMarginBadge margin={(r as any).profit_margin ?? 0} />
                  </td>
                  <td style={{ color: 'var(--text-2)' }}>{formatRp(r.total_tax)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      ) : tab === 'profit' ? (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
          {/* Period group selector */}
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <span style={{ fontSize: '0.85rem', color: 'var(--text-2)' }}>Kelompokkan per:</span>
            {(['day', 'week', 'month'] as GroupBy[]).map(g => (
              <button
                key={g}
                onClick={() => setGroupBy(g)}
                style={{
                  padding: '5px 14px',
                  borderRadius: 8,
                  cursor: 'pointer',
                  fontSize: '0.82rem',
                  fontWeight: 500,
                  transition: 'all 0.12s',
                  background: groupBy === g ? 'var(--accent-in)' : 'var(--bg-card)',
                  color: groupBy === g ? '#fff' : 'var(--text-2)',
                  border: `1px solid ${groupBy === g ? 'transparent' : 'var(--border)'}`,
                }}
              >
                {{ day: 'Hari', week: 'Minggu', month: 'Bulan' }[g]}
              </button>
            ))}

            {/* period-level summary pills */}
            <div style={{ marginLeft: 'auto', display: 'flex', gap: 10 }}>
              {[
                {
                  label: 'Total Revenue',
                  val: formatRp(profit?.total_sales ?? 0),
                  color: '#10b981',
                },
                { label: 'Total HPP', val: formatRp(profit?.total_cost ?? 0), color: '#ef4444' },
                {
                  label: 'Gross Profit',
                  val: formatRp(profit?.gross_profit ?? 0),
                  color: '#f59e0b',
                },
              ].map(({ label, val, color }) => (
                <div key={label} style={{ textAlign: 'right' }}>
                  <div style={{ fontSize: '0.7rem', color: 'var(--text-3)' }}>{label}</div>
                  <div style={{ fontWeight: 800, fontSize: '0.92rem', color }}>{val}</div>
                </div>
              ))}
              {(profit?.profit_margin ?? 0) > 0 && (
                <div style={{ textAlign: 'right' }}>
                  <div style={{ fontSize: '0.7rem', color: 'var(--text-3)' }}>Margin</div>
                  <ProfitMarginBadge margin={profit.profit_margin} />
                </div>
              )}
            </div>
          </div>

          {/* Stacked area chart */}
          <div className="card" style={{ padding: 20 }}>
            <div style={{ fontWeight: 700, marginBottom: 16 }}>Revenue vs HPP vs Profit</div>
            {profitChartData.length > 0 ? (
              <ResponsiveContainer width="100%" height={260}>
                <ComposedChart data={profitChartData}>
                  <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.05)" />
                  <XAxis
                    dataKey="period"
                    tick={{ fill: 'var(--text-3)', fontSize: 11 }}
                    axisLine={false}
                    tickLine={false}
                  />
                  <YAxis
                    yAxisId="left"
                    tick={{ fill: 'var(--text-3)', fontSize: 11 }}
                    axisLine={false}
                    tickLine={false}
                    tickFormatter={v =>
                      v >= 1000000 ? `${(v / 1000000).toFixed(1)}M` : `${(v / 1000).toFixed(0)}K`
                    }
                  />
                  <YAxis
                    yAxisId="right"
                    orientation="right"
                    tick={{ fill: 'var(--text-3)', fontSize: 11 }}
                    axisLine={false}
                    tickLine={false}
                    tickFormatter={v => `${v}%`}
                    domain={[0, 100]}
                  />
                  <Tooltip
                    contentStyle={CUSTOM_TOOLTIP_STYLE}
                    formatter={(v: unknown, name: unknown) =>
                      name === 'margin'
                        ? [`${Number(v).toFixed(1)}%`, 'Margin']
                        : [
                            formatRp(Number(v)),
                            name === 'revenue' ? 'Revenue' : name === 'cost' ? 'HPP' : 'Profit',
                          ]
                    }
                  />
                  <Area
                    yAxisId="left"
                    type="monotone"
                    dataKey="revenue"
                    fill="rgba(99,102,241,0.1)"
                    stroke="#6366f1"
                    strokeWidth={2}
                    name="revenue"
                  />
                  <Area
                    yAxisId="left"
                    type="monotone"
                    dataKey="cost"
                    fill="rgba(239,68,68,0.1)"
                    stroke="#ef4444"
                    strokeWidth={2}
                    name="cost"
                  />
                  <Bar
                    yAxisId="left"
                    dataKey="profit"
                    fill="rgba(245,158,11,0.8)"
                    radius={[4, 4, 0, 0]}
                    name="profit"
                  />
                  <Line
                    yAxisId="right"
                    type="monotone"
                    dataKey="margin"
                    stroke="#10b981"
                    strokeWidth={2}
                    dot={false}
                    name="margin"
                  />
                </ComposedChart>
              </ResponsiveContainer>
            ) : (
              <div className="empty-state" style={{ height: 200 }}>
                <BarChart3 size={32} />
                <p>Belum ada data profit</p>
              </div>
            )}
          </div>

          {/* Profit table */}
          <div className="card" style={{ overflow: 'hidden' }}>
            <table className="tbl">
              <thead>
                <tr>
                  <th>Periode</th>
                  <th>Revenue</th>
                  <th>HPP (Harga Pokok)</th>
                  <th>Gross Profit</th>
                  <th>Margin</th>
                </tr>
              </thead>
              <tbody>
                {(profit?.rows ?? []).map((r: any) => (
                  <tr key={r.period}>
                    <td style={{ fontFamily: 'monospace', fontWeight: 600 }}>{r.period}</td>
                    <td style={{ color: 'var(--accent-em)', fontWeight: 600 }}>
                      {formatRp(r.total_sales)}
                    </td>
                    <td style={{ color: '#ef4444' }}>{formatRp(r.total_cost)}</td>
                    <td style={{ fontWeight: 800, color: '#f59e0b' }}>
                      {formatRp(r.gross_profit)}
                    </td>
                    <td>
                      <ProfitMarginBadge margin={r.profit_margin} />
                    </td>
                  </tr>
                ))}
                {profit && (
                  <tr style={{ background: 'var(--bg-elevated)', fontWeight: 700 }}>
                    <td>TOTAL</td>
                    <td style={{ color: 'var(--accent-em)' }}>{formatRp(profit.total_sales)}</td>
                    <td style={{ color: '#ef4444' }}>{formatRp(profit.total_cost)}</td>
                    <td style={{ color: '#f59e0b' }}>{formatRp(profit.gross_profit)}</td>
                    <td>
                      <ProfitMarginBadge margin={profit.profit_margin} />
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </div>
      ) : (
        <div className="card" style={{ overflow: 'hidden' }}>
          <div
            style={{
              padding: '12px 14px',
              borderBottom: '1px solid var(--border)',
              display: 'flex',
              justifyContent: 'space-between',
              fontWeight: 600,
            }}
          >
            <span>Grand Total Valuasi Stok</span>
            <span style={{ color: 'var(--accent-em)' }}>
              {formatRp(valuation?.grand_total ?? 0)}
            </span>
          </div>
          <table className="tbl">
            <thead>
              <tr>
                <th>Produk</th>
                <th>SKU</th>
                <th>Stok</th>
                <th>Harga Beli</th>
                <th>Nilai Total</th>
              </tr>
            </thead>
            <tbody>
              {(valuation?.rows ?? []).map((r: any) => (
                <tr key={r.product_id}>
                  <td style={{ fontWeight: 600 }}>{r.product_name}</td>
                  <td
                    style={{ fontFamily: 'monospace', fontSize: '0.82rem', color: 'var(--text-2)' }}
                  >
                    {r.sku}
                  </td>
                  <td>
                    {r.quantity} {r.unit}
                  </td>
                  <td style={{ color: 'var(--text-2)' }}>{formatRp(r.cost_price)}</td>
                  <td style={{ fontWeight: 700, color: 'var(--accent-em)' }}>
                    {formatRp(r.total_value)}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
