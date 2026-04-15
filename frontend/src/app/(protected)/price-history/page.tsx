'use client';

import { useEffect, useState, useCallback } from 'react';
import { History, Loader2, TrendingUp, TrendingDown, Search, Filter } from 'lucide-react';
import { useAuth } from '@/lib/auth/AuthContext';
import { priceHistoryApi } from '@/lib/api/store-apis';
import { formatRp } from '@/lib/utils';

interface PriceHistoryRow {
  id: string;
  product_id: string;
  product_name: string;
  changed_by_name: string;
  old_cost: number;
  new_cost: number;
  old_sell: number;
  new_sell: number;
  source: string;
  ref_id?: string;
  notes?: string;
  changed_at: string;
}

// ── Helpers ───────────────────────────────────────────────────────────────────

function DeltaBadge({ old: o, next: n, label }: { old: number; next: number; label: string }) {
  const delta = n - o;
  if (Math.abs(delta) < 0.001) {
    return <span style={{ color: 'var(--text-3)', fontSize: '0.78rem' }}>—</span>;
  }
  const color = delta > 0 ? '#10b981' : '#ef4444';
  const Icon = delta > 0 ? TrendingUp : TrendingDown;
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 4, fontSize: '0.78rem', color }}>
        <Icon size={11} />
        {delta > 0 ? '+' : ''}
        {formatRp(delta)}
      </div>
      <div style={{ fontSize: '0.7rem', color: 'var(--text-3)' }}>
        {label}: {formatRp(n)}
      </div>
    </div>
  );
}

function PricePair({ old: o, next: n }: { old: number; next: number }) {
  const changed = Math.abs(n - o) > 0.001;
  return (
    <div
      style={{
        display: 'flex',
        alignItems: 'center',
        gap: 6,
        fontSize: '0.82rem',
        flexWrap: 'wrap',
      }}
    >
      <span style={{ color: 'var(--text-3)', textDecoration: changed ? 'line-through' : 'none' }}>
        {formatRp(o)}
      </span>
      {changed && (
        <>
          <span style={{ color: 'var(--text-3)' }}>→</span>
          <span style={{ fontWeight: 700, color: n > o ? '#10b981' : '#ef4444' }}>
            {formatRp(n)}
          </span>
        </>
      )}
    </div>
  );
}

function SourceBadge({ source }: { source: string }) {
  const cfg =
    source === 'manual'
      ? { bg: 'rgba(99,102,241,0.15)', color: '#818cf8', label: 'Manual' }
      : { bg: 'rgba(245,158,11,0.15)', color: '#f59e0b', label: 'Pembelian' };
  return (
    <span
      style={{
        padding: '2px 8px',
        borderRadius: 6,
        fontSize: '0.72rem',
        fontWeight: 600,
        background: cfg.bg,
        color: cfg.color,
        whiteSpace: 'nowrap',
      }}
    >
      {cfg.label}
    </span>
  );
}

// ── Main Page ─────────────────────────────────────────────────────────────────

export default function PriceHistoryPage() {
  const { selectedStore } = useAuth();
  const [rows, setRows] = useState<PriceHistoryRow[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(false);
  const [search, setSearch] = useState('');
  const [source, setSource] = useState('');
  const storeId = selectedStore?.store_id;
  const PER_PAGE = 20;

  const load = useCallback(
    async (p = 1) => {
      if (!storeId) return;
      setLoading(true);
      try {
        const res = await priceHistoryApi.listByStore(storeId, {
          source: source || undefined,
          page: p,
          per_page: PER_PAGE,
        });
        setRows(res.data?.data ?? []);
        setTotal(res.data?.meta?.total ?? 0);
        setPage(p);
      } catch (e) {
        console.error(e);
      } finally {
        setLoading(false);
      }
    },
    [storeId, source]
  );

  useEffect(() => {
    load(1);
  }, [load]);

  // Client-side search filter
  const filtered = search.trim()
    ? rows.filter(
        r =>
          r.product_name.toLowerCase().includes(search.toLowerCase()) ||
          r.changed_by_name.toLowerCase().includes(search.toLowerCase())
      )
    : rows;

  const totalPages = Math.ceil(total / PER_PAGE);

  if (!selectedStore)
    return (
      <div style={{ padding: 32 }}>
        <div className="empty-state card" style={{ padding: 40 }}>
          <History size={40} />
          <p>Pilih toko terlebih dahulu</p>
        </div>
      </div>
    );

  return (
    <div className="w-full p-6">
      {/* Header */}
      <div className="reveal-animate" style={{ marginBottom: 20 }}>
        <h1 className="page-title" style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
          <History size={22} style={{ color: 'var(--accent-em)' }} />
          Riwayat Harga
        </h1>
        <p className="page-subtitle">
          {selectedStore.store_name} · {total} perubahan tercatat
        </p>
      </div>

      {/* Filters */}
      <div
        className="reveal-animate"
        style={{
          display: 'flex',
          gap: 10,
          marginBottom: 16,
          flexWrap: 'wrap',
          alignItems: 'center',
          animationDelay: '0.1s',
        }}
      >
        {/* Search */}
        <div style={{ position: 'relative', flex: 1, minWidth: 220 }}>
          <Search
            size={14}
            style={{
              position: 'absolute',
              left: 10,
              top: '50%',
              transform: 'translateY(-50%)',
              color: 'var(--text-3)',
            }}
          />
          <input
            className="input"
            style={{ paddingLeft: 32, width: '100%' }}
            placeholder="Cari produk atau kasir..."
            value={search}
            onChange={e => setSearch(e.target.value)}
          />
        </div>

        {/* Source filter */}
        <div style={{ display: 'flex', gap: 4 }}>
          {[
            { val: '', label: 'Semua' },
            { val: 'manual', label: 'Manual' },
            { val: 'purchase_order', label: 'Pembelian' },
          ].map(({ val, label }) => (
            <button
              key={val}
              onClick={() => {
                setSource(val);
              }}
              style={{
                padding: '7px 14px',
                borderRadius: 8,
                border: `1px solid ${source === val ? 'transparent' : 'var(--border)'}`,
                cursor: 'pointer',
                fontSize: '0.82rem',
                fontWeight: 500,
                background: source === val ? 'var(--accent-in)' : 'var(--bg-card)',
                color: source === val ? '#fff' : 'var(--text-2)',
                transition: 'all 0.12s',
              }}
            >
              {label}
            </button>
          ))}
        </div>

        <button className="btn btn-primary" onClick={() => load(1)}>
          <Filter size={14} /> Tampilkan
        </button>
      </div>

      {/* Table */}
      <div className="card reveal-animate" style={{ padding: 0, animationDelay: '0.2s' }}>
        {loading ? (
          <div style={{ display: 'flex', justifyContent: 'center', padding: 40 }}>
            <Loader2 size={24} className="loading-spin" style={{ color: 'var(--accent-em)' }} />
          </div>
        ) : filtered.length === 0 ? (
          <div className="empty-state" style={{ padding: 48 }}>
            <History size={36} />
            <p>Belum ada riwayat perubahan harga</p>
            <span className="text-3" style={{ fontSize: '0.82rem' }}>
              Perubahan akan tercatat otomatis saat produk diupdate atau PO diterima
            </span>
          </div>
        ) : (
          <div className="tbl-container">
            <table className="tbl">
              <thead>
                <tr>
                  <th>Produk</th>
                  <th>Sumber</th>
                  <th>Harga Beli / HPP</th>
                  <th>Harga Jual / HJ</th>
                  <th>Diubah Oleh</th>
                  <th>Waktu</th>
                  <th>Catatan</th>
                </tr>
              </thead>
              <tbody>
                {filtered.map((row, i) => {
                  const costChanged = Math.abs(row.new_cost - row.old_cost) > 0.001;
                  const sellChanged = Math.abs(row.new_sell - row.old_sell) > 0.001;
                  return (
                    <tr
                      key={row.id}
                      className="reveal-animate"
                      style={{ animationDelay: `${0.25 + i * 0.02}s` }}
                    >
                      {/* Product */}
                      <td>
                        <div style={{ fontWeight: 600, maxWidth: 220 }}>{row.product_name}</div>
                      </td>

                      {/* Source */}
                      <td>
                        <SourceBadge source={row.source} />
                      </td>

                      {/* Cost section */}
                      <td>
                        <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
                          {costChanged ? (
                            <PricePair old={row.old_cost} next={row.new_cost} />
                          ) : (
                            <span style={{ color: 'var(--text-3)', fontSize: '0.82rem' }}>
                              {formatRp(row.old_cost)}
                            </span>
                          )}
                          <DeltaBadge old={row.old_cost} next={row.new_cost} label="HPP baru" />
                        </div>
                      </td>

                      {/* Sell section */}
                      <td>
                        <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
                          {sellChanged ? (
                            <PricePair old={row.old_sell} next={row.new_sell} />
                          ) : (
                            <span style={{ color: 'var(--text-3)', fontSize: '0.82rem' }}>
                              {formatRp(row.old_sell)}
                            </span>
                          )}
                          <DeltaBadge old={row.old_sell} next={row.new_sell} label="HJ baru" />
                        </div>
                      </td>

                      {/* Changed by */}
                      <td>
                        <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
                          <div
                            style={{
                              width: 28,
                              height: 28,
                              borderRadius: '50%',
                              background: 'var(--accent-in)',
                              display: 'flex',
                              alignItems: 'center',
                              justifyContent: 'center',
                              fontSize: '0.7rem',
                              fontWeight: 700,
                              color: '#fff',
                              flexShrink: 0,
                            }}
                          >
                            {row.changed_by_name.charAt(0).toUpperCase()}
                          </div>
                          <span style={{ fontSize: '0.82rem' }}>{row.changed_by_name}</span>
                        </div>
                      </td>

                      {/* Timestamp */}
                      <td
                        style={{ whiteSpace: 'nowrap', color: 'var(--text-2)', fontSize: '0.8rem' }}
                      >
                        <div>
                          {new Date(row.changed_at).toLocaleDateString('id-ID', {
                            day: '2-digit',
                            month: 'short',
                            year: 'numeric',
                          })}
                        </div>
                        <div style={{ color: 'var(--text-3)', fontSize: '0.72rem' }}>
                          {new Date(row.changed_at).toLocaleTimeString('id-ID', {
                            hour: '2-digit',
                            minute: '2-digit',
                          })}
                        </div>
                      </td>

                      {/* Notes */}
                      <td style={{ maxWidth: 200 }}>
                        {row.notes ? (
                          <span
                            style={{
                              fontSize: '0.78rem',
                              color: 'var(--text-2)',
                              fontStyle: 'italic',
                            }}
                          >
                            {row.notes}
                          </span>
                        ) : (
                          <span style={{ color: 'var(--text-3)', fontSize: '0.78rem' }}>—</span>
                        )}
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Pagination */}
      {totalPages > 1 && (
        <div style={{ display: 'flex', justifyContent: 'center', gap: 6, marginTop: 16 }}>
          <button className="btn btn-ghost" onClick={() => load(page - 1)} disabled={page <= 1}>
            ‹ Prev
          </button>
          {Array.from({ length: Math.min(totalPages, 7) }, (_, i) => {
            const p = i + 1;
            return (
              <button
                key={p}
                onClick={() => load(p)}
                style={{
                  padding: '6px 12px',
                  borderRadius: 8,
                  border: 'none',
                  cursor: 'pointer',
                  background: p === page ? 'var(--accent-in)' : 'var(--bg-card)',
                  color: p === page ? '#fff' : 'var(--text-2)',
                  fontWeight: p === page ? 700 : 400,
                  fontSize: '0.85rem',
                }}
              >
                {p}
              </button>
            );
          })}
          <button
            className="btn btn-ghost"
            onClick={() => load(page + 1)}
            disabled={page >= totalPages}
          >
            Next ›
          </button>
        </div>
      )}
    </div>
  );
}
