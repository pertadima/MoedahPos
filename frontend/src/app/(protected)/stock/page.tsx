'use client';

import { useEffect, useState, useCallback } from 'react';
import { Warehouse, AlertTriangle, Loader2, TrendingUp, TrendingDown } from 'lucide-react';
import { useAuth } from '@/lib/auth/AuthContext';
import { stockApi } from '@/lib/api/store-apis';
import { formatDateTime } from '@/lib/utils';
import type { StockLevel, StockMovement } from '@/types';

export default function StockPage() {
  const { selectedStore } = useAuth();
  const [levels, setLevels] = useState<StockLevel[]>([]);
  const [movements, setMovements] = useState<StockMovement[]>([]);
  const [tab, setTab] = useState<'levels' | 'movements'>('levels');
  const [loading, setLoading] = useState(true);

  const storeId = selectedStore?.store_id;

  const loadData = useCallback(() => {
    if (!storeId) return;
    setLoading(true);
    Promise.all([
      stockApi.levels(storeId),
      stockApi.movements(storeId, { per_page: 30 }),
    ]).then(([l, m]) => {
      setLevels(l.data as StockLevel[]);
      setMovements(((m.data as any).data ?? []) as StockMovement[]);
    }).catch(console.error).finally(() => setLoading(false));
  }, [storeId]);

  // eslint-disable-next-line react-hooks/set-state-in-effect
  useEffect(() => { loadData(); }, [loadData]);

  const lowCount = levels.filter(l => l.is_low_stock).length;

  if (!selectedStore) {
    return <div style={{ padding: 32 }}><div className="empty-state card" style={{ padding: 40 }}><Warehouse size={40} /><p>Pilih toko terlebih dahulu</p></div></div>;
  }

  return (
    <div style={{ padding: 24 }}>
      <div style={{ marginBottom: 20 }}>
        <h1 className="page-title">Manajemen Stok</h1>
        <p className="page-subtitle">{selectedStore.store_name} · {levels.length} produk{lowCount > 0 ? ` · ⚠ ${lowCount} stok menipis` : ''}</p>
      </div>

      {/* Tabs */}
      <div style={{ display: 'flex', gap: 4, background: 'var(--bg-card)', borderRadius: 10, padding: 4, marginBottom: 16, width: 'fit-content', border: '1px solid var(--border)' }}>
        {[['levels','Level Stok'],['movements','Riwayat Mutasi']].map(([v, l]) => (
          <button key={v} onClick={() => setTab(v as any)}
            style={{ padding: '7px 18px', borderRadius: 7, border: 'none', cursor: 'pointer', fontSize: '0.85rem', fontWeight: 500, transition: 'all 0.12s',
              background: tab === v ? 'var(--accent-em)' : 'transparent', color: tab === v ? '#fff' : 'var(--text-2)' }}>
            {l}
          </button>
        ))}
      </div>

      {loading ? (
        <div style={{ display: 'flex', justifyContent: 'center', padding: 40 }}><Loader2 size={24} className="loading-spin" style={{ color: 'var(--accent-em)' }} /></div>
      ) : tab === 'levels' ? (
        <div className="card" style={{ overflow: 'hidden' }}>
          <table className="tbl">
            <thead><tr><th>Produk</th><th>SKU</th><th>Stok</th><th>Min Stok</th><th>Terakhir Update</th><th>Status</th></tr></thead>
            <tbody>
              {levels.map(l => (
                <tr key={l.product_id}>
                  <td style={{ fontWeight: 600 }}>{l.product_name}</td>
                  <td style={{ fontFamily: 'monospace', fontSize: '0.82rem', color: 'var(--text-2)' }}>{l.product_sku}</td>
                  <td style={{ fontWeight: 700, color: l.quantity <= 0 ? 'var(--accent-rd)' : l.is_low_stock ? '#f59e0b' : 'var(--accent-em)' }}>{l.quantity} {l.unit}</td>
                  <td style={{ color: 'var(--text-2)' }}>{l.min_quantity} {l.unit}</td>
                  <td style={{ color: 'var(--text-3)', fontSize: '0.8rem' }}>{formatDateTime(l.updated_at)}</td>
                  <td>
                    {l.quantity <= 0
                      ? <span className="badge badge-red">Habis</span>
                      : l.is_low_stock ? <span className="badge badge-amber"><AlertTriangle size={10} style={{ marginRight: 3 }} />Menipis</span>
                      : <span className="badge badge-green">OK</span>}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      ) : (
        <div className="card" style={{ overflow: 'hidden' }}>
          <table className="tbl">
            <thead><tr><th>Produk</th><th>Tipe</th><th>Delta</th><th>Catatan</th><th>Oleh</th><th>Waktu</th></tr></thead>
            <tbody>
              {movements.map(m => (
                <tr key={m.id}>
                  <td style={{ fontWeight: 600 }}>{m.product_name}</td>
                  <td><span className={`badge ${m.ref_type === 'sale' || m.ref_type === 'void_deduct' ? 'badge-red' : m.ref_type === 'purchase' || m.ref_type === 'void' ? 'badge-green' : 'badge-blue'}`}>{m.ref_type}</span></td>
                  <td style={{ fontWeight: 700, color: m.quantity_delta < 0 ? 'var(--accent-rd)' : 'var(--accent-em)' }}>
                    {m.quantity_delta > 0 ? <TrendingUp size={13} style={{ display: 'inline', marginRight: 4 }} /> : <TrendingDown size={13} style={{ display: 'inline', marginRight: 4 }} />}
                    {m.quantity_delta > 0 ? '+' : ''}{m.quantity_delta}
                  </td>
                  <td style={{ color: 'var(--text-2)', fontSize: '0.82rem' }}>{m.notes || '–'}</td>
                  <td style={{ color: 'var(--text-2)', fontSize: '0.82rem' }}>{m.created_by_name}</td>
                  <td style={{ color: 'var(--text-3)', fontSize: '0.8rem' }}>{formatDateTime(m.created_at)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
