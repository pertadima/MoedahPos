'use client';

import { useCallback, useEffect, useState, useMemo } from 'react';
import {
  AlertTriangle,
  ChevronDown,
  ChevronRight,
  Layers,
  Loader2,
  TrendingDown,
  TrendingUp,
  Warehouse,
  Search,
  Check,
  Plus,
  Package,
} from 'lucide-react';
import { useAuth } from '@/lib/auth/AuthContext';
import { stockApi } from '@/lib/api/store-apis';
import { api } from '@/lib/api/client';
import { stockAdjustmentApi, type CreateAdjustmentInput } from '@/lib/api/stock-adjustments';
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

interface Product {
  id: string;
  name: string;
  sku: string;
  unit: string;
}

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
  const [currentPage, setCurrentPage] = useState(1);
  const [searchQuery, setSearchQuery] = useState('');

  useEffect(() => {
    setCurrentPage(1);
  }, [searchQuery]);

  const storeId = selectedStore?.store_id;
  const role = selectedStore?.role;
  const canUpdateStock = ['superadmin', 'admin', 'manager'].includes(role || '');

  // ── Modal state ─────────────────────────────────────────────────────────────
  const [products, setProducts] = useState<Product[]>([]);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [formData, setFormData] = useState<CreateAdjustmentInput>({
    product_id: '',
    type: 'OUT',
    reason: 'DAMAGED',
    quantity: 1,
    notes: '',
  });

  const [productSearch, setProductSearch] = useState('');
  const [debouncedSearch, setDebouncedSearch] = useState('');
  const [showDropdown, setShowDropdown] = useState(false);
  const [isSearchingProducts, setIsSearchingProducts] = useState(false);

  useEffect(() => {
    const handler = setTimeout(() => {
      setDebouncedSearch(productSearch);
    }, 300);
    return () => clearTimeout(handler);
  }, [productSearch]);

  useEffect(() => {
    if (!storeId) return;
    const fetchSearchedProducts = async () => {
      try {
        setIsSearchingProducts(true);
        const params = new URLSearchParams({ per_page: '20' });
        if (debouncedSearch && !formData.product_id) {
          params.append('search', debouncedSearch);
        }
        const res = await api.get<{ data: Product[] }>(
          `/stores/${storeId}/products?${params.toString()}`
        );
        const prodData = Array.isArray(res?.data) ? res.data : res?.data?.data || [];
        setProducts(prodData);
      } catch (err) {
        console.error(err);
      } finally {
        setIsSearchingProducts(false);
      }
    };
    fetchSearchedProducts();
  }, [storeId, debouncedSearch, formData.product_id]);

  const handleAdjustmentSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!storeId) return;
    try {
      setSubmitting(true);
      await stockAdjustmentApi.create(storeId, formData);
      setIsModalOpen(false);
      setFormData({
        product_id: '',
        type: 'OUT',
        reason: 'DAMAGED',
        quantity: 1,
        notes: '',
      });
      setProductSearch('');
      setShowDropdown(false);
      loadAll();
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Gagal menyimpan penyesuaian';
      alert(msg);
    } finally {
      setSubmitting(false);
    }
  };

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

  // ── Pagination Logic ────────────────────────────────────────────────────────
  const filteredLevels = useMemo(() => {
    if (!searchQuery) return levels;
    const lowerQ = searchQuery.toLowerCase();
    return levels.filter(
      l =>
        l.product_name.toLowerCase().includes(lowerQ) ||
        l.product_sku.toLowerCase().includes(lowerQ)
    );
  }, [levels, searchQuery]);

  const itemsPerPage = 20;
  const totalItems = filteredLevels.length;
  const totalPages = Math.max(1, Math.ceil(totalItems / itemsPerPage));
  const startIndex = (currentPage - 1) * itemsPerPage;
  const paginatedLevels = filteredLevels.slice(startIndex, startIndex + itemsPerPage);

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
    <div className="w-full p-6">
      {/* Header */}
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          marginBottom: 20,
        }}
      >
        <div>
          <h1 className="page-title" style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
            <Package size={22} style={{ color: 'var(--accent-em)' }} />
            Manajemen Stok
          </h1>
          <p className="page-subtitle">
            {selectedStore.store_name} · {levels.length} produk
            {lowCount > 0 ? ` · ⚠ ${lowCount} stok menipis` : ''}
          </p>
        </div>
        {canUpdateStock && (
          <button className="btn btn-primary btn-sm" onClick={() => setIsModalOpen(true)}>
            <Plus size={16} /> Buat Penyesuaian
          </button>
        )}
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
          {/* Search & Legend */}
          <div
            style={{
              padding: '10px 16px',
              borderBottom: '1px solid var(--border)',
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
              gap: 16,
              flexWrap: 'wrap',
            }}
          >
            <div style={{ position: 'relative', width: 280, maxWidth: '100%' }}>
              <Search
                size={15}
                style={{
                  position: 'absolute',
                  left: 12,
                  top: '50%',
                  transform: 'translateY(-50%)',
                  color: 'var(--text-3)',
                }}
              />
              <input
                type="text"
                placeholder="Cari produk atau SKU..."
                value={searchQuery}
                onChange={e => setSearchQuery(e.target.value)}
                style={{
                  width: '100%',
                  padding: '8px 12px 8px 36px',
                  borderRadius: 6,
                  border: '1px solid var(--border)',
                  background: 'var(--bg-surface)',
                  color: 'var(--text-1)',
                  fontSize: '0.85rem',
                }}
              />
            </div>
            <div style={{ display: 'flex', gap: 16, fontSize: '0.78rem', color: 'var(--text-3)' }}>
              <span style={{ display: 'flex', alignItems: 'center', gap: 5 }}>
                <ChevronRight size={12} />
                Klik baris untuk lihat detail batch FIFO
              </span>
              <span style={{ display: 'flex', alignItems: 'center', gap: 5 }}>
                <Layers size={12} />
                Qty batch = total dari semua batch aktif
              </span>
            </div>
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
              {paginatedLevels.map(level => (
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

          {/* Pagination Controls */}
          {totalPages > 1 && (
            <div
              style={{
                padding: '12px 16px',
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
                borderTop: '1px solid var(--border)',
                background: 'var(--bg-card)',
              }}
            >
              <div style={{ fontSize: '0.85rem', color: 'var(--text-3)' }}>
                Menampilkan {startIndex + 1}-{Math.min(startIndex + itemsPerPage, totalItems)} dari{' '}
                {totalItems} produk
              </div>
              <div style={{ display: 'flex', gap: 8 }}>
                <button
                  className="btn btn-ghost btn-sm"
                  disabled={currentPage === 1}
                  onClick={() => setCurrentPage(p => Math.max(1, p - 1))}
                >
                  Sebelumnya
                </button>
                <div
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    margin: '0 8px',
                    fontSize: '0.85rem',
                    fontWeight: 500,
                  }}
                >
                  {currentPage} / {totalPages}
                </div>
                <button
                  className="btn btn-ghost btn-sm"
                  disabled={currentPage === totalPages}
                  onClick={() => setCurrentPage(p => Math.min(totalPages, p + 1))}
                >
                  Selanjutnya
                </button>
              </div>
            </div>
          )}
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

      {/* ── Create Modal ─────────────────────────────────────────────────────── */}
      {isModalOpen && (
        <div
          style={{
            position: 'fixed',
            inset: 0,
            zIndex: 1000,
            background: 'rgba(0,0,0,0.6)',
            backdropFilter: 'blur(4px)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            padding: 16,
          }}
        >
          <div
            className="card"
            style={{
              width: '100%',
              maxWidth: 500,
              padding: 28,
              animation: 'slideIn 0.2s ease',
            }}
          >
            <h3 style={{ margin: '0 0 16px', color: 'var(--text-1)' }}>Catat Penyesuaian Stok</h3>

            <form
              onSubmit={handleAdjustmentSubmit}
              style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}
            >
              {/* Searchable Product Dropdown */}
              <div>
                <label
                  style={{
                    display: 'block',
                    marginBottom: '6px',
                    fontSize: '0.8rem',
                    fontWeight: 600,
                  }}
                >
                  Produk <span style={{ color: 'var(--danger)' }}>*</span>
                </label>
                <div
                  style={{ position: 'relative' }}
                  onBlur={e => {
                    // Cek jika elemen yang di-klik berada di luar div ini
                    if (!e.currentTarget.contains(e.relatedTarget)) {
                      setShowDropdown(false);
                      // Jika ada product_id tapi search dikosongkan, kembalikan ke nama produk terpilih
                      if (formData.product_id) {
                        const selectedProduct = products.find(p => p.id === formData.product_id);
                        if (selectedProduct)
                          setProductSearch(`${selectedProduct.name} (${selectedProduct.sku})`);
                      }
                    }
                  }}
                >
                  <Search
                    size={16}
                    style={{ position: 'absolute', left: 12, top: 11, color: 'var(--text-3)' }}
                  />
                  <input
                    type="text"
                    className="input"
                    placeholder="Cari nama komponen atau SKU..."
                    style={{ paddingLeft: 36 }}
                    value={
                      productSearch ||
                      (formData.product_id && !showDropdown
                        ? `${products.find(p => p.id === formData.product_id)?.name} (${products.find(p => p.id === formData.product_id)?.sku})`
                        : productSearch)
                    }
                    onChange={e => {
                      setProductSearch(e.target.value);
                      setShowDropdown(true);
                      if (formData.product_id) setFormData({ ...formData, product_id: '' });
                    }}
                    onFocus={() => setShowDropdown(true)}
                  />
                  {showDropdown && (
                    <div
                      style={{
                        position: 'absolute',
                        top: '100%',
                        left: 0,
                        right: 0,
                        zIndex: 50,
                        background: 'var(--bg-elevated)',
                        border: '1px solid var(--border-md)',
                        borderRadius: 8,
                        maxHeight: 220,
                        overflowY: 'auto',
                        marginTop: 4,
                        boxShadow:
                          '0 10px 15px -3px rgba(0,0,0,0.1), 0 4px 6px -4px rgba(0,0,0,0.1)',
                      }}
                    >
                      {products.map(p => (
                        <button
                          key={p.id}
                          type="button"
                          style={{
                            width: '100%',
                            textAlign: 'left',
                            padding: '10px 14px',
                            background:
                              formData.product_id === p.id ? 'var(--bg-surface)' : 'transparent',
                            border: 'none',
                            borderBottom: '1px solid var(--border-light)',
                            cursor: 'pointer',
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'space-between',
                            transition: 'background 0.2s',
                          }}
                          onClick={() => {
                            setFormData({ ...formData, product_id: p.id });
                            setProductSearch(`${p.name} (${p.sku})`);
                            setShowDropdown(false);
                          }}
                          onMouseOver={e =>
                            (e.currentTarget.style.background = 'var(--bg-surface)')
                          }
                          onMouseOut={e =>
                            (e.currentTarget.style.background =
                              formData.product_id === p.id ? 'var(--bg-surface)' : 'transparent')
                          }
                        >
                          <div>
                            <div
                              style={{
                                fontWeight: 600,
                                fontSize: '0.85rem',
                                color: 'var(--text-1)',
                              }}
                            >
                              {p.name}
                            </div>
                            <div style={{ fontSize: '0.75rem', color: 'var(--text-3)' }}>
                              SKU: {p.sku} | Unit: {p.unit}
                            </div>
                          </div>
                          {formData.product_id === p.id && (
                            <Check size={16} style={{ color: 'var(--accent-em)' }} />
                          )}
                        </button>
                      ))}
                      {isSearchingProducts && (
                        <div
                          style={{
                            padding: '16px',
                            fontSize: '0.8rem',
                            color: 'var(--text-3)',
                            textAlign: 'center',
                          }}
                        >
                          Mencari produk...
                        </div>
                      )}
                      {!isSearchingProducts && products.length === 0 && (
                        <div
                          style={{
                            padding: '16px',
                            fontSize: '0.8rem',
                            color: 'var(--text-3)',
                            textAlign: 'center',
                          }}
                        >
                          Tidak ada produk yang cocok dengan pencarian.
                        </div>
                      )}
                    </div>
                  )}
                </div>
              </div>

              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '16px' }}>
                <div>
                  <label
                    style={{
                      display: 'block',
                      marginBottom: '6px',
                      fontSize: '0.8rem',
                      fontWeight: 600,
                    }}
                  >
                    Tipe Permintaan
                  </label>
                  <select
                    className="input"
                    value={formData.type}
                    onChange={e => {
                      const newType = e.target.value as 'IN' | 'OUT';
                      setFormData({
                        ...formData,
                        type: newType,
                        reason: newType === 'IN' ? 'MANUAL_CORRECTION' : 'DAMAGED',
                      });
                    }}
                  >
                    <option value="OUT">Keluar (OUT)</option>
                    <option value="IN">Masuk (IN)</option>
                  </select>
                </div>

                <div>
                  <label
                    style={{
                      display: 'block',
                      marginBottom: '6px',
                      fontSize: '0.8rem',
                      fontWeight: 600,
                    }}
                  >
                    Alasan
                  </label>
                  <select
                    className="input"
                    value={formData.reason}
                    onChange={e =>
                      setFormData({
                        ...formData,
                        reason: e.target.value as CreateAdjustmentInput['reason'],
                      })
                    }
                  >
                    {formData.type === 'OUT' && (
                      <>
                        <option value="DAMAGED">Barang Rusak</option>
                        <option value="LOST">Barang Hilang</option>
                      </>
                    )}
                    <option value="MANUAL_CORRECTION">Koreksi Manual</option>
                  </select>
                </div>
              </div>

              <div>
                <label
                  style={{
                    display: 'block',
                    marginBottom: '6px',
                    fontSize: '0.8rem',
                    fontWeight: 600,
                  }}
                >
                  Kuantitas <span style={{ color: 'var(--danger)' }}>*</span>
                </label>
                <input
                  type="number"
                  min="0.1"
                  step="0.1"
                  className="input"
                  required
                  value={formData.quantity}
                  onChange={e => setFormData({ ...formData, quantity: parseFloat(e.target.value) })}
                />
              </div>

              <div>
                <label
                  style={{
                    display: 'block',
                    marginBottom: '6px',
                    fontSize: '0.8rem',
                    fontWeight: 600,
                  }}
                >
                  Catatan{' '}
                  {formData.reason === 'MANUAL_CORRECTION' && (
                    <span style={{ color: 'var(--danger)' }}>*</span>
                  )}
                </label>
                <textarea
                  className="input"
                  rows={3}
                  required={formData.reason === 'MANUAL_CORRECTION'}
                  value={formData.notes}
                  placeholder={
                    formData.reason === 'MANUAL_CORRECTION' ? 'Wajib isi alasan detail' : 'Opsional'
                  }
                  onChange={e => setFormData({ ...formData, notes: e.target.value })}
                />
              </div>

              <div
                style={{
                  display: 'flex',
                  justifyContent: 'flex-end',
                  gap: '8px',
                  marginTop: '8px',
                }}
              >
                <button
                  type="button"
                  className="btn btn-ghost"
                  onClick={() => setIsModalOpen(false)}
                  disabled={submitting}
                >
                  Batal
                </button>
                <button
                  type="submit"
                  className="btn btn-primary"
                  disabled={submitting || !formData.product_id}
                >
                  {submitting ? 'Menyimpan...' : 'Simpan Penyesuaian'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
