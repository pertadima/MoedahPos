'use client';

import { useEffect, useState } from 'react';
import { BarChart3, Loader2, TrendingUp, Package } from 'lucide-react';
import { useAuth } from '@/lib/auth/AuthContext';
import { reportsApi } from '@/lib/api/store-apis';
import { formatRp, todayStr, thirtyDaysAgoStr } from '@/lib/utils';
import type { SalesSummaryResponse, SalesByProductRow } from '@/types';
import { BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer, CartesianGrid } from 'recharts';

export default function ReportsPage() {
  const { selectedStore } = useAuth();
  const [dateFrom, setDateFrom] = useState(thirtyDaysAgoStr());
  const [dateTo, setDateTo] = useState(todayStr());
  const [summary, setSummary] = useState<SalesSummaryResponse | null>(null);
  const [byProduct, setByProduct] = useState<SalesByProductRow[]>([]);
  const [tab, setTab] = useState<'sales' | 'products' | 'valuation'>('sales');
  const [valuation, setValuation] = useState<any>(null);
  const [loading, setLoading] = useState(false);

  const storeId = selectedStore?.store_id;

  const load = () => {
    if (!storeId) return;
    setLoading(true);
    Promise.all([
      reportsApi.salesSummary(storeId, dateFrom, dateTo),
      reportsApi.byProduct(storeId, dateFrom, dateTo),
      reportsApi.stockValuation(storeId),
    ]).then(([s, bp, sv]) => {
      setSummary(s.data as SalesSummaryResponse);
      setByProduct(bp.data as SalesByProductRow[]);
      setValuation(sv.data);
    }).catch(console.error).finally(() => setLoading(false));
  };

  useEffect(() => { load(); }, [storeId]);

  const chartData = (summary?.rows ?? []).slice().reverse().map(r => ({
    date: r.date.slice(5),
    sales: r.total_sales,
    txn: r.transaction_count,
  }));

  if (!selectedStore) return <div style={{ padding: 32 }}><div className="empty-state card" style={{ padding: 40 }}><BarChart3 size={40} /><p>Pilih toko terlebih dahulu</p></div></div>;

  return (
    <div style={{ padding: 24 }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: 20 }}>
        <div><h1 className="page-title">Laporan</h1><p className="page-subtitle">{selectedStore.store_name}</p></div>
        <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
          <input type="date" className="input" style={{ width: 150 }} value={dateFrom} onChange={e => setDateFrom(e.target.value)} />
          <span className="text-2">s/d</span>
          <input type="date" className="input" style={{ width: 150 }} value={dateTo} onChange={e => setDateTo(e.target.value)} />
          <button className="btn btn-primary" onClick={load}>Tampilkan</button>
        </div>
      </div>

      {/* Summary Cards */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 14, marginBottom: 20 }}>
        <div className="stat-card">
          <div className="stat-icon" style={{ background: 'rgba(16,185,129,0.12)' }}><TrendingUp size={20} style={{ color: '#10b981' }} /></div>
          <div><div className="stat-label">Total Penjualan</div><div className="stat-val">{formatRp(summary?.total_sales ?? 0)}</div></div>
        </div>
        <div className="stat-card">
          <div className="stat-icon" style={{ background: 'rgba(99,102,241,0.12)' }}><BarChart3 size={20} style={{ color: '#6366f1' }} /></div>
          <div><div className="stat-label">Total Transaksi</div><div className="stat-val">{summary?.total_transactions ?? 0}</div></div>
        </div>
        <div className="stat-card">
          <div className="stat-icon" style={{ background: 'rgba(245,158,11,0.12)' }}><Package size={20} style={{ color: '#f59e0b' }} /></div>
          <div><div className="stat-label">Nilai Stok</div><div className="stat-val">{formatRp(valuation?.grand_total ?? 0)}</div></div>
        </div>
      </div>

      {/* Tab Nav */}
      <div style={{ display: 'flex', gap: 4, background: 'var(--bg-card)', borderRadius: 10, padding: 4, marginBottom: 16, width: 'fit-content', border: '1px solid var(--border)' }}>
        {[['sales','Penjualan Harian'],['products','Per Produk'],['valuation','Valuasi Stok']].map(([v, l]) => (
          <button key={v} onClick={() => setTab(v as any)}
            style={{ padding: '7px 16px', borderRadius: 7, border: 'none', cursor: 'pointer', fontSize: '0.85rem', fontWeight: 500, transition: 'all 0.12s',
              background: tab === v ? 'var(--accent-in)' : 'transparent', color: tab === v ? '#fff' : 'var(--text-2)' }}>
            {l}
          </button>
        ))}
      </div>

      {loading ? (
        <div style={{ display: 'flex', justifyContent: 'center', padding: 40 }}><Loader2 size={24} className="loading-spin" style={{ color: 'var(--accent-em)' }} /></div>
      ) : tab === 'sales' ? (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
          {/* Bar Chart */}
          <div className="card" style={{ padding: 20 }}>
            <div style={{ fontWeight: 700, marginBottom: 16 }}>Penjualan Harian</div>
            {chartData.length > 0 ? (
              <ResponsiveContainer width="100%" height={240}>
                <BarChart data={chartData}>
                  <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.05)" />
                  <XAxis dataKey="date" tick={{ fill: 'var(--text-3)', fontSize: 11 }} axisLine={false} tickLine={false} />
                  <YAxis tick={{ fill: 'var(--text-3)', fontSize: 11 }} axisLine={false} tickLine={false}
                    tickFormatter={v => v >= 1000000 ? `${(v/1000000).toFixed(1)}M` : `${(v/1000).toFixed(0)}K`} />
                  <Tooltip contentStyle={{ background: 'var(--bg-elevated)', border: '1px solid var(--border-md)', borderRadius: 8, color: 'var(--text-1)', fontSize: 12 }}
                    formatter={(v: unknown) => [formatRp(Number(v)), 'Penjualan']} />
                  <Bar dataKey="sales" fill="#10b981" radius={[4, 4, 0, 0]} />
                </BarChart>
              </ResponsiveContainer>
            ) : <div className="empty-state" style={{ height: 200 }}><BarChart3 size={32} /><p>Belum ada data</p></div>}
          </div>
          {/* Daily Table */}
          <div className="card" style={{ overflow: 'hidden' }}>
            <table className="tbl">
              <thead><tr><th>Tanggal</th><th>Transaksi</th><th>Penjualan</th><th>Pajak</th><th>Diskon</th><th>Nett</th></tr></thead>
              <tbody>
                {(summary?.rows ?? []).map(r => (
                  <tr key={r.date}>
                    <td style={{ fontFamily: 'monospace' }}>{r.date}</td>
                    <td>{r.transaction_count}</td>
                    <td style={{ fontWeight: 600, color: 'var(--accent-em)' }}>{formatRp(r.total_sales)}</td>
                    <td style={{ color: 'var(--text-2)' }}>{formatRp(r.total_tax)}</td>
                    <td style={{ color: 'var(--accent-rd)' }}>{formatRp(r.total_discount)}</td>
                    <td style={{ fontWeight: 600 }}>{formatRp(r.total_net)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      ) : tab === 'products' ? (
        <div className="card" style={{ overflow: 'hidden' }}>
          <table className="tbl">
            <thead><tr><th>Produk</th><th>SKU</th><th>Qty Terjual</th><th>Revenue</th><th>Pajak</th></tr></thead>
            <tbody>
              {byProduct.map((r, i) => (
                <tr key={r.product_id}>
                  <td><div style={{ display: 'flex', alignItems: 'center', gap: 8 }}><span style={{ background: 'var(--bg-elevated)', width: 24, height: 24, borderRadius: 6, display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: '0.7rem', fontWeight: 700, color: 'var(--text-2)' }}>{i+1}</span>{r.product_name}</div></td>
                  <td style={{ fontFamily: 'monospace', fontSize: '0.82rem', color: 'var(--text-2)' }}>{r.sku}</td>
                  <td style={{ fontWeight: 600 }}>{r.total_quantity}</td>
                  <td style={{ fontWeight: 700, color: 'var(--accent-em)' }}>{formatRp(r.total_revenue)}</td>
                  <td style={{ color: 'var(--text-2)' }}>{formatRp(r.total_tax)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      ) : (
        <div className="card" style={{ overflow: 'hidden' }}>
          <div style={{ padding: '12px 14px', borderBottom: '1px solid var(--border)', display: 'flex', justifyContent: 'space-between', fontWeight: 600 }}>
            <span>Grand Total Valuasi Stok</span>
            <span style={{ color: 'var(--accent-em)' }}>{formatRp(valuation?.grand_total ?? 0)}</span>
          </div>
          <table className="tbl">
            <thead><tr><th>Produk</th><th>SKU</th><th>Stok</th><th>Harga Beli</th><th>Nilai Total</th></tr></thead>
            <tbody>
              {(valuation?.rows ?? []).map((r: any) => (
                <tr key={r.product_id}>
                  <td style={{ fontWeight: 600 }}>{r.product_name}</td>
                  <td style={{ fontFamily: 'monospace', fontSize: '0.82rem', color: 'var(--text-2)' }}>{r.sku}</td>
                  <td>{r.quantity} {r.unit}</td>
                  <td style={{ color: 'var(--text-2)' }}>{formatRp(r.cost_price)}</td>
                  <td style={{ fontWeight: 700, color: 'var(--accent-em)' }}>{formatRp(r.total_value)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
