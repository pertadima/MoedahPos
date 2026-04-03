'use client';

import { useEffect, useState, useCallback } from 'react';
import {
  Tag,
  Plus,
  Pencil,
  Trash2,
  Check,
  X,
  Loader2,
  ChevronRight,
  FolderOpen,
  AlertTriangle,
} from 'lucide-react';
import { useAuth } from '@/lib/auth/AuthContext';
import { usePermission } from '@/hooks/usePermission';
import { categoriesApi } from '@/lib/api/store-apis';
import type { Category } from '@/types';
import { formatDate } from '@/lib/utils';

// ── Types ─────────────────────────────────────────────────────────────────────

interface ModalState {
  open: boolean;
  mode: 'create' | 'edit';
  category?: Category;
}

interface DeleteConfirm {
  open: boolean;
  category?: Category;
}

// ── Page ──────────────────────────────────────────────────────────────────────

export default function CategoriesPage() {
  const { selectedStore } = useAuth();
  const { can } = usePermission();
  const [categories, setCategories] = useState<Category[]>([]);
  const [loading, setLoading] = useState(true);
  const [modal, setModal] = useState<ModalState>({ open: false, mode: 'create' });
  const [deleteConfirm, setDeleteConfirm] = useState<DeleteConfirm>({ open: false });
  const [submitting, setSubmitting] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const [toast, setToast] = useState<{ msg: string; type: 'success' | 'error' } | null>(null);

  const storeId = selectedStore?.store_id ?? '';

  // ── Data fetching ─────────────────────────────────────────────────────────

  const fetchCategories = useCallback(async () => {
    if (!storeId) return;
    setLoading(true);
    try {
      const res = await categoriesApi.list(storeId);
      setCategories((res.data as any).data ?? res.data ?? []);
    } catch {
      showToast('Gagal memuat kategori', 'error');
    } finally {
      setLoading(false);
    }
  }, [storeId]);

  useEffect(() => {
    fetchCategories();
  }, [fetchCategories]);

  // ── Toast helper ──────────────────────────────────────────────────────────

  const showToast = (msg: string, type: 'success' | 'error') => {
    setToast({ msg, type });
    setTimeout(() => setToast(null), 3000);
  };

  // ── Form state ────────────────────────────────────────────────────────────

  const [form, setForm] = useState({ name: '', parent_id: '' });
  const [formError, setFormError] = useState('');

  const openCreate = () => {
    setForm({ name: '', parent_id: '' });
    setFormError('');
    setModal({ open: true, mode: 'create' });
  };

  const openEdit = (cat: Category) => {
    setForm({ name: cat.name, parent_id: cat.parent_id ?? '' });
    setFormError('');
    setModal({ open: true, mode: 'edit', category: cat });
  };

  const closeModal = () => setModal({ open: false, mode: 'create' });

  // ── Submit ────────────────────────────────────────────────────────────────

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!form.name.trim()) {
      setFormError('Nama kategori wajib diisi');
      return;
    }
    setSubmitting(true);
    setFormError('');
    const payload: { name: string; parent_id?: string } = { name: form.name.trim() };
    if (form.parent_id) payload.parent_id = form.parent_id;

    try {
      if (modal.mode === 'create') {
        await categoriesApi.create(storeId, payload);
        showToast('Kategori berhasil ditambahkan ✓', 'success');
      } else {
        await categoriesApi.update(storeId, modal.category!.id, payload);
        showToast('Kategori berhasil diperbarui ✓', 'success');
      }
      closeModal();
      fetchCategories();
    } catch (err: any) {
      setFormError(err?.response?.data?.message ?? 'Terjadi kesalahan');
    } finally {
      setSubmitting(false);
    }
  };

  // ── Delete (soft) ─────────────────────────────────────────────────────────

  const confirmDelete = (cat: Category) => setDeleteConfirm({ open: true, category: cat });
  const cancelDelete = () => setDeleteConfirm({ open: false });

  const handleDelete = async () => {
    if (!deleteConfirm.category) return;
    setDeleting(true);
    try {
      await categoriesApi.softDelete(storeId, deleteConfirm.category.id);
      showToast(`Kategori "${deleteConfirm.category.name}" dihapus`, 'success');
      cancelDelete();
      fetchCategories();
    } catch (err: any) {
      showToast(err?.response?.data?.message ?? 'Gagal menghapus kategori', 'error');
      cancelDelete();
    } finally {
      setDeleting(false);
    }
  };

  // ── Parent options (exclude self on edit) ─────────────────────────────────

  const parentOptions = categories.filter(
    c => modal.mode === 'create' || c.id !== modal.category?.id
  );

  // ── No store ──────────────────────────────────────────────────────────────

  if (!selectedStore) {
    return (
      <div style={{ padding: 32 }}>
        <div className="empty-state card" style={{ padding: 40 }}>
          <Tag size={40} style={{ color: 'var(--text-3)' }} />
          <p style={{ fontWeight: 600, color: 'var(--text-2)' }}>Pilih toko terlebih dahulu</p>
        </div>
      </div>
    );
  }

  // ── Render ────────────────────────────────────────────────────────────────

  return (
    <div style={{ padding: 24, maxWidth: 900, margin: '0 auto' }}>
      {/* Toast */}
      {toast && (
        <div
          style={{
            position: 'fixed',
            top: 20,
            right: 20,
            zIndex: 9999,
            background: toast.type === 'success' ? 'rgba(16,185,129,0.15)' : 'rgba(239,68,68,0.15)',
            border: `1px solid ${toast.type === 'success' ? '#10b981' : '#ef4444'}`,
            color: toast.type === 'success' ? '#10b981' : '#ef4444',
            padding: '12px 20px',
            borderRadius: 10,
            fontWeight: 600,
            fontSize: '0.85rem',
            backdropFilter: 'blur(12px)',
            animation: 'slideIn 0.2s ease',
          }}
        >
          {toast.msg}
        </div>
      )}

      {/* Header */}
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          marginBottom: 24,
        }}
      >
        <div>
          <h1 className="page-title" style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
            <Tag size={22} style={{ color: 'var(--accent-em)' }} />
            Manajemen Kategori
          </h1>
          <p className="page-subtitle">
            {selectedStore.store_name} · {categories.length} kategori
          </p>
        </div>
        {can('products.create') && (
          <button className="btn btn-primary" onClick={openCreate} style={{ gap: 8 }}>
            <Plus size={16} />
            Tambah Kategori
          </button>
        )}
      </div>

      {/* Category grid */}
      {loading ? (
        <div style={{ display: 'flex', justifyContent: 'center', padding: 60 }}>
          <Loader2 size={28} className="loading-spin" style={{ color: 'var(--accent-em)' }} />
        </div>
      ) : categories.length === 0 ? (
        <div className="empty-state card" style={{ padding: 60 }}>
          <FolderOpen size={48} style={{ color: 'var(--text-3)' }} />
          <p style={{ fontWeight: 600, color: 'var(--text-2)' }}>Belum ada kategori</p>
          <p style={{ fontSize: '0.85rem' }}>
            Klik &ldquo;Tambah Kategori&rdquo; untuk membuat kategori pertama.
          </p>
        </div>
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
          {/* Top-level categories */}
          {categories
            .filter(c => !c.parent_id)
            .map(parent => {
              const children = categories.filter(c => c.parent_id === parent.id);
              return (
                <div key={parent.id} className="card" style={{ padding: 0, overflow: 'hidden' }}>
                  {/* Parent row */}
                  <CategoryRow
                    cat={parent}
                    isParent
                    onEdit={openEdit}
                    onDelete={confirmDelete}
                    canEdit={can('products.update')}
                    canDelete={can('products.delete')}
                  />
                  {/* Child rows */}
                  {children.map(child => (
                    <CategoryRow
                      key={child.id}
                      cat={child}
                      isParent={false}
                      onEdit={openEdit}
                      onDelete={confirmDelete}
                      canEdit={can('products.update')}
                      canDelete={can('products.delete')}
                    />
                  ))}
                </div>
              );
            })}

          {/* Orphaned categories (parent was deleted) */}
          {categories
            .filter(c => c.parent_id && !categories.find(p => p.id === c.parent_id))
            .map(cat => (
              <div key={cat.id} className="card" style={{ padding: 0, overflow: 'hidden' }}>
                <CategoryRow
                  cat={cat}
                  isParent={false}
                  onEdit={openEdit}
                  onDelete={confirmDelete}
                  canEdit={can('products.update')}
                  canDelete={can('products.delete')}
                />
              </div>
            ))}
        </div>
      )}

      {/* ── Create / Edit Modal ─────────────────────────────────────────────── */}
      {modal.open && (
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
              maxWidth: 480,
              padding: 28,
              animation: 'slideIn 0.2s ease',
            }}
          >
            {/* Modal header */}
            <div
              style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
                marginBottom: 20,
              }}
            >
              <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                <div
                  style={{
                    width: 36,
                    height: 36,
                    borderRadius: 10,
                    background: 'rgba(16,185,129,0.15)',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                  }}
                >
                  <Tag size={18} style={{ color: 'var(--accent-em)' }} />
                </div>
                <div>
                  <div style={{ fontWeight: 700, fontSize: '1rem' }}>
                    {modal.mode === 'create' ? 'Tambah Kategori' : 'Edit Kategori'}
                  </div>
                  <div style={{ fontSize: '0.75rem', color: 'var(--text-3)' }}>
                    {selectedStore.store_name}
                  </div>
                </div>
              </div>
              <button
                onClick={closeModal}
                className="btn btn-ghost btn-sm"
                style={{ padding: '6px' }}
              >
                <X size={16} />
              </button>
            </div>

            <form onSubmit={handleSubmit}>
              {/* Name */}
              <div style={{ marginBottom: 16 }}>
                <label
                  style={{
                    display: 'block',
                    fontSize: '0.8rem',
                    fontWeight: 600,
                    marginBottom: 6,
                    color: 'var(--text-2)',
                  }}
                >
                  Nama Kategori <span style={{ color: '#ef4444' }}>*</span>
                </label>
                <input
                  className="input"
                  placeholder="cth. Minuman, Makanan..."
                  value={form.name}
                  autoFocus
                  onChange={e => setForm(f => ({ ...f, name: e.target.value }))}
                />
              </div>

              {/* Parent */}
              <div style={{ marginBottom: 20 }}>
                <label
                  style={{
                    display: 'block',
                    fontSize: '0.8rem',
                    fontWeight: 600,
                    marginBottom: 6,
                    color: 'var(--text-2)',
                  }}
                >
                  Kategori Induk{' '}
                  <span style={{ fontSize: '0.72rem', color: 'var(--text-3)', fontWeight: 400 }}>
                    (opsional)
                  </span>
                </label>
                <select
                  className="input"
                  value={form.parent_id}
                  onChange={e => setForm(f => ({ ...f, parent_id: e.target.value }))}
                >
                  <option value="">— Tidak ada (kategori utama) —</option>
                  {parentOptions.map(p => (
                    <option key={p.id} value={p.id}>
                      {p.name}
                    </option>
                  ))}
                </select>
              </div>

              {formError && (
                <div
                  style={{
                    background: 'rgba(239,68,68,0.12)',
                    border: '1px solid rgba(239,68,68,0.3)',
                    borderRadius: 8,
                    padding: '10px 14px',
                    marginBottom: 16,
                    fontSize: '0.82rem',
                    color: '#ef4444',
                  }}
                >
                  {formError}
                </div>
              )}

              <div style={{ display: 'flex', gap: 10, justifyContent: 'flex-end' }}>
                <button type="button" className="btn btn-ghost" onClick={closeModal}>
                  Batal
                </button>
                <button
                  type="submit"
                  className="btn btn-primary"
                  disabled={submitting}
                  style={{ gap: 8 }}
                >
                  {submitting ? (
                    <>
                      <Loader2 size={14} className="loading-spin" /> Menyimpan...
                    </>
                  ) : (
                    <>
                      <Check size={14} /> {modal.mode === 'create' ? 'Simpan' : 'Perbarui'}
                    </>
                  )}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* ── Delete Confirmation Modal ───────────────────────────────────────── */}
      {deleteConfirm.open && deleteConfirm.category && (
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
              maxWidth: 420,
              padding: 28,
              animation: 'slideIn 0.2s ease',
            }}
          >
            <div style={{ textAlign: 'center', marginBottom: 20 }}>
              <div
                style={{
                  width: 52,
                  height: 52,
                  borderRadius: '50%',
                  background: 'rgba(239,68,68,0.15)',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  margin: '0 auto 14px',
                }}
              >
                <AlertTriangle size={24} style={{ color: '#ef4444' }} />
              </div>
              <div style={{ fontWeight: 700, fontSize: '1rem', marginBottom: 8 }}>
                Hapus Kategori?
              </div>
              <div style={{ fontSize: '0.85rem', color: 'var(--text-2)', lineHeight: 1.6 }}>
                Kategori{' '}
                <strong style={{ color: 'var(--text-1)' }}>
                  &ldquo;{deleteConfirm.category.name}&rdquo;
                </strong>{' '}
                akan dihapus secara <em>soft-delete</em> (data tetap tersimpan di database). Produk
                dalam kategori ini tidak akan terhapus.
              </div>
            </div>
            <div style={{ display: 'flex', gap: 10, justifyContent: 'center' }}>
              <button className="btn btn-ghost" onClick={cancelDelete} style={{ flex: 1 }}>
                Batal
              </button>
              <button
                className="btn"
                onClick={handleDelete}
                disabled={deleting}
                style={{
                  flex: 1,
                  gap: 8,
                  background: 'rgba(239,68,68,0.15)',
                  color: '#ef4444',
                  border: '1px solid rgba(239,68,68,0.3)',
                }}
              >
                {deleting ? (
                  <>
                    <Loader2 size={14} className="loading-spin" /> Menghapus...
                  </>
                ) : (
                  <>
                    <Trash2 size={14} /> Hapus
                  </>
                )}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

// ── CategoryRow component ──────────────────────────────────────────────────────

function CategoryRow({
  cat,
  isParent,
  onEdit,
  onDelete,
  canEdit,
  canDelete,
}: {
  cat: Category;
  isParent: boolean;
  onEdit: (c: Category) => void;
  onDelete: (c: Category) => void;
  canEdit?: boolean;
  canDelete?: boolean;
}) {
  return (
    <div
      style={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        padding: isParent ? '14px 20px' : '11px 20px 11px 44px',
        borderBottom: '1px solid var(--border)',
        background: isParent ? 'var(--bg-card)' : 'var(--bg-base)',
        transition: 'background 0.15s ease',
      }}
      className="category-row"
    >
      <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
        {!isParent && <ChevronRight size={13} style={{ color: 'var(--text-3)', flexShrink: 0 }} />}
        <div
          style={{
            width: 32,
            height: 32,
            borderRadius: 8,
            background: isParent
              ? 'linear-gradient(135deg, rgba(16,185,129,0.2), rgba(99,102,241,0.15))'
              : 'rgba(255,255,255,0.05)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            flexShrink: 0,
          }}
        >
          <Tag size={14} style={{ color: isParent ? 'var(--accent-em)' : 'var(--text-3)' }} />
        </div>
        <div>
          <div
            style={{
              fontWeight: isParent ? 600 : 400,
              fontSize: isParent ? '0.9rem' : '0.85rem',
              color: 'var(--text-1)',
            }}
          >
            {cat.name}
          </div>
          {cat.parent_name && (
            <div style={{ fontSize: '0.72rem', color: 'var(--text-3)' }}>
              dalam {cat.parent_name}
            </div>
          )}
        </div>
      </div>

      <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
        <span style={{ fontSize: '0.72rem', color: 'var(--text-3)' }}>
          {formatDate(cat.created_at)}
        </span>
        <div style={{ display: 'flex', gap: 6 }}>
          {canEdit !== false && (
            <button
              className="btn btn-ghost btn-sm"
              onClick={() => onEdit(cat)}
              title="Edit kategori"
              style={{ padding: '5px 8px' }}
            >
              <Pencil size={14} />
            </button>
          )}
          {canDelete !== false && (
            <button
              className="btn btn-ghost btn-sm"
              onClick={() => onDelete(cat)}
              title="Hapus kategori"
              style={{ padding: '5px 8px', color: 'rgba(239,68,68,0.7)' }}
            >
              <Trash2 size={14} />
            </button>
          )}
        </div>
      </div>
    </div>
  );
}
