'use client';

import { useCallback, useEffect, useState } from 'react';
import {
  AlertTriangle,
  ChevronDown,
  ChevronRight,
  Layers,
  Loader2,
  TrendingDown,
  TrendingUp,
  Warehouse,
} from 'lucide-react';
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

function formatCurrency(n: number) {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(n);
}

// ─── Types ────────────────────────────────────────────────────────────────────

type Tab = 'stok' | 'movements';

const TABS: [Tab, string][] = [
  ['stok', 'Stok & Batch'],
  ['movements', 'Riwayat Mutasi'],
];

// ─── Expandable Stock Row ─────────────────────────────────────────────────────

interface StockRowProps {
  level: StockLevel;
  summary?: BatchStockSummary;
  batches: StockBatch[];
  isExpanded: boolean;
  onToggle: () => void;
}

function StockRow({ level, summary, batches, isExpanded, onToggle }: StockRowProps) {
  const hasBatches = batches.length > 0;

  return (
    <>
      {/* ── Main product row ─────────────────────────────── */}
      <tr
        onClick={hasBatches ? onToggle : undefined}
        style={{
          cursor: hasBatches ? 'pointer' : 'default',
          background: isExpanded ? 'var(--bg-hover, rgba(255,255,255,0.04))' : undefined,
          transition: 'background 0.12s',
        }}
      >
        {/* Expand icon */}
        <td style={{ width: 36, textAlign: 'center', color: 'var(--text-3)' }}>
          {hasBatches ? (
            isExpanded ? (
              <ChevronDown size={14} />
            ) : (
              <ChevronRight size={14} />
            )
          ) : (
            <span style={{ display: 'inline-block', width: 14 }} />
          )}
        </td>

        {/* Product name */}
        <td style={{ fontWeight: 600 }}>{level.product_name}</td>

        {/* SKU */}
        <td style={{ fontFamily: 'monospace', fontSize: '0.82rem', color: 'var(--text-2)' }}>
          {level.product_sku}
        </td>

        {/* Current stock (from stock_levels — authoritative) */}
        <td
          style={{
            fontWeight: 700,
            color:
              level.quantity <= 0
                ? 'var(--accent-rd)'
                : level.is_low_stock
                  ? '#f59e0b'
                  : 'var(--accent-em)',
          }}
        >
          {level.quantity} {level.unit}
        </td>

        {/* Batch qty (from FIFO batches) */}
        <td style={{ color: 'var(--text-2)' }}>
          {summary ? (
            <span style={{ display: 'flex', alignItems: 'center', gap: 5 }}>
              <Layers size={12} style={{ color: 'var(--text-3)' }} />
              {summary.total_qty} {level.unit}
              <span
                style={{
                  background: 'var(--bg-hover, rgba(255,255,255,0.07))',
                  borderRadius: 5,
                  padding: '1px 6px',
                  fontSize: '0.74rem',
                  color: 'var(--text-3)',
                }}
              >
                {summary.batch_count} batch
              </span>
            </span>
          ) : (
            <span style={{ color: 'var(--text-3)', fontSize: '0.8rem' }}>—</span>
          )}
        </td>

        {/* Avg cost price */}
        <td style={{ color: 'var(--text-2)', fontSize: '0.85rem' }}>
          {summary ? formatCurrency(summary.avg_cost_price) : '—'}
        </td>

        {/* Min stock */}
        <td style={{ color: 'var(--text-2)' }}>
          {level.min_quantity} {level.unit}
        </td>

        {/* Status */}
        <td>
          {level.quantity <= 0 ? (
            <span className="badge badge-red">Habis</span>
          ) : level.is_low_stock ? (
            <span className="badge badge-amber">
              <AlertTriangle size={10} style={{ marginRight: 3 }} />
              Menipis
            </span>
          ) : (
            <span className="badge badge-green">OK</span>
          )}
        </td>
      </tr>

      {/* ── Expanded batch sub-rows ──────────────────────── */}
      {isExpanded &&
        batches.map((b, idx) => (
          <tr
            key={b.id}
            style={{
              background: 'var(--bg-hover, rgba(255,255,255,0.025))',
              borderLeft: '3px solid var(--accent-em)',
            }}
          >
            {/* indent + batch number */}
            <td style={{ textAlign: 'center', color: 'var(--text-3)', fontSize: '0.75rem' }}>
              #{idx + 1}
            </td>

            {/* Batch ID (shortened) */}
            <td colSpan={2}>
              <span
                title={b.id}
                style={{
                  fontFamily: 'monospace',
                  fontSize: '0.78rem',
                  color: 'var(--text-3)',
                  paddingLeft: 12,
                }}
              >
                {b.id.slice(0, 8)}…
                {b.po_id && (
                  <span style={{ marginLeft: 8, color: 'var(--text-3)' }}>
                    PO: {b.po_id.slice(0, 8)}…
                  </span>
                )}
              </span>
            </td>

            {/* Qty remaining */}
            <td style={{ fontWeight: 600, color: 'var(--accent-em)', fontSize: '0.85rem' }}>
              {b.quantity_remaining} {b.unit}
            </td>

            {/* Batch qty col (reused for FIFO indicator) */}
            <td style={{ fontSize: '0.78rem', color: 'var(--text-3)' }}>
              <span
                style={{
                  background: 'rgba(59,130,246,0.12)',
                  color: '#60a5fa',
                  borderRadius: 4,
                  padding: '1px 7px',
                  fontSize: '0.74rem',
                }}
              >
                FIFO #{idx + 1}
              </span>
            </td>

            {/* Purchase price */}
            <td style={{ color: 'var(--text-2)', fontSize: '0.85rem' }}>
              {formatCurrency(b.purchase_price)}
            </td>

            {/* Received at */}
            <td style={{ color: 'var(--text-3)', fontSize: '0.78rem' }}>
              {formatDateTime(b.received_at)}
            </td>

            {/* Empty status cell */}
            <td />
          </tr>
        ))}
    </>
  );
}

// ─── Page ─────────────────────────────────────────────────────────────────────

export default function StockPage() {
  const { selectedStore } = useAuth();

  const [levels, setLevels] = useState<StockLevel[]>([]);
  const [movements, setMovements] = useState<StockMovement[]>([]);
  const [batches, setBatches] = useState<StockBatch[]>([]);
  const [batchSummary, setBatchSummary] = useState<BatchStockSummary[]>([]);

  const [tab, setTab] = useState<Tab>('stok');
  const [loading, setLoading] = useState(true);
  const [expanded, setExpanded] = useState<Set<string>>(new Set());

  const storeId = selectedStore?.store_id;

  // ── Data loaders ────────────────────────────────────────────────────────────

  const loadAll = useCallback(() => {
    if (!storeId) return;
    setLoading(true);
    Promise.all([
      stockApi.levels(storeId),
      stockApi.movements(storeId, { per_page: 30 }),
      listBatches(storeId),
      getBatchSummary(storeId),
    ])
      .then(([l, m, b, s]) => {
        setLevels(l.data as StockLevel[]);
        setMovements(
          // eslint-disable-next-line @typescript-eslint/no-explicit-any
          ((m.data as any).data ?? []) as StockMovement[]
        );
        setBatches(b);
        setBatchSummary(s);
      })
      .catch(console.error)
      .finally(() => setLoading(false));
  }, [storeId]);

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    loadAll();
  }, [loadAll]);

  // ── Derived maps ────────────────────────────────────────────────────────────

  // summary map: productId → BatchStockSummary
  const summaryMap = new Map<string, BatchStockSummary>(batchSummary.map(s => [s.product_id, s]));

  // batches map: productId → StockBatch[] (oldest first = FIFO order)
  const batchMap = new Map<string, StockBatch[]>();
  for (const b of batches) {
    const arr = batchMap.get(b.product_id) ?? [];
    arr.push(b);
    batchMap.set(b.product_id, arr);
  }

  const lowCount = levels.filter(l => l.is_low_stock).length;

  const toggleExpand = (productId: string) => {
    setExpanded(prev => {
      const next = new Set(prev);
      if (next.has(productId)) next.delete(productId);
      else next.add(productId);
      return next;
    });
  };

  // ── No store selected ──────────────────────────────────────────────────────

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

      {loading ? (
        <div style={{ display: 'flex', justifyContent: 'center', padding: 40 }}>
          <Loader2 size={24} className="loading-spin" style={{ color: 'var(--accent-em)' }} />
        </div>
      ) : tab === 'stok' ? (
        // ── Combined Stock + Batch (expandable) ─────────────────────────────
        <div className="card" style={{ overflow: 'hidden' }}>
          {/* Legend */}
          <div
            style={{
              padding: '10px 16px',
              borderBottom: '1px solid var(--border)',
              display: 'flex',
              alignItems: 'center',
              gap: 16,
              fontSize: '0.78rem',
              color: 'var(--text-3)',
            }}
          >
            <span style={{ display: 'flex', alignItems: 'center', gap: 5 }}>
              <ChevronRight size={12} />
              Klik baris untuk lihat detail batch FIFO
            </span>
            <span style={{ display: 'flex', alignItems: 'center', gap: 5 }}>
              <Layers size={12} />
              Qty batch = total dari semua batch aktif
            </span>
          </div>

          <table className="tbl">
            <thead>
              <tr>
                <th style={{ width: 36 }} />
                <th>Produk</th>
                <th>SKU</th>
                <th>Stok Saat Ini</th>
                <th>Qty Batch (FIFO)</th>
                <th>Rata-rata HPP</th>
                <th>Min Stok</th>
                <th>Status</th>
              </tr>
            </thead>
            <tbody>
              {levels.map(level => (
                <StockRow
                  key={level.product_id}
                  level={level}
                  summary={summaryMap.get(level.product_id)}
                  batches={batchMap.get(level.product_id) ?? []}
                  isExpanded={expanded.has(level.product_id)}
                  onToggle={() => toggleExpand(level.product_id)}
                />
              ))}
            </tbody>
          </table>
        </div>
      ) : (
        // ── Riwayat Mutasi ──────────────────────────────────────────────────
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
                      className={`badge ${
                        m.ref_type === 'sale' || m.ref_type === 'void_deduct'
                          ? 'badge-red'
                          : m.ref_type === 'purchase' || m.ref_type === 'void'
                            ? 'badge-green'
                            : 'badge-blue'
                      }`}
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
      )}
    </div>
  );
}
