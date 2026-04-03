'use client';

import { useCallback, useEffect, useState } from 'react';
import { AlertTriangle, Layers, Loader2, TrendingDown, TrendingUp, Warehouse } from 'lucide-react';
import { useAuth } from '@/lib/auth/AuthContext';
import { stockApi } from '@/lib/api/store-apis';
import {
  getBatchSummary,
  listBatches,
  type BatchStockSummary,
  type StockBatch,
} from '@/lib/api/stock-batches';
import { formatDateTime } from '@/lib/utils';
import type { StockLevel, StockMovement } from '@/types';

// ─── Helpers ──────────────────────────────────────────────────────────────────

/** Format a number as Indonesian Rupiah without decimals. */
function formatCurrency(n: number) {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(n);
}

// ─── Tab definition ───────────────────────────────────────────────────────────

type Tab = 'levels' | 'movements' | 'batches' | 'batch-summary';

const TABS: [Tab, string][] = [
  ['levels', 'Level Stok'],
  ['movements', 'Riwayat Mutasi'],
  ['batches', 'Batch FIFO'],
  ['batch-summary', 'Ringkasan Batch'],
];

// ─── Page component ───────────────────────────────────────────────────────────

export default function StockPage() {
  const { selectedStore } = useAuth();

  const [levels, setLevels] = useState<StockLevel[]>([]);
  const [movements, setMovements] = useState<StockMovement[]>([]);
  const [batches, setBatches] = useState<StockBatch[]>([]);
  const [batchSummary, setBatchSummary] = useState<BatchStockSummary[]>([]);

  const [tab, setTab] = useState<Tab>('levels');
  const [loading, setLoading] = useState(true);
  const [productFilter, setProductFilter] = useState('');

  const storeId = selectedStore?.store_id;

  // ── Load stock levels & movements ──────────────────────────────────────────

  const loadData = useCallback(() => {
    if (!storeId) return;
    setLoading(true);
    Promise.all([stockApi.levels(storeId), stockApi.movements(storeId, { per_page: 30 })])
      .then(([l, m]) => {
        setLevels(l.data as StockLevel[]);
        setMovements(((m.data as any).data ?? []) as StockMovement[]); // eslint-disable-line @typescript-eslint/no-explicit-any
      })
      .catch(console.error)
      .finally(() => setLoading(false));
  }, [storeId]);

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    loadData();
  }, [loadData]);

  // ── Load FIFO batch data ───────────────────────────────────────────────────

  const loadBatches = useCallback(() => {
    if (!storeId) return;
    setLoading(true);
    Promise.all([listBatches(storeId, productFilter || undefined), getBatchSummary(storeId)])
      .then(([b, s]) => {
        setBatches(b);
        setBatchSummary(s);
      })
      .catch(console.error)
      .finally(() => setLoading(false));
  }, [storeId, productFilter]);

  // Reload batches when switching to a batch tab or when filter changes.
  useEffect(() => {
    if (tab === 'batches' || tab === 'batch-summary') {
      // eslint-disable-next-line react-hooks/set-state-in-effect
      loadBatches();
    }
  }, [tab, loadBatches]);

  // ── Derived stats ─────────────────────────────────────────────────────────

  const lowCount = levels.filter(l => l.is_low_stock).length;

  // ── No store selected ─────────────────────────────────────────────────────

  if (!selectedStore) {
    return (
      <div style={{ padding: 32 }}>
        <div className="empty-state card" style={{ padding: 40 }}>
          <Warehouse size={40} />
          <p>Pilih toko terlebih dahulu</p>
        </div>
      </div>
    );
  }

  // ── Page ──────────────────────────────────────────────────────────────────

  return (
    <div style={{ padding: 24 }}>
      {/* Header */}
      <div style={{ marginBottom: 20 }}>
        <h1 className="page-title">Manajemen Stok</h1>
        <p className="page-subtitle">
          {selectedStore.store_name} · {levels.length} produk
          {lowCount > 0 ? ` · ⚠ ${lowCount} stok menipis` : ''}
        </p>
      </div>

      {/* Tabs */}
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
        {TABS.map(([v, l]) => (
          <button
            key={v}
            onClick={() => setTab(v)}
            style={{
              padding: '7px 18px',
              borderRadius: 7,
              border: 'none',
              cursor: 'pointer',
              fontSize: '0.85rem',
              fontWeight: 500,
              transition: 'all 0.12s',
              background: tab === v ? 'var(--accent-em)' : 'transparent',
              color: tab === v ? '#fff' : 'var(--text-2)',
            }}
          >
            {l}
          </button>
        ))}
      </div>

      {/* Product filter for batch tabs */}
      {(tab === 'batches' || tab === 'batch-summary') && (
        <div style={{ marginBottom: 12 }}>
          <input
            id="batch-product-filter"
            type="text"
            placeholder="Filter by Product ID (UUID) — leave blank for all"
            value={productFilter}
            onChange={e => setProductFilter(e.target.value)}
            onKeyDown={e => e.key === 'Enter' && loadBatches()}
            style={{
              padding: '7px 12px',
              borderRadius: 8,
              border: '1px solid var(--border)',
              background: 'var(--bg-card)',
              color: 'var(--text-1)',
              fontSize: '0.85rem',
              width: 360,
            }}
          />
          <button
            onClick={loadBatches}
            style={{
              marginLeft: 8,
              padding: '7px 16px',
              borderRadius: 8,
              border: 'none',
              background: 'var(--accent-em)',
              color: '#fff',
              cursor: 'pointer',
              fontSize: '0.85rem',
            }}
          >
            Terapkan
          </button>
        </div>
      )}

      {loading ? (
        <div style={{ display: 'flex', justifyContent: 'center', padding: 40 }}>
          <Loader2 size={24} className="loading-spin" style={{ color: 'var(--accent-em)' }} />
        </div>
      ) : tab === 'levels' ? (
        // ── Stock Levels ────────────────────────────────────────────────────
        <div className="card" style={{ overflow: 'hidden' }}>
          <table className="tbl">
            <thead>
              <tr>
                <th>Produk</th>
                <th>SKU</th>
                <th>Stok</th>
                <th>Min Stok</th>
                <th>Terakhir Update</th>
                <th>Status</th>
              </tr>
            </thead>
            <tbody>
              {levels.map(l => (
                <tr key={l.product_id}>
                  <td style={{ fontWeight: 600 }}>{l.product_name}</td>
                  <td
                    style={{ fontFamily: 'monospace', fontSize: '0.82rem', color: 'var(--text-2)' }}
                  >
                    {l.product_sku}
                  </td>
                  <td
                    style={{
                      fontWeight: 700,
                      color:
                        l.quantity <= 0
                          ? 'var(--accent-rd)'
                          : l.is_low_stock
                            ? '#f59e0b'
                            : 'var(--accent-em)',
                    }}
                  >
                    {l.quantity} {l.unit}
                  </td>
                  <td style={{ color: 'var(--text-2)' }}>
                    {l.min_quantity} {l.unit}
                  </td>
                  <td style={{ color: 'var(--text-3)', fontSize: '0.8rem' }}>
                    {formatDateTime(l.updated_at)}
                  </td>
                  <td>
                    {l.quantity <= 0 ? (
                      <span className="badge badge-red">Habis</span>
                    ) : l.is_low_stock ? (
                      <span className="badge badge-amber">
                        <AlertTriangle size={10} style={{ marginRight: 3 }} />
                        Menipis
                      </span>
                    ) : (
                      <span className="badge badge-green">OK</span>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      ) : tab === 'movements' ? (
        // ── Stock Movements ─────────────────────────────────────────────────
        <div className="card" style={{ overflow: 'hidden' }}>
          <table className="tbl">
            <thead>
              <tr>
                <th>Produk</th>
                <th>Tipe</th>
                <th>Delta</th>
                <th>Catatan</th>
                <th>Oleh</th>
                <th>Waktu</th>
              </tr>
            </thead>
            <tbody>
              {movements.map(m => (
                <tr key={m.id}>
                  <td style={{ fontWeight: 600 }}>{m.product_name}</td>
                  <td>
                    <span
                      className={`badge ${m.ref_type === 'sale' || m.ref_type === 'void_deduct' ? 'badge-red' : m.ref_type === 'purchase' || m.ref_type === 'void' ? 'badge-green' : 'badge-blue'}`}
                    >
                      {m.ref_type}
                    </span>
                  </td>
                  <td
                    style={{
                      fontWeight: 700,
                      color: m.quantity_delta < 0 ? 'var(--accent-rd)' : 'var(--accent-em)',
                    }}
                  >
                    {m.quantity_delta > 0 ? (
                      <TrendingUp size={13} style={{ display: 'inline', marginRight: 4 }} />
                    ) : (
                      <TrendingDown size={13} style={{ display: 'inline', marginRight: 4 }} />
                    )}
                    {m.quantity_delta > 0 ? '+' : ''}
                    {m.quantity_delta}
                  </td>
                  <td style={{ color: 'var(--text-2)', fontSize: '0.82rem' }}>{m.notes || '–'}</td>
                  <td style={{ color: 'var(--text-2)', fontSize: '0.82rem' }}>
                    {m.created_by_name}
                  </td>
                  <td style={{ color: 'var(--text-3)', fontSize: '0.8rem' }}>
                    {formatDateTime(m.created_at)}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      ) : tab === 'batches' ? (
        // ── FIFO Batch Detail ───────────────────────────────────────────────
        <div className="card" style={{ overflow: 'hidden' }}>
          <div
            style={{
              padding: '12px 16px',
              borderBottom: '1px solid var(--border)',
              display: 'flex',
              alignItems: 'center',
              gap: 8,
              color: 'var(--text-2)',
              fontSize: '0.82rem',
            }}
          >
            <Layers size={14} />
            {batches.length} batch aktif · deducted FIFO (received_at ASC)
          </div>
          <table className="tbl">
            <thead>
              <tr>
                <th>Batch ID</th>
                <th>Produk</th>
                <th>SKU</th>
                <th>Sisa Qty</th>
                <th>Harga Beli</th>
                <th>Diterima</th>
                <th>PO ID</th>
              </tr>
            </thead>
            <tbody>
              {batches.length === 0 ? (
                <tr>
                  <td
                    colSpan={7}
                    style={{ textAlign: 'center', color: 'var(--text-3)', padding: 32 }}
                  >
                    Tidak ada batch aktif
                  </td>
                </tr>
              ) : (
                batches.map(b => (
                  <tr key={b.id}>
                    {/* Batch ID — shortened for readability, full UUID in title */}
                    <td title={b.id}>
                      <span
                        style={{
                          fontFamily: 'monospace',
                          fontSize: '0.78rem',
                          color: 'var(--text-3)',
                        }}
                      >
                        {b.id.slice(0, 8)}…
                      </span>
                    </td>
                    <td style={{ fontWeight: 600 }}>{b.product_name}</td>
                    <td
                      style={{
                        fontFamily: 'monospace',
                        fontSize: '0.82rem',
                        color: 'var(--text-2)',
                      }}
                    >
                      {b.product_sku}
                    </td>
                    <td style={{ fontWeight: 700, color: 'var(--accent-em)' }}>
                      {b.quantity_remaining} {b.unit}
                    </td>
                    <td style={{ color: 'var(--text-2)' }}>{formatCurrency(b.purchase_price)}</td>
                    <td style={{ color: 'var(--text-3)', fontSize: '0.8rem' }}>
                      {formatDateTime(b.received_at)}
                    </td>
                    <td title={b.po_id ?? ''}>
                      {b.po_id ? (
                        <span
                          style={{
                            fontFamily: 'monospace',
                            fontSize: '0.78rem',
                            color: 'var(--text-3)',
                          }}
                        >
                          {b.po_id.slice(0, 8)}…
                        </span>
                      ) : (
                        <span style={{ color: 'var(--text-3)' }}>—</span>
                      )}
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      ) : (
        // ── Batch Summary (per-product) ─────────────────────────────────────
        <div className="card" style={{ overflow: 'hidden' }}>
          <div
            style={{
              padding: '12px 16px',
              borderBottom: '1px solid var(--border)',
              display: 'flex',
              alignItems: 'center',
              gap: 8,
              color: 'var(--text-2)',
              fontSize: '0.82rem',
            }}
          >
            <Layers size={14} />
            {batchSummary.length} produk · total stock dari semua batch aktif
          </div>
          <table className="tbl">
            <thead>
              <tr>
                <th>Produk</th>
                <th>SKU</th>
                <th>Total Stok</th>
                <th>Jumlah Batch</th>
                <th>Rata-rata Harga Beli</th>
              </tr>
            </thead>
            <tbody>
              {batchSummary.length === 0 ? (
                <tr>
                  <td
                    colSpan={5}
                    style={{ textAlign: 'center', color: 'var(--text-3)', padding: 32 }}
                  >
                    Tidak ada data batch
                  </td>
                </tr>
              ) : (
                batchSummary.map(s => (
                  <tr key={s.product_id}>
                    <td style={{ fontWeight: 600 }}>{s.product_name}</td>
                    <td
                      style={{
                        fontFamily: 'monospace',
                        fontSize: '0.82rem',
                        color: 'var(--text-2)',
                      }}
                    >
                      {s.product_sku}
                    </td>
                    <td style={{ fontWeight: 700, color: 'var(--accent-em)' }}>
                      {s.total_qty} {s.unit}
                    </td>
                    <td>
                      <span className="badge badge-blue">{s.batch_count} batch</span>
                    </td>
                    <td style={{ color: 'var(--text-2)' }}>{formatCurrency(s.avg_cost_price)}</td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
