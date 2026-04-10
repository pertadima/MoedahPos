'use client';

import { useEffect, useState, useCallback, useMemo } from 'react';
import {
  ArrowDownToLine,
  Plus,
  Loader2,
  X,
  Edit2,
  Trash2,
  TrendingUp,
  Banknote,
  CreditCard,
  Smartphone,
} from 'lucide-react';
import { useAuth } from '@/lib/auth/AuthContext';
import { incomesApi } from '@/lib/api/store-apis';
import { formatRp, formatDate, parseNumberInput, formatNumberInput } from '@/lib/utils';
import type { Income, IncomeCategory } from '@/types';

const PAYMENT_METHODS = [
  { value: 'cash', label: 'Tunai' },
  { value: 'transfer', label: 'Transfer Bank' },
  { value: 'qris', label: 'QRIS' },
  { value: 'other', label: 'Lainnya' },
] as const;

type PaymentMethod = (typeof PAYMENT_METHODS)[number]['value'];

function methodIcon(m: string) {
  if (m === 'qris') return <Smartphone size={14} />;
  if (m === 'transfer') return <CreditCard size={14} />;
  if (m === 'cash') return <Banknote size={14} />;
  return <CreditCard size={14} />;
}

function methodLabel(m: string) {
  return PAYMENT_METHODS.find(p => p.value === m)?.label ?? m;
}

function methodBadgeStyle(m: string): React.CSSProperties {
  const colors: Record<string, { bg: string; color: string }> = {
    cash: { bg: 'rgba(16,185,129,0.12)', color: '#10b981' },
    transfer: { bg: 'rgba(99,102,241,0.12)', color: '#6366f1' },
    qris: { bg: 'rgba(245,158,11,0.12)', color: '#f59e0b' },
    other: { bg: 'rgba(107,114,128,0.12)', color: '#6b7280' },
  };
  const c = colors[m] ?? colors.other;
  return {
    display: 'inline-flex',
    alignItems: 'center',
    gap: 4,
    padding: '3px 8px',
    borderRadius: 6,
    fontSize: '0.72rem',
    fontWeight: 600,
    background: c.bg,
    color: c.color,
  };
}

export default function IncomesPage() {
  const { selectedStore } = useAuth();

  const [incomes, setIncomes] = useState<Income[]>([]);
  const [categories, setCategories] = useState<IncomeCategory[]>([]);
  const [loading, setLoading] = useState(true);
  const [showModal, setShowModal] = useState(false);
  const [editTarget, setEditTarget] = useState<Income | null>(null);

  const [filter, setFilter] = useState({
    category_id: '',
    date_from: '',
    date_to: '',
  });

  const [meta, setMeta] = useState({ page: 1, per_page: 20, total: 0, total_pages: 0 });

  // ── Data Fetching ─────────────────────────────────────────────────────────

  const loadCategories = useCallback(async () => {
    try {
      const res = await incomesApi.listCategories();
      setCategories(res.data || []);
    } catch (err) {
      console.error(err);
    }
  }, []);

  const loadIncomes = useCallback(async () => {
    if (!selectedStore) return;
    setLoading(true);
    try {
      const res = await incomesApi.list(selectedStore.store_id, {
        ...filter,
        page: meta.page,
        per_page: meta.per_page,
      });
      setIncomes(res.data.data ?? []);
      setMeta(res.data.meta);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  }, [selectedStore, filter, meta.page, meta.per_page]);

  useEffect(() => {
    loadCategories();
  }, [loadCategories]);

  useEffect(() => {
    loadIncomes();
  }, [loadIncomes]);

  // ── Totals ────────────────────────────────────────────────────────────────

  const periodTotal = useMemo(() => incomes.reduce((s, i) => s + i.amount, 0), [incomes]);

  const byMethodTotals = useMemo(() => {
    const map: Record<string, number> = {};
    incomes.forEach(i => {
      map[i.payment_method] = (map[i.payment_method] ?? 0) + i.amount;
    });
    return Object.entries(map).sort((a, b) => b[1] - a[1]);
  }, [incomes]);

  // ── Actions ───────────────────────────────────────────────────────────────

  const handleDelete = async (id: string) => {
    if (!selectedStore) return;
    if (!confirm('Hapus catatan pemasukan ini?')) return;
    try {
      await incomesApi.delete(selectedStore.store_id, id);
      loadIncomes();
    } catch (err) {
      alert('Gagal menghapus pemasukan');
      console.error(err);
    }
  };

  const handleFilterChange = (key: string, value: string) => {
    setFilter(prev => ({ ...prev, [key]: value }));
    setMeta(prev => ({ ...prev, page: 1 }));
  };

  // ── UI ────────────────────────────────────────────────────────────────────

  if (!selectedStore) {
    return (
      <div style={{ padding: 32 }}>
        <div
          style={{
            textAlign: 'center',
            padding: '60px 20px',
            background: 'var(--bg-card)',
            borderRadius: 12,
            border: '1px dashed var(--border)',
          }}
        >
          <ArrowDownToLine size={48} style={{ color: 'var(--text-4)', margin: '0 auto 16px' }} />
          <p style={{ color: 'var(--text-3)' }}>Pilih toko terlebih dahulu</p>
        </div>
      </div>
    );
  }

  return (
    <div style={{ padding: '24px 32px' }}>
      {/* Header */}
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          marginBottom: 24,
        }}
      >
        <div>
          <h1
            style={{
              fontSize: '1.4rem',
              fontWeight: 800,
              margin: '0 0 4px',
              color: 'var(--text-1)',
              display: 'flex',
              alignItems: 'center',
              gap: 8,
            }}
          >
            <ArrowDownToLine size={22} style={{ color: '#10b981' }} />
            Pemasukan
          </h1>
          <p style={{ margin: 0, fontSize: '0.85rem', color: 'var(--text-3)' }}>
            Catat pendapatan di luar penjualan kasir — modal, refund supplier, dan lainnya.
          </p>
        </div>
        <button
          onClick={() => {
            setEditTarget(null);
            setShowModal(true);
          }}
          className="btn btn-primary"
          style={{ display: 'flex', alignItems: 'center', gap: 6 }}
        >
          <Plus size={16} /> Tambah Pemasukan
        </button>
      </div>

      {/* Summary cards */}
      {incomes.length > 0 && (
        <div
          style={{
            display: 'flex',
            gap: 12,
            marginBottom: 20,
            flexWrap: 'wrap',
          }}
        >
          <div
            className="card"
            style={{
              padding: '14px 20px',
              borderTop: '3px solid #10b981',
              flex: '1 1 180px',
              minWidth: 180,
            }}
          >
            <div
              style={{
                fontSize: '0.68rem',
                fontWeight: 700,
                textTransform: 'uppercase',
                color: 'var(--text-3)',
                letterSpacing: '0.07em',
                marginBottom: 4,
              }}
            >
              Total Periode Ini
            </div>
            <div style={{ fontSize: '1.4rem', fontWeight: 800, color: '#10b981' }}>
              {formatRp(periodTotal)}
            </div>
            <div style={{ fontSize: '0.72rem', color: 'var(--text-3)', marginTop: 2 }}>
              {incomes.length} catatan
            </div>
          </div>
          {byMethodTotals.map(([method, total]) => (
            <div
              key={method}
              className="card"
              style={{ padding: '14px 20px', flex: '1 1 160px', minWidth: 150 }}
            >
              <div style={{ display: 'flex', alignItems: 'center', gap: 6, marginBottom: 4 }}>
                <span style={{ color: '#10b981' }}>{methodIcon(method)}</span>
                <span
                  style={{
                    fontSize: '0.68rem',
                    fontWeight: 700,
                    textTransform: 'uppercase',
                    color: 'var(--text-3)',
                    letterSpacing: '0.07em',
                  }}
                >
                  {methodLabel(method)}
                </span>
              </div>
              <div style={{ fontSize: '1.1rem', fontWeight: 700 }}>{formatRp(total)}</div>
            </div>
          ))}
        </div>
      )}

      {/* Filters */}
      <div
        style={{
          background: 'var(--bg-card)',
          padding: 16,
          borderRadius: 12,
          border: '1px solid var(--border)',
          marginBottom: 20,
          display: 'flex',
          gap: 12,
          alignItems: 'flex-end',
          flexWrap: 'wrap',
        }}
      >
        <div style={{ flex: '1 1 180px' }}>
          <label className="label">Kategori</label>
          <select
            className="input"
            style={{ width: '100%', height: 38 }}
            value={filter.category_id}
            onChange={e => handleFilterChange('category_id', e.target.value)}
          >
            <option value="">Semua Kategori</option>
            {categories.map(c => (
              <option key={c.id} value={c.id}>
                {c.name}
              </option>
            ))}
          </select>
        </div>
        <div style={{ flex: '1 1 150px' }}>
          <label className="label">Dari Tanggal</label>
          <input
            type="date"
            className="input"
            style={{ width: '100%', height: 38 }}
            value={filter.date_from}
            onChange={e => handleFilterChange('date_from', e.target.value)}
          />
        </div>
        <div style={{ flex: '1 1 150px' }}>
          <label className="label">Sampai Tanggal</label>
          <input
            type="date"
            className="input"
            style={{ width: '100%', height: 38 }}
            value={filter.date_to}
            onChange={e => handleFilterChange('date_to', e.target.value)}
          />
        </div>
        <button
          onClick={() => {
            setFilter({ category_id: '', date_from: '', date_to: '' });
            setMeta(prev => ({ ...prev, page: 1 }));
          }}
          className="btn btn-secondary"
          style={{ height: 38 }}
        >
          Reset
        </button>
      </div>

      {/* Table */}
      {loading && incomes.length === 0 ? (
        <div style={{ textAlign: 'center', padding: 40, color: 'var(--text-3)' }}>
          <Loader2 size={32} className="loading-spin" style={{ margin: '0 auto 12px' }} />
          Memuat data pemasukan...
        </div>
      ) : incomes.length === 0 ? (
        <div
          style={{
            textAlign: 'center',
            padding: '60px 20px',
            background: 'var(--bg-card)',
            borderRadius: 12,
            border: '1px dashed var(--border)',
          }}
        >
          <TrendingUp size={48} style={{ color: 'var(--text-4)', margin: '0 auto 16px' }} />
          <h3 style={{ margin: '0 0 8px', fontSize: '1.1rem', color: 'var(--text-1)' }}>
            Belum Ada Pemasukan
          </h3>
          <p style={{ margin: 0, color: 'var(--text-3)', fontSize: '0.9rem' }}>
            {filter.category_id || filter.date_from
              ? 'Tidak ada data yang cocok dengan filter.'
              : 'Mulai catat injeksi modal, refund supplier, atau pendapatan lain-lain.'}
          </p>
        </div>
      ) : (
        <div
          style={{
            background: 'var(--bg-card)',
            borderRadius: 12,
            border: '1px solid var(--border)',
            overflow: 'hidden',
          }}
        >
          <div style={{ overflowX: 'auto' }}>
            <table className="tbl" style={{ width: '100%', minWidth: 700 }}>
              <thead>
                <tr>
                  <th>Tanggal</th>
                  <th>Kategori</th>
                  <th>Metode</th>
                  <th>Referensi</th>
                  <th>Catatan</th>
                  <th style={{ textAlign: 'right' }}>Jumlah</th>
                  <th style={{ textAlign: 'right', width: 90 }}>Aksi</th>
                </tr>
              </thead>
              <tbody>
                {incomes.map(inc => (
                  <tr key={inc.id}>
                    <td style={{ fontWeight: 600 }}>{formatDate(inc.income_date)}</td>
                    <td>
                      <span
                        style={{
                          background: 'rgba(16,185,129,0.1)',
                          color: '#10b981',
                          padding: '4px 8px',
                          borderRadius: 6,
                          fontSize: '0.75rem',
                          fontWeight: 600,
                        }}
                      >
                        {inc.category_name}
                      </span>
                    </td>
                    <td>
                      <span style={methodBadgeStyle(inc.payment_method)}>
                        {methodIcon(inc.payment_method)}
                        {methodLabel(inc.payment_method)}
                      </span>
                    </td>
                    <td style={{ color: 'var(--text-3)', fontSize: '0.82rem' }}>
                      {inc.reference || '—'}
                    </td>
                    <td style={{ color: 'var(--text-2)', maxWidth: 200 }}>
                      <span
                        style={{
                          display: 'block',
                          overflow: 'hidden',
                          textOverflow: 'ellipsis',
                          whiteSpace: 'nowrap',
                          maxWidth: 200,
                        }}
                        title={inc.notes}
                      >
                        {inc.notes || '—'}
                      </span>
                    </td>
                    <td style={{ textAlign: 'right', fontWeight: 700, color: '#10b981' }}>
                      {formatRp(inc.amount)}
                    </td>
                    <td style={{ textAlign: 'right' }}>
                      <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 6 }}>
                        <button
                          onClick={() => {
                            setEditTarget(inc);
                            setShowModal(true);
                          }}
                          style={{
                            background: 'none',
                            border: 'none',
                            cursor: 'pointer',
                            color: 'var(--text-3)',
                          }}
                          title="Edit"
                        >
                          <Edit2 size={15} />
                        </button>
                        <button
                          onClick={() => handleDelete(inc.id)}
                          style={{
                            background: 'none',
                            border: 'none',
                            cursor: 'pointer',
                            color: '#ef4444',
                          }}
                          title="Hapus"
                        >
                          <Trash2 size={15} />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {/* Pagination */}
          {meta.total_pages > 1 && (
            <div
              style={{
                padding: 16,
                borderTop: '1px solid var(--border)',
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
              }}
            >
              <div style={{ fontSize: '0.85rem', color: 'var(--text-3)' }}>
                Hal {meta.page} dari {meta.total_pages} ({meta.total} data)
              </div>
              <div style={{ display: 'flex', gap: 8 }}>
                <button
                  className="btn btn-secondary btn-sm"
                  disabled={meta.page <= 1}
                  onClick={() => setMeta(prev => ({ ...prev, page: prev.page - 1 }))}
                >
                  Prev
                </button>
                <button
                  className="btn btn-secondary btn-sm"
                  disabled={meta.page >= meta.total_pages}
                  onClick={() => setMeta(prev => ({ ...prev, page: prev.page + 1 }))}
                >
                  Next
                </button>
              </div>
            </div>
          )}
        </div>
      )}

      {/* Modal */}
      {showModal && (
        <IncomeModal
          categories={categories}
          income={editTarget}
          onClose={() => setShowModal(false)}
          onSuccess={() => {
            setShowModal(false);
            loadIncomes();
          }}
        />
      )}
    </div>
  );
}

// ── Modal ──────────────────────────────────────────────────────────────────────

function IncomeModal({
  categories,
  income,
  onClose,
  onSuccess,
}: {
  categories: IncomeCategory[];
  income: Income | null;
  onClose: () => void;
  onSuccess: () => void;
}) {
  const { selectedStore } = useAuth();
  const [saving, setSaving] = useState(false);
  const [err, setErr] = useState('');

  const [form, setForm] = useState({
    category_id: income?.category_id ?? categories[0]?.id ?? '',
    amount: income ? formatNumberInput(income.amount) : '',
    income_date: income?.income_date ?? new Date().toISOString().slice(0, 10),
    payment_method: (income?.payment_method ?? 'cash') as PaymentMethod,
    reference: income?.reference ?? '',
    notes: income?.notes ?? '',
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedStore) return;

    const amt = parseNumberInput(form.amount);
    if (!amt || amt <= 0) {
      setErr('Jumlah pemasukan tidak valid');
      return;
    }
    if (!form.category_id) {
      setErr('Pilih kategori');
      return;
    }

    setSaving(true);
    setErr('');
    try {
      const payload = { ...form, amount: amt };
      if (income) {
        await incomesApi.update(selectedStore.store_id, income.id, payload);
      } else {
        await incomesApi.create(selectedStore.store_id, payload);
      }
      onSuccess();
    } catch (error: unknown) {
      const msg = error instanceof Error ? error.message : 'Gagal menyimpan pemasukan';
      setErr(msg);
    } finally {
      setSaving(false);
    }
  };

  return (
    <div
      style={{
        position: 'fixed',
        inset: 0,
        zIndex: 200,
        background: 'rgba(0,0,0,0.55)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
      }}
      onClick={e => e.target === e.currentTarget && onClose()}
    >
      <div
        style={{
          background: 'var(--bg-card)',
          padding: 28,
          borderRadius: 14,
          width: 440,
          maxWidth: '94vw',
          boxShadow: '0 20px 60px rgba(0,0,0,0.3)',
        }}
      >
        {/* Modal header */}
        <div
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            marginBottom: 22,
          }}
        >
          <div>
            <h2 style={{ margin: 0, fontSize: '1.15rem', fontWeight: 800 }}>
              {income ? 'Ubah Pemasukan' : 'Tambah Pemasukan'}
            </h2>
            <p style={{ margin: '2px 0 0', fontSize: '0.8rem', color: 'var(--text-3)' }}>
              Pendapatan di luar penjualan kasir
            </p>
          </div>
          <button
            onClick={onClose}
            style={{
              background: 'none',
              border: 'none',
              cursor: 'pointer',
              color: 'var(--text-3)',
            }}
          >
            <X size={20} />
          </button>
        </div>

        {err && (
          <div
            style={{
              background: 'rgba(239,68,68,0.1)',
              border: '1px solid rgba(239,68,68,0.2)',
              color: '#ef4444',
              fontSize: '0.85rem',
              padding: '10px 14px',
              borderRadius: 8,
              marginBottom: 16,
            }}
          >
            {err}
          </div>
        )}

        <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
          {/* Category */}
          <div>
            <label className="label">Kategori</label>
            <select
              className="input"
              style={{ width: '100%' }}
              value={form.category_id}
              onChange={e => setForm({ ...form, category_id: e.target.value })}
              required
            >
              {categories.map(c => (
                <option key={c.id} value={c.id}>
                  {c.name}
                </option>
              ))}
            </select>
          </div>

          {/* Amount */}
          <div>
            <label className="label">Jumlah (Rp)</label>
            <input
              type="text"
              className="input"
              style={{ width: '100%', fontSize: '1.2rem', fontWeight: 700, color: '#10b981' }}
              value={form.amount}
              onChange={e =>
                setForm({ ...form, amount: formatNumberInput(parseNumberInput(e.target.value)) })
              }
              placeholder="0"
              required
            />
          </div>

          {/* Date */}
          <div>
            <label className="label">Tanggal</label>
            <input
              type="date"
              className="input"
              style={{ width: '100%' }}
              value={form.income_date}
              onChange={e => setForm({ ...form, income_date: e.target.value })}
              required
            />
          </div>

          {/* Payment method */}
          <div>
            <label className="label">Metode Pembayaran</label>
            <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
              {PAYMENT_METHODS.map(m => (
                <button
                  key={m.value}
                  type="button"
                  onClick={() => setForm({ ...form, payment_method: m.value })}
                  style={{
                    padding: '7px 14px',
                    borderRadius: 8,
                    fontSize: '0.82rem',
                    fontWeight: 600,
                    border:
                      form.payment_method === m.value
                        ? '2px solid #10b981'
                        : '2px solid var(--border)',
                    background:
                      form.payment_method === m.value
                        ? 'rgba(16,185,129,0.1)'
                        : 'var(--bg-elevated)',
                    color: form.payment_method === m.value ? '#10b981' : 'var(--text-2)',
                    cursor: 'pointer',
                    display: 'flex',
                    alignItems: 'center',
                    gap: 5,
                  }}
                >
                  {methodIcon(m.value)}
                  {m.label}
                </button>
              ))}
            </div>
          </div>

          {/* Reference */}
          <div>
            <label className="label">Referensi / No. Invoice (Opsional)</label>
            <input
              type="text"
              className="input"
              style={{ width: '100%' }}
              value={form.reference}
              onChange={e => setForm({ ...form, reference: e.target.value })}
              placeholder="Contoh: INV-001, TF-20240410"
            />
          </div>

          {/* Notes */}
          <div>
            <label className="label">Catatan (Opsional)</label>
            <textarea
              className="input"
              style={{ width: '100%', height: 72, resize: 'vertical' }}
              value={form.notes}
              onChange={e => setForm({ ...form, notes: e.target.value })}
              placeholder="Detail pemasukan..."
            />
          </div>

          <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 10, marginTop: 4 }}>
            <button type="button" className="btn btn-secondary" onClick={onClose}>
              Batal
            </button>
            <button type="submit" className="btn btn-primary" disabled={saving}>
              {saving ? (
                <Loader2 size={16} className="loading-spin" />
              ) : income ? (
                'Simpan Perubahan'
              ) : (
                'Tambah'
              )}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
