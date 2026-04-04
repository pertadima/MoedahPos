'use client';

import { useEffect, useState, useCallback } from 'react';
import {
  Store,
  Plus,
  Pencil,
  Trash2,
  Check,
  X,
  Loader2,
  AlertTriangle,
  Search,
  UtensilsCrossed,
  ShoppingBag,
  MapPin,
  Phone,
  Hash,
  ToggleLeft,
  ToggleRight,
} from 'lucide-react';
import { usePermission } from '@/hooks/usePermission';
import { storesApi } from '@/lib/api/store-apis';
import type { Store as StoreType } from '@/types';
import { formatDate } from '@/lib/utils';

// ── Form defaults ─────────────────────────────────────────────────────────────

interface StoreForm {
  name: string;
  address: string;
  phone: string;
  tax_number: string;
  currency: string;
  store_type: 'retail' | 'restaurant';
}

const emptyForm = (): StoreForm => ({
  name: '',
  address: '',
  phone: '',
  tax_number: '',
  currency: 'IDR',
  store_type: 'retail',
});

const STORE_TYPE_CONFIG = {
  retail: { label: 'Retail', icon: ShoppingBag, color: '#10b981', bg: 'rgba(16,185,129,0.12)' },
  restaurant: {
    label: 'Restaurant',
    icon: UtensilsCrossed,
    color: '#fb923c',
    bg: 'rgba(251,146,60,0.12)',
  },
};

// ── Page ──────────────────────────────────────────────────────────────────────

export default function StoresPage() {
  const { can } = usePermission();
  const [stores, setStores] = useState<StoreType[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState('');
  const [page, setPage] = useState(1);
  const perPage = 12;

  const [modal, setModal] = useState<{ open: boolean; mode: 'create' | 'edit'; store?: StoreType }>(
    { open: false, mode: 'create' }
  );
  const [deleteConfirm, setDeleteConfirm] = useState<{ open: boolean; store?: StoreType }>({
    open: false,
  });
  const [form, setForm] = useState<StoreForm>(emptyForm());
  const [formError, setFormError] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [toast, setToast] = useState<{ msg: string; type: 'success' | 'error' } | null>(null);

  // ── Fetch ─────────────────────────────────────────────────────────────────

  const fetchStores = useCallback(async () => {
    setLoading(true);
    try {
      const res = await storesApi.list({ page, per_page: perPage, search: search || undefined });
      // Response: { success: true, data: Store[], meta: { total, page, ... } }
      const body = res.data as any;
      setStores(body?.data ?? []);
      setTotal(body?.meta?.total ?? 0);
    } catch {
      showToast('Gagal memuat data toko', 'error');
    } finally {
      setLoading(false);
    }
  }, [page, search]);

  useEffect(() => {
    fetchStores();
  }, [fetchStores]);

  // Debounce search → reset page
  useEffect(() => {
    const t = setTimeout(() => setPage(1), 400);
    return () => clearTimeout(t);
  }, [search]);

  // ── Toast ─────────────────────────────────────────────────────────────────

  const showToast = (msg: string, type: 'success' | 'error') => {
    setToast({ msg, type });
    setTimeout(() => setToast(null), 3500);
  };

  // ── Modal helpers ─────────────────────────────────────────────────────────

  const openCreate = () => {
    setForm(emptyForm());
    setFormError('');
    setModal({ open: true, mode: 'create' });
  };

  const openEdit = (s: StoreType) => {
    setForm({
      name: s.name,
      address: s.address,
      phone: s.phone,
      tax_number: s.tax_number,
      currency: s.currency,
      store_type: (s.store_type as 'retail' | 'restaurant') || 'retail',
    });
    setFormError('');
    setModal({ open: true, mode: 'edit', store: s });
  };

  const closeModal = () => setModal({ open: false, mode: 'create' });

  // ── Submit ────────────────────────────────────────────────────────────────

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!form.name.trim()) {
      setFormError('Nama toko wajib diisi');
      return;
    }
    setSubmitting(true);
    setFormError('');
    const payload = {
      name: form.name.trim(),
      address: form.address,
      phone: form.phone,
      tax_number: form.tax_number,
      currency: form.currency || 'IDR',
      store_type: form.store_type,
    };
    try {
      if (modal.mode === 'create') {
        await storesApi.create(payload);
        showToast('Toko berhasil ditambahkan ✓', 'success');
      } else {
        await storesApi.update(modal.store!.id, { ...payload, is_active: modal.store!.is_active });
        showToast('Toko berhasil diperbarui ✓', 'success');
      }
      closeModal();
      fetchStores();
    } catch (err: any) {
      setFormError(err?.response?.data?.message ?? 'Terjadi kesalahan');
    } finally {
      setSubmitting(false);
    }
  };

  // ── Toggle active ─────────────────────────────────────────────────────────

  const toggleActive = async (s: StoreType) => {
    try {
      await storesApi.update(s.id, {
        name: s.name,
        address: s.address,
        phone: s.phone,
        tax_number: s.tax_number,
        currency: s.currency,
        store_type: s.store_type ?? 'retail',
        is_active: !s.is_active,
      });
      showToast(`${s.name} ${!s.is_active ? 'diaktifkan' : 'dinonaktifkan'}`, 'success');
      fetchStores();
    } catch {
      showToast('Gagal mengubah status', 'error');
    }
  };

  // ── Soft Delete ───────────────────────────────────────────────────────────

  const handleDelete = async () => {
    if (!deleteConfirm.store) return;
    try {
      await storesApi.softDelete(deleteConfirm.store.id);
      showToast(`Toko "${deleteConfirm.store.name}" dihapus`, 'success');
      setDeleteConfirm({ open: false });
      fetchStores();
    } catch (err: any) {
      showToast(err?.response?.data?.message ?? 'Gagal menghapus', 'error');
      setDeleteConfirm({ open: false });
    }
  };

  const totalPages = Math.ceil(total / perPage);

  // ── Render ────────────────────────────────────────────────────────────────

  return (
    <div style={{ padding: 24, maxWidth: '100%', margin: '0 auto' }}>
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
            <Store size={22} style={{ color: 'var(--accent-em)' }} />
            Manajemen Toko
          </h1>
          <p className="page-subtitle">{total} toko terdaftar</p>
        </div>
        {can('stores.create') && (
          <button className="btn btn-primary" onClick={openCreate} style={{ gap: 8 }}>
            <Plus size={16} /> Tambah Toko
          </button>
        )}
      </div>

      {/* Search bar */}
      <div style={{ position: 'relative', marginBottom: 20, maxWidth: 380 }}>
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
          className="input"
          placeholder="Cari nama toko…"
          value={search}
          onChange={e => setSearch(e.target.value)}
          style={{ paddingLeft: 36 }}
        />
      </div>

      {/* Grid */}
      {loading ? (
        <div style={{ display: 'flex', justifyContent: 'center', padding: 60 }}>
          <Loader2 size={28} className="loading-spin" style={{ color: 'var(--accent-em)' }} />
        </div>
      ) : stores.length === 0 ? (
        <div className="empty-state card" style={{ padding: 60 }}>
          <Store size={48} style={{ color: 'var(--text-3)' }} />
          <p style={{ fontWeight: 600, color: 'var(--text-2)' }}>
            {search ? 'Tidak ditemukan' : 'Belum ada toko'}
          </p>
          <p style={{ fontSize: '0.85rem' }}>
            {search
              ? `Tidak ada toko dengan nama "${search}"`
              : 'Klik &ldquo;Tambah Toko" untuk membuat toko pertama.'}
          </p>
        </div>
      ) : (
        <div
          style={{
            display: 'grid',
            gridTemplateColumns: 'repeat(auto-fill, minmax(300px, 1fr))',
            gap: 16,
          }}
        >
          {stores.map(store => {
            const typeCfg =
              STORE_TYPE_CONFIG[store.store_type as 'retail' | 'restaurant'] ??
              STORE_TYPE_CONFIG.retail;
            const TypeIcon = typeCfg.icon;
            return (
              <div
                key={store.id}
                className="card"
                style={{ padding: 0, overflow: 'hidden', opacity: store.is_active ? 1 : 0.65 }}
              >
                {/* Top accent strip */}
                <div
                  style={{
                    height: 3,
                    background: store.is_active ? typeCfg.color : 'var(--border)',
                  }}
                />

                <div style={{ padding: '16px 18px 14px' }}>
                  {/* Header row */}
                  <div
                    style={{
                      display: 'flex',
                      alignItems: 'flex-start',
                      justifyContent: 'space-between',
                      marginBottom: 10,
                    }}
                  >
                    <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                      <div
                        style={{
                          width: 38,
                          height: 38,
                          borderRadius: 10,
                          flexShrink: 0,
                          background: typeCfg.bg,
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'center',
                        }}
                      >
                        <TypeIcon size={18} style={{ color: typeCfg.color }} />
                      </div>
                      <div>
                        <div
                          style={{
                            fontWeight: 700,
                            fontSize: '0.92rem',
                            color: 'var(--text-1)',
                            lineHeight: 1.2,
                          }}
                        >
                          {store.name}
                        </div>
                        <div
                          style={{
                            display: 'inline-flex',
                            alignItems: 'center',
                            gap: 3,
                            marginTop: 3,
                            fontSize: '0.67rem',
                            padding: '1px 6px',
                            borderRadius: 5,
                            fontWeight: 600,
                            background: typeCfg.bg,
                            color: typeCfg.color,
                          }}
                        >
                          <TypeIcon size={9} />
                          {typeCfg.label}
                        </div>
                      </div>
                    </div>

                    {/* Status badge — display only */}
                    <div
                      style={{
                        fontSize: '0.68rem',
                        padding: '2px 8px',
                        borderRadius: 6,
                        fontWeight: 600,
                        background: store.is_active
                          ? 'rgba(16,185,129,0.12)'
                          : 'rgba(239,68,68,0.10)',
                        color: store.is_active ? '#10b981' : '#ef4444',
                        border: `1px solid ${store.is_active ? 'rgba(16,185,129,0.25)' : 'rgba(239,68,68,0.2)'}`,
                      }}
                    >
                      {store.is_active ? 'Aktif' : 'Nonaktif'}
                    </div>
                  </div>

                  {/* Details */}
                  <div
                    style={{ display: 'flex', flexDirection: 'column', gap: 5, marginBottom: 12 }}
                  >
                    {store.address && (
                      <div
                        style={{
                          display: 'flex',
                          gap: 6,
                          alignItems: 'flex-start',
                          fontSize: '0.78rem',
                          color: 'var(--text-2)',
                        }}
                      >
                        <MapPin
                          size={12}
                          style={{ color: 'var(--text-3)', marginTop: 1, flexShrink: 0 }}
                        />
                        <span style={{ lineHeight: 1.4 }}>{store.address}</span>
                      </div>
                    )}
                    {store.phone && (
                      <div
                        style={{
                          display: 'flex',
                          gap: 6,
                          alignItems: 'center',
                          fontSize: '0.78rem',
                          color: 'var(--text-2)',
                        }}
                      >
                        <Phone size={12} style={{ color: 'var(--text-3)', flexShrink: 0 }} />
                        {store.phone}
                      </div>
                    )}
                    {store.tax_number && (
                      <div
                        style={{
                          display: 'flex',
                          gap: 6,
                          alignItems: 'center',
                          fontSize: '0.78rem',
                          color: 'var(--text-2)',
                        }}
                      >
                        <Hash size={12} style={{ color: 'var(--text-3)', flexShrink: 0 }} />
                        NPWP: {store.tax_number}
                      </div>
                    )}
                  </div>

                  {/* Footer */}
                  <div
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'space-between',
                      borderTop: '1px solid var(--border)',
                      paddingTop: 10,
                    }}
                  >
                    <span style={{ fontSize: '0.7rem', color: 'var(--text-3)' }}>
                      {store.currency} · dibuat {formatDate(store.created_at)}
                    </span>
                    <div style={{ display: 'flex', gap: 6 }}>
                      {can('stores.update') && (
                        <button
                          className="btn btn-ghost btn-sm"
                          onClick={() => toggleActive(store)}
                          title={store.is_active ? 'Nonaktifkan toko' : 'Aktifkan toko'}
                          style={{
                            padding: '4px 9px',
                            fontSize: '0.72rem',
                            gap: 4,
                            color: store.is_active ? '#f59e0b' : '#10b981',
                          }}
                        >
                          {store.is_active ? (
                            <>
                              <ToggleLeft size={13} /> Nonaktifkan
                            </>
                          ) : (
                            <>
                              <ToggleRight size={13} /> Aktifkan
                            </>
                          )}
                        </button>
                      )}
                      {can('stores.update') && (
                        <button
                          className="btn btn-ghost btn-sm"
                          onClick={() => openEdit(store)}
                          title="Edit toko"
                          style={{ padding: '4px 9px', fontSize: '0.72rem', gap: 4 }}
                        >
                          <Pencil size={13} /> Edit
                        </button>
                      )}
                      {can('stores.delete') && (
                        <button
                          className="btn btn-ghost btn-sm"
                          onClick={() => setDeleteConfirm({ open: true, store })}
                          title="Hapus toko"
                          style={{
                            padding: '4px 9px',
                            fontSize: '0.72rem',
                            gap: 4,
                            color: 'rgba(239,68,68,0.8)',
                          }}
                        >
                          <Trash2 size={13} /> Hapus
                        </button>
                      )}
                    </div>
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      )}

      {/* Pagination */}
      {totalPages > 1 && (
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            gap: 10,
            marginTop: 28,
          }}
        >
          <button
            className="btn btn-ghost btn-sm"
            disabled={page === 1}
            onClick={() => setPage(p => p - 1)}
          >
            ← Prev
          </button>
          <span style={{ fontSize: '0.82rem', color: 'var(--text-2)' }}>
            Halaman {page} dari {totalPages}
          </span>
          <button
            className="btn btn-ghost btn-sm"
            disabled={page === totalPages}
            onClick={() => setPage(p => p + 1)}
          >
            Next →
          </button>
        </div>
      )}

      {/* ── Create / Edit Modal ─────────────────────────────────────────────── */}
      {modal.open && (
        <div
          style={{
            position: 'fixed',
            inset: 0,
            zIndex: 1000,
            background: 'rgba(0,0,0,0.65)',
            backdropFilter: 'blur(4px)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            padding: 16,
          }}
        >
          <div
            className="card"
            style={{ width: '100%', maxWidth: 520, padding: 28, animation: 'slideIn 0.2s ease' }}
          >
            {/* Modal header */}
            <div
              style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
                marginBottom: 22,
              }}
            >
              <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                <div
                  style={{
                    width: 36,
                    height: 36,
                    borderRadius: 10,
                    background: 'rgba(16,185,129,0.12)',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                  }}
                >
                  <Store size={18} style={{ color: 'var(--accent-em)' }} />
                </div>
                <div>
                  <div style={{ fontWeight: 700, fontSize: '1rem' }}>
                    {modal.mode === 'create' ? 'Tambah Toko Baru' : `Edit: ${modal.store?.name}`}
                  </div>
                  <div style={{ fontSize: '0.72rem', color: 'var(--text-3)' }}>
                    {modal.mode === 'create'
                      ? 'Isi informasi toko baru'
                      : 'Perbarui informasi toko'}
                  </div>
                </div>
              </div>
              <button onClick={closeModal} className="btn btn-ghost btn-sm" style={{ padding: 6 }}>
                <X size={16} />
              </button>
            </div>

            <form onSubmit={handleSubmit}>
              {/* Store Type selector */}
              <div style={{ marginBottom: 18 }}>
                <label
                  style={{
                    display: 'block',
                    fontSize: '0.8rem',
                    fontWeight: 600,
                    marginBottom: 8,
                    color: 'var(--text-2)',
                  }}
                >
                  Tipe Toko <span style={{ color: '#ef4444' }}>*</span>
                </label>
                <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 10 }}>
                  {(['retail', 'restaurant'] as const).map(t => {
                    const cfg = STORE_TYPE_CONFIG[t];
                    const Icon = cfg.icon;
                    const active = form.store_type === t;
                    return (
                      <button
                        key={t}
                        type="button"
                        onClick={() => setForm(f => ({ ...f, store_type: t }))}
                        style={{
                          display: 'flex',
                          alignItems: 'center',
                          gap: 10,
                          padding: '12px 14px',
                          borderRadius: 10,
                          cursor: 'pointer',
                          background: active ? cfg.bg : 'var(--bg-elevated)',
                          border: active ? `1.5px solid ${cfg.color}` : '1.5px solid var(--border)',
                          color: active ? cfg.color : 'var(--text-2)',
                          transition: 'all 0.15s ease',
                        }}
                      >
                        <Icon size={18} />
                        <div style={{ textAlign: 'left' }}>
                          <div style={{ fontWeight: 600, fontSize: '0.85rem' }}>{cfg.label}</div>
                          <div style={{ fontSize: '0.68rem', opacity: 0.7 }}>
                            {t === 'retail' ? 'Minimarket, toko, dll.' : 'Restoran, kafe, dll.'}
                          </div>
                        </div>
                        {active && (
                          <Check size={14} style={{ marginLeft: 'auto', flexShrink: 0 }} />
                        )}
                      </button>
                    );
                  })}
                </div>
              </div>

              {/* Name */}
              <div style={{ marginBottom: 14 }}>
                <label
                  style={{
                    display: 'block',
                    fontSize: '0.8rem',
                    fontWeight: 600,
                    marginBottom: 5,
                    color: 'var(--text-2)',
                  }}
                >
                  Nama Toko <span style={{ color: '#ef4444' }}>*</span>
                </label>
                <input
                  className="input"
                  placeholder="cth. Warung Kopi Makmur"
                  autoFocus
                  value={form.name}
                  onChange={e => setForm(f => ({ ...f, name: e.target.value }))}
                />
              </div>

              {/* Address */}
              <div style={{ marginBottom: 14 }}>
                <label
                  style={{
                    display: 'block',
                    fontSize: '0.8rem',
                    fontWeight: 600,
                    marginBottom: 5,
                    color: 'var(--text-2)',
                  }}
                >
                  Alamat
                </label>
                <textarea
                  className="input"
                  rows={2}
                  placeholder="Jl. Contoh No. 1, Jakarta"
                  value={form.address}
                  onChange={e => setForm(f => ({ ...f, address: e.target.value }))}
                  style={{ resize: 'vertical', minHeight: 52 }}
                />
              </div>

              {/* Phone & Currency */}
              <div
                style={{
                  display: 'grid',
                  gridTemplateColumns: '1fr 1fr',
                  gap: 12,
                  marginBottom: 14,
                }}
              >
                <div>
                  <label
                    style={{
                      display: 'block',
                      fontSize: '0.8rem',
                      fontWeight: 600,
                      marginBottom: 5,
                      color: 'var(--text-2)',
                    }}
                  >
                    No. Telepon
                  </label>
                  <input
                    className="input"
                    placeholder="021-12345678"
                    value={form.phone}
                    onChange={e => setForm(f => ({ ...f, phone: e.target.value }))}
                  />
                </div>
                <div>
                  <label
                    style={{
                      display: 'block',
                      fontSize: '0.8rem',
                      fontWeight: 600,
                      marginBottom: 5,
                      color: 'var(--text-2)',
                    }}
                  >
                    Mata Uang
                  </label>
                  <select
                    className="input"
                    value={form.currency}
                    onChange={e => setForm(f => ({ ...f, currency: e.target.value }))}
                  >
                    <option value="IDR">IDR — Rupiah</option>
                    <option value="USD">USD — US Dollar</option>
                    <option value="SGD">SGD — Singapore Dollar</option>
                    <option value="MYR">MYR — Ringgit</option>
                  </select>
                </div>
              </div>

              {/* NPWP */}
              <div style={{ marginBottom: 20 }}>
                <label
                  style={{
                    display: 'block',
                    fontSize: '0.8rem',
                    fontWeight: 600,
                    marginBottom: 5,
                    color: 'var(--text-2)',
                  }}
                >
                  NPWP{' '}
                  <span style={{ fontSize: '0.72rem', color: 'var(--text-3)', fontWeight: 400 }}>
                    (opsional)
                  </span>
                </label>
                <input
                  className="input"
                  placeholder="00.000.000.0-000.000"
                  value={form.tax_number}
                  onChange={e => setForm(f => ({ ...f, tax_number: e.target.value }))}
                />
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
                      <Check size={14} />{' '}
                      {modal.mode === 'create' ? 'Buat Toko' : 'Simpan Perubahan'}
                    </>
                  )}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* ── Delete Confirmation ─────────────────────────────────────────────── */}
      {deleteConfirm.open && deleteConfirm.store && (
        <div
          style={{
            position: 'fixed',
            inset: 0,
            zIndex: 1000,
            background: 'rgba(0,0,0,0.65)',
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
              maxWidth: 400,
              padding: 28,
              textAlign: 'center',
              animation: 'slideIn 0.2s ease',
            }}
          >
            <div
              style={{
                width: 52,
                height: 52,
                borderRadius: '50%',
                background: 'rgba(239,68,68,0.15)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                margin: '0 auto 16px',
              }}
            >
              <AlertTriangle size={24} style={{ color: '#ef4444' }} />
            </div>
            <div style={{ fontWeight: 700, fontSize: '1rem', marginBottom: 8 }}>Hapus Toko?</div>
            <div
              style={{
                fontSize: '0.85rem',
                color: 'var(--text-2)',
                lineHeight: 1.7,
                marginBottom: 6,
              }}
            >
              Toko{' '}
              <strong style={{ color: 'var(--text-1)' }}>
                &ldquo;{deleteConfirm.store.name}&rdquo;
              </strong>{' '}
              akan dihapus secara <em>soft-delete</em>.
            </div>
            <div
              style={{
                fontSize: '0.78rem',
                color: 'var(--text-3)',
                background: 'rgba(245,158,11,0.08)',
                border: '1px solid rgba(245,158,11,0.2)',
                borderRadius: 8,
                padding: '8px 12px',
                marginBottom: 20,
                textAlign: 'left',
              }}
            >
              ⚠️ Data produk, transaksi, dan anggota toko ini tidak akan terhapus — hanya toko yang
              tersembunyi.
            </div>
            <div style={{ display: 'flex', gap: 10 }}>
              <button
                className="btn btn-ghost"
                onClick={() => setDeleteConfirm({ open: false })}
                style={{ flex: 1 }}
              >
                Batal
              </button>
              <button
                className="btn"
                onClick={handleDelete}
                style={{
                  flex: 1,
                  gap: 8,
                  background: 'rgba(239,68,68,0.15)',
                  color: '#ef4444',
                  border: '1px solid rgba(239,68,68,0.3)',
                }}
              >
                <Trash2 size={14} /> Hapus
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
