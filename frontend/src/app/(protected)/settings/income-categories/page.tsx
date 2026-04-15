'use client';

import { useState, useEffect, useCallback } from 'react';
import { Tag, Plus, Loader2, Edit3, Archive, Play, X } from 'lucide-react';
import { incomesApi } from '@/lib/api/store-apis';
import { formatDate } from '@/lib/utils';
import { ApiError } from '@/lib/api/client';
import type { Category } from '@/types';

export interface FullCategory extends Category {
  description: string;
  is_active: boolean;
  updated_at: string;
}

// ── Form Modal ─────────────────────────────────────────────────────────────

interface FormModalProps {
  initial?: FullCategory | null;
  onSuccess: () => void;
  onClose: () => void;
}

function FormModal({ initial, onSuccess, onClose }: FormModalProps) {
  const isEdit = !!initial;
  const [form, setForm] = useState({
    name: initial?.name ?? '',
    description: initial?.description ?? '',
    is_active: initial?.is_active ?? true,
  });
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');

  const set =
    (k: keyof typeof form) =>
    (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) =>
      setForm(f => ({
        ...f,
        [k]: e.target.type === 'checkbox' ? (e.target as HTMLInputElement).checked : e.target.value,
      }));

  const handleSave = async () => {
    if (!form.name.trim()) {
      setError('Nama kategori wajib diisi');
      return;
    }
    setSaving(true);
    setError('');
    try {
      if (isEdit && initial) {
        await incomesApi.updateCategory(initial.id, form);
      } else {
        await incomesApi.createCategory(form);
      }
      onSuccess();
    } catch (e) {
      setError(e instanceof ApiError ? e.message : 'Gagal menyimpan kategori');
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-box" style={{ maxWidth: 460 }} onClick={e => e.stopPropagation()}>
        <div
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            marginBottom: 18,
          }}
        >
          <h2 style={{ fontWeight: 800 }}>
            {isEdit ? 'Edit Kategori Pemasukan' : 'Tambah Kategori Pemasukan'}
          </h2>
          <button className="btn btn-ghost btn-sm" onClick={onClose}>
            <X size={15} />
          </button>
        </div>

        {error && (
          <div
            style={{
              background: 'rgba(239,68,68,0.1)',
              border: '1px solid rgba(239,68,68,0.3)',
              borderRadius: 8,
              padding: '8px 12px',
              color: '#f87171',
              fontSize: '0.83rem',
              marginBottom: 14,
            }}
          >
            {error}
          </div>
        )}

        <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
          <div className="input-group">
            <label className="input-label">
              Nama Kategori <span style={{ color: '#ef4444' }}>*</span>
            </label>
            <input
              className="input"
              value={form.name}
              onChange={set('name')}
              placeholder="Contoh: Investasi, Penjualan Aset"
              autoFocus
            />
          </div>
          <div className="input-group">
            <label className="input-label">Deskripsi</label>
            <textarea
              className="input"
              value={form.description}
              onChange={set('description')}
              rows={2}
              placeholder="Keterangan kategori"
              style={{ resize: 'vertical' }}
            />
          </div>
          {isEdit && (
            <div
              className="input-group"
              style={{ flexDirection: 'row', alignItems: 'center', gap: 10 }}
            >
              <input
                type="checkbox"
                checked={form.is_active}
                onChange={set('is_active')}
                id="is_active"
              />
              <label htmlFor="is_active" style={{ cursor: 'pointer', margin: 0, fontWeight: 600 }}>
                Kategori Aktif
              </label>
            </div>
          )}
        </div>

        <div style={{ display: 'flex', gap: 8, marginTop: 18 }}>
          <button className="btn btn-secondary" style={{ flex: 1 }} onClick={onClose}>
            Batal
          </button>
          <button
            className="btn btn-primary"
            style={{ flex: 1 }}
            onClick={handleSave}
            disabled={saving}
          >
            {saving ? <Loader2 size={14} className="loading-spin" /> : null}
            {saving ? 'Menyimpan...' : 'Simpan'}
          </button>
        </div>
      </div>
    </div>
  );
}

// ── Delete Confirm ────────────────────────────────────────────────────────────
function DeleteConfirm({
  category,
  onSuccess,
  onClose,
}: {
  category: FullCategory;
  onSuccess: () => void;
  onClose: () => void;
}) {
  const [loading, setLoading] = useState(false);
  const handleDelete = async () => {
    setLoading(true);
    try {
      await incomesApi.deleteCategory(category.id);
      onSuccess();
    } catch (e) {
      alert(e instanceof ApiError ? e.message : 'Gagal menonaktifkan kategori');
    } finally {
      setLoading(false);
    }
  };
  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-box" style={{ maxWidth: 400 }} onClick={e => e.stopPropagation()}>
        <div style={{ textAlign: 'center', padding: '8px 0 16px' }}>
          <Archive size={28} style={{ color: '#f59e0b', marginBottom: 12 }} />
          <h2 style={{ fontWeight: 800, marginBottom: 8 }}>Nonaktifkan Kategori?</h2>
          <p style={{ color: 'var(--text-2)', fontSize: '0.875rem', lineHeight: 1.6 }}>
            <strong>{category.name}</strong> akan dinonaktifkan dan tidak bisa dipilih pada laporan
            baru, tetapi riwayat historis tetap aman.
          </p>
        </div>
        <div style={{ display: 'flex', gap: 8 }}>
          <button
            className="btn btn-secondary"
            style={{ flex: 1 }}
            onClick={onClose}
            disabled={loading}
          >
            Batal
          </button>
          <button
            style={{
              flex: 1,
              background: '#f59e0b',
              color: '#fff',
              border: 'none',
              borderRadius: 8,
              padding: '8px 0',
              fontWeight: 700,
              cursor: 'pointer',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              gap: 6,
            }}
            onClick={handleDelete}
            disabled={loading}
          >
            {loading ? <Loader2 size={14} className="loading-spin" /> : <Archive size={14} />}
            {loading ? 'Memproses...' : 'Ya, Nonaktifkan'}
          </button>
        </div>
      </div>
    </div>
  );
}

// ── Main Page ─────────────────────────────────────────────────────────────────

export default function IncomeCategoriesPage() {
  const [categories, setCategories] = useState<FullCategory[]>([]);
  const [loading, setLoading] = useState(true);
  const [form, setForm] = useState<'create' | FullCategory | null>(null);
  const [deleting, setDeleting] = useState<FullCategory | null>(null);

  const load = useCallback(() => {
    setLoading(true);
    incomesApi
      .listCategories({ include_deleted: true })
      .then(r => setCategories(r.data))
      .catch(console.error)
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    load();
  }, [load]);

  const onSuccess = () => {
    setForm(null);
    setDeleting(null);
    load();
  };

  const handleToggleActive = async (category: FullCategory) => {
    try {
      await incomesApi.updateCategory(category.id, {
        name: category.name,
        description: category.description,
        is_active: !category.is_active,
      });
      load();
    } catch (e) {
      alert(e instanceof ApiError ? e.message : 'Gagal mengubah status kategori');
    }
  };

  return (
    <div className="w-full p-6 max-w-[1400px] mx-auto">
      {/* Header */}
      <div
        className="reveal-animate"
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'flex-start',
          marginBottom: 20,
        }}
      >
        <div>
          <h1 className="page-title" style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
            <Tag size={22} style={{ color: 'var(--accent-em)' }} /> Kategori Pemasukan
          </h1>
          <p className="page-subtitle">Kelola kategori global untuk pemasukan seluruh toko</p>
        </div>
        <button className="btn btn-primary" onClick={() => setForm('create')}>
          <Plus size={15} /> Tambah Kategori
        </button>
      </div>

      {/* Table */}
      <div className="card reveal-animate" style={{ padding: 0, animationDelay: '0.1s' }}>
        {loading ? (
          <div style={{ display: 'flex', justifyContent: 'center', padding: 40 }}>
            <Loader2 size={24} className="loading-spin" style={{ color: 'var(--accent-em)' }} />
          </div>
        ) : categories.length === 0 ? (
          <div className="empty-state">
            <Tag size={36} />
            <p>Belum ada kategori pemasukan</p>
          </div>
        ) : (
          <div className="tbl-container">
            <table className="tbl">
              <thead>
                <tr>
                  <th>Nama Kategori</th>
                  <th>Deskripsi</th>
                  <th>Status</th>
                  <th>Ditambahkan</th>
                  <th style={{ textAlign: 'right' }}>Aksi</th>
                </tr>
              </thead>
              <tbody>
                {categories.map((c, i) => (
                  <tr
                    key={c.id}
                    className="reveal-animate"
                    style={{
                      opacity: c.is_active ? 1 : 0.6,
                      animationDelay: `${0.15 + i * 0.02}s`,
                    }}
                  >
                    <td style={{ fontWeight: 600 }}>{c.name}</td>
                    <td style={{ color: 'var(--text-2)', fontSize: '0.85rem' }}>
                      {c.description || '—'}
                    </td>
                    <td>
                      {c.is_active ? (
                        <span
                          style={{
                            padding: '4px 8px',
                            borderRadius: 6,
                            fontSize: '0.72rem',
                            fontWeight: 700,
                            background: 'rgba(16,185,129,0.1)',
                            color: '#10b981',
                          }}
                        >
                          Aktif
                        </span>
                      ) : (
                        <span
                          style={{
                            padding: '4px 8px',
                            borderRadius: 6,
                            fontSize: '0.72rem',
                            fontWeight: 700,
                            background: 'rgba(239,68,68,0.1)',
                            color: '#ef4444',
                          }}
                        >
                          Tidak Aktif
                        </span>
                      )}
                    </td>
                    <td style={{ color: 'var(--text-3)', fontSize: '0.8rem' }}>
                      {formatDate(c.created_at)}
                    </td>
                    <td>
                      <div style={{ display: 'flex', gap: 4, justifyContent: 'flex-end' }}>
                        <button
                          className="btn btn-ghost btn-sm"
                          onClick={() => setForm(c)}
                          title="Edit"
                        >
                          <Edit3 size={13} />
                        </button>
                        {c.is_active ? (
                          <button
                            className="btn btn-ghost btn-sm"
                            style={{ color: '#f59e0b' }}
                            onClick={() => setDeleting(c)}
                            title="Nonaktifkan"
                          >
                            <Archive size={13} />
                          </button>
                        ) : (
                          <button
                            className="btn btn-ghost btn-sm"
                            style={{ color: '#10b981' }}
                            onClick={() => handleToggleActive(c)}
                            title="Aktifkan Kembali"
                          >
                            <Play size={13} />
                          </button>
                        )}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {form === 'create' && <FormModal onSuccess={onSuccess} onClose={() => setForm(null)} />}
      {form && form !== 'create' && (
        <FormModal
          initial={form as FullCategory}
          onSuccess={onSuccess}
          onClose={() => setForm(null)}
        />
      )}
      {deleting && (
        <DeleteConfirm
          category={deleting}
          onSuccess={onSuccess}
          onClose={() => setDeleting(null)}
        />
      )}
    </div>
  );
}
