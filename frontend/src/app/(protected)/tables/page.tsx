'use client';

import { useEffect, useState, useCallback } from 'react';
import {
  Grid3x3,
  Plus,
  Pencil,
  Trash2,
  Check,
  X,
  Loader2,
  Users,
  AlertTriangle,
} from 'lucide-react';
import { useAuth } from '@/lib/auth/AuthContext';
import { usePermission } from '@/hooks/usePermission';
import { tablesApi } from '@/lib/api/store-apis';
import type { RestaurantTable, TableStatus } from '@/types';
import { getErrorMessage } from '@/lib/utils';

// ── Status helpers ─────────────────────────────────────────────────────────────

const STATUS_CONFIG: Record<
  TableStatus,
  { label: string; color: string; bg: string; dot: string; furniture: string }
> = {
  available: {
    label: 'Available',
    color: '#047857',
    bg: '#ecfdf5',
    dot: '#10b981',
    furniture: '#c6f6d5',
  },
  occupied: {
    label: 'On Dine',
    color: '#be123c',
    bg: '#fff1f2',
    dot: '#f43f5e',
    furniture: '#fed7e2',
  },
  reserved: {
    label: 'Reserved',
    color: '#1e40af',
    bg: '#eff6ff',
    dot: '#3b82f6',
    furniture: '#dbeafe',
  },
  unavailable: {
    label: 'Unavailable',
    color: '#4b5563',
    bg: '#f3f4f6',
    dot: '#6b7280',
    furniture: '#e5e7eb',
  },
};

// ── Page ──────────────────────────────────────────────────────────────────────

export default function TablesPage() {
  const { selectedStore } = useAuth();
  const { can } = usePermission();
  const [tables, setTables] = useState<RestaurantTable[]>([]);
  const [loading, setLoading] = useState(true);
  const [modal, setModal] = useState<{
    open: boolean;
    mode: 'create' | 'edit';
    table?: RestaurantTable;
  }>({ open: false, mode: 'create' });
  const [deleteConfirm, setDeleteConfirm] = useState<{ open: boolean; table?: RestaurantTable }>({
    open: false,
  });
  const [submitting, setSubmitting] = useState(false);
  const [statusUpdating, setStatusUpdating] = useState<string | null>(null);
  const [toast, setToast] = useState<{ msg: string; type: 'success' | 'error' } | null>(null);
  const [form, setForm] = useState({ table_number: '', capacity: 4, notes: '' });
  const [formError, setFormError] = useState('');

// ── Components ────────────────────────────────────────────────────────────────

  const storeId = selectedStore?.store_id ?? '';

  const fetchTables = useCallback(async () => {
    if (!storeId) return;
    setLoading(true);
    try {
      const res = await tablesApi.list(storeId);
      setTables(res.data);
    } catch (error) {
      showToast(getErrorMessage(error, 'Gagal memuat data meja'), 'error');
    } finally {
      setLoading(false);
    }
  }, [storeId]);

  useEffect(() => {
    fetchTables();
  }, [fetchTables]);

  const showToast = (msg: string, type: 'success' | 'error') => {
    setToast({ msg, type });
    setTimeout(() => setToast(null), 3000);
  };

  const openCreate = () => {
    setForm({ table_number: '', capacity: 4, notes: '' });
    setFormError('');
    setModal({ open: true, mode: 'create' });
  };

  const openEdit = (t: RestaurantTable) => {
    setForm({ table_number: t.table_number, capacity: t.capacity, notes: t.notes ?? '' });
    setFormError('');
    setModal({ open: true, mode: 'edit', table: t });
  };

  const closeModal = () => setModal({ open: false, mode: 'create' });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!form.table_number.trim()) {
      setFormError('Nomor meja wajib diisi');
      return;
    }
    if (form.capacity < 1) {
      setFormError('Kapasitas minimal 1');
      return;
    }
    setSubmitting(true);
    setFormError('');
    const payload = {
      table_number: form.table_number.trim(),
      capacity: form.capacity,
      notes: form.notes || undefined,
    };
    try {
      if (modal.mode === 'create') {
        await tablesApi.create(storeId, payload);
        showToast('Meja berhasil ditambahkan ✓', 'success');
      } else if (modal.table) {
        await tablesApi.update(storeId, modal.table.id, payload);
        showToast('Meja berhasil diperbarui ✓', 'success');
      } else {
        setFormError('Meja tidak ditemukan');
        return;
      }
      closeModal();
      fetchTables();
    } catch (error) {
      setFormError(getErrorMessage(error, 'Terjadi kesalahan'));
    } finally {
      setSubmitting(false);
    }
  };

  const handleStatusChange = async (table: RestaurantTable, status: TableStatus) => {
    setStatusUpdating(table.id);
    try {
      await tablesApi.updateStatus(storeId, table.id, status);
      showToast(`Meja ${table.table_number} → ${STATUS_CONFIG[status].label}`, 'success');
      fetchTables();
    } catch {
      showToast('Gagal update status meja', 'error');
    } finally {
      setStatusUpdating(null);
    }
  };

  const handleDelete = async () => {
    if (!deleteConfirm.table) return;
    try {
      await tablesApi.delete(storeId, deleteConfirm.table.id);
      showToast(`Meja ${deleteConfirm.table.table_number} dihapus`, 'success');
      setDeleteConfirm({ open: false });
      fetchTables();
    } catch (error) {
      showToast(getErrorMessage(error, 'Gagal menghapus'), 'error');
    }
  };

  if (!selectedStore)
    return (
      <div style={{ padding: 32 }}>
        <div className="empty-state card" style={{ padding: 48 }}>
          <Grid3x3 size={40} style={{ color: 'var(--text-3)' }} />
          <p style={{ fontWeight: 600, color: 'var(--text-2)' }}>Pilih toko terlebih dahulu</p>
        </div>
      </div>
    );

  if (selectedStore.store_type !== 'restaurant')
    return (
      <div style={{ padding: 32 }}>
        <div className="empty-state card" style={{ padding: 48 }}>
          <Grid3x3 size={40} style={{ color: 'var(--text-3)' }} />
          <p style={{ fontWeight: 600, color: 'var(--text-2)' }}>
            Manajemen Meja hanya tersedia untuk tipe Restoran
          </p>
          <p style={{ fontSize: '0.82rem' }}>
            Ubah tipe toko ke &ldquo;Restaurant&rdquo; di pengaturan toko.
          </p>
        </div>
      </div>
    );

  const available = tables.filter(t => t.status === 'available').length;
  const occupied = tables.filter(t => t.status === 'occupied').length;

  return (
    <div className="w-full p-6">
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
            <Grid3x3 size={22} style={{ color: '#fb923c' }} />
            Manajemen Meja
          </h1>
          <p className="page-subtitle">
            {selectedStore.store_name} · {tables.length} meja
          </p>
        </div>
        {can('products.create') && (
          <button className="btn btn-primary" onClick={openCreate} style={{ gap: 8 }}>
            <Plus size={16} /> Tambah Meja
          </button>
        )}
      </div>

      {/* Legend & Stats */}
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'flex-end',
          marginBottom: 24,
          gap: 16,
        }}
      >
        <div style={{ display: 'flex', gap: 20 }}>
          {(Object.entries(STATUS_CONFIG) as [TableStatus, (typeof STATUS_CONFIG)['available']][])
            .filter(([k]) => k !== 'unavailable')
            .map(([key, cfg]) => (
              <div key={key} style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                <div
                  style={{ width: 10, height: 10, borderRadius: '50%', background: cfg.dot }}
                />
                <span style={{ fontSize: '0.82rem', fontWeight: 500, color: 'var(--text-2)' }}>
                  {cfg.label}
                </span>
              </div>
            ))}
        </div>

        <div style={{ display: 'flex', gap: 12 }}>
          {[
            { label: 'Total', value: tables.length, color: 'var(--text-3)' },
            { label: 'Available', value: available, color: '#10b981' },
            { label: 'On Dine', value: occupied, color: '#ef4444' },
          ].map(stat => (
            <div
              key={stat.label}
              style={{
                display: 'flex',
                alignItems: 'baseline',
                gap: 6,
                padding: '4px 12px',
                background: 'var(--bg-card)',
                borderRadius: 8,
                border: '1px solid var(--border)',
              }}
            >
              <span style={{ fontSize: '1.2rem', fontWeight: 700, color: stat.color }}>
                {stat.value}
              </span>
              <span style={{ fontSize: '0.72rem', color: 'var(--text-3)', fontWeight: 500 }}>
                {stat.label}
              </span>
            </div>
          ))}
        </div>
      </div>

      {/* Table grid */}
      {loading ? (
        <div style={{ display: 'flex', justifyContent: 'center', padding: 60 }}>
          <Loader2 size={28} className="loading-spin" style={{ color: 'var(--accent-em)' }} />
        </div>
      ) : tables.length === 0 ? (
        <div className="empty-state card" style={{ padding: 60 }}>
          <Grid3x3 size={48} style={{ color: 'var(--text-3)' }} />
          <p style={{ fontWeight: 600, color: 'var(--text-2)' }}>Belum ada meja</p>
          <p style={{ fontSize: '0.85rem' }}>
            Klik &ldquo;Tambah Meja&rdquo; untuk menambahkan meja pertama.
          </p>
        </div>
      ) : (
        <div
          style={{
            display: 'grid',
            gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))',
            gap: 24,
          }}
        >
          {tables.map(table => {
            const cfg = STATUS_CONFIG[table.status as TableStatus];
            const isUpdating = statusUpdating === table.id;
            return (
              <div
                key={table.id}
                className="card"
                style={{
                  padding: 16,
                  display: 'flex',
                  flexDirection: 'column',
                  alignItems: 'center',
                  background: 'var(--bg-card)',
                  border: `1px solid ${table.status === 'available' ? 'var(--border)' : cfg.color + '33'}`,
                  boxShadow: table.status !== 'available' ? `0 4px 12px ${cfg.color}11` : 'none',
                }}
              >
                {/* Visual Table */}
                <TableVisual table={table} config={cfg} />

                {/* Status selector & Actions */}
                <div style={{ width: '100%', marginTop: 12 }}>
                  <div style={{ display: 'flex', gap: 6 }}>
                    <select
                      value={table.status}
                      disabled={isUpdating}
                      onChange={e => handleStatusChange(table, e.target.value as TableStatus)}
                      className="select-minimal"
                      style={{
                        flex: 1,
                        fontSize: '0.75rem',
                        height: 32,
                        padding: '0 8px',
                      }}
                    >
                      <option value="available">🟢 Available</option>
                      <option value="occupied">🔴 On Dine</option>
                      <option value="reserved">🔵 Reserved</option>
                    </select>

                    {can('products.update') && (
                      <button
                        className="btn btn-ghost btn-sm"
                        onClick={() => openEdit(table)}
                        style={{ width: 32, height: 32, padding: 0 }}
                      >
                        <Pencil size={13} />
                      </button>
                    )}
                    {can('products.delete') && (
                      <button
                        className="btn btn-ghost btn-sm"
                        onClick={() => setDeleteConfirm({ open: true, table })}
                        style={{ width: 32, height: 32, padding: 0, color: '#ef4444' }}
                      >
                        <Trash2 size={13} />
                      </button>
                    )}
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      )}

      {/* ── Create / Edit Modal ──────────────────────────────────────────────── */}
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
            style={{ width: '100%', maxWidth: 420, padding: 28, animation: 'slideIn 0.2s ease' }}
          >
            <div
              style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
                marginBottom: 20,
              }}
            >
              <div
                style={{
                  fontWeight: 700,
                  fontSize: '1rem',
                  display: 'flex',
                  alignItems: 'center',
                  gap: 8,
                }}
              >
                <Grid3x3 size={18} style={{ color: '#fb923c' }} />
                {modal.mode === 'create' ? 'Tambah Meja' : `Edit Meja ${modal.table?.table_number}`}
              </div>
              <button onClick={closeModal} className="btn btn-ghost btn-sm" style={{ padding: 6 }}>
                <X size={16} />
              </button>
            </div>
            <form onSubmit={handleSubmit}>
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
                    Nomor Meja <span style={{ color: '#ef4444' }}>*</span>
                  </label>
                  <input
                    className="input"
                    placeholder="1, A1, VIP-1…"
                    value={form.table_number}
                    autoFocus
                    onChange={e => setForm(f => ({ ...f, table_number: e.target.value }))}
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
                    Kapasitas
                  </label>
                  <input
                    className="input"
                    type="number"
                    min={1}
                    max={100}
                    value={form.capacity}
                    onChange={e =>
                      setForm(f => ({ ...f, capacity: parseInt(e.target.value) || 1 }))
                    }
                  />
                </div>
              </div>
              <div style={{ marginBottom: 18 }}>
                <label
                  style={{
                    display: 'block',
                    fontSize: '0.8rem',
                    fontWeight: 600,
                    marginBottom: 5,
                    color: 'var(--text-2)',
                  }}
                >
                  Catatan{' '}
                  <span style={{ fontSize: '0.72rem', color: 'var(--text-3)', fontWeight: 400 }}>
                    (opsional)
                  </span>
                </label>
                <input
                  className="input"
                  placeholder="cth. Di dekat jendela, meja roda…"
                  value={form.notes}
                  onChange={e => setForm(f => ({ ...f, notes: e.target.value }))}
                />
              </div>
              {formError && (
                <div
                  style={{
                    background: 'rgba(239,68,68,0.12)',
                    border: '1px solid rgba(239,68,68,0.3)',
                    borderRadius: 8,
                    padding: '10px 14px',
                    marginBottom: 14,
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
                      <Check size={14} /> Simpan
                    </>
                  )}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* ── Delete Confirm ──────────────────────────────────────────────────── */}
      {deleteConfirm.open && deleteConfirm.table && (
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
              maxWidth: 380,
              padding: 28,
              textAlign: 'center',
              animation: 'slideIn 0.2s ease',
            }}
          >
            <div
              style={{
                width: 48,
                height: 48,
                borderRadius: '50%',
                background: 'rgba(239,68,68,0.15)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                margin: '0 auto 14px',
              }}
            >
              <AlertTriangle size={22} style={{ color: '#ef4444' }} />
            </div>
            <div style={{ fontWeight: 700, fontSize: '1rem', marginBottom: 8 }}>Hapus Meja?</div>
            <div
              style={{
                fontSize: '0.85rem',
                color: 'var(--text-2)',
                marginBottom: 20,
                lineHeight: 1.6,
              }}
            >
              Meja <strong>&ldquo;{deleteConfirm.table.table_number}&rdquo;</strong> akan dihapus
              secara permanen.
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

// ── Components ────────────────────────────────────────────────────────────────

function TableVisual({ table, config }: { table: RestaurantTable; config: any }) {
  const isRect = table.capacity >= 5;
  const chairs = Array.from({ length: table.capacity });

  // Distribute chairs: Top, Bottom, Left, Right
  let top = 0,
    bottom = 0,
    left = 0,
    right = 0;
  if (isRect) {
    top = Math.ceil(table.capacity / 2);
    bottom = table.capacity - top;
  } else {
    // Square distribution
    top = 1;
    bottom = table.capacity > 1 ? 1 : 0;
    left = table.capacity > 2 ? 1 : 0;
    right = table.capacity > 3 ? 1 : 0;
  }

  const Chair = ({ active }: { active: boolean }) => (
    <div
      style={{
        width: 14,
        height: 14,
        borderRadius: 4,
        border: `1.5px solid ${active ? config.color : '#cbd5e1'}`,
        background: active ? config.furniture : 'transparent',
        opacity: active ? 1 : 0.3,
      }}
    />
  );

  return (
    <div
      style={{
        position: 'relative',
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        padding: '20px 0',
      }}
    >
      {/* Top Chairs */}
      <div style={{ display: 'flex', gap: 6, marginBottom: 8 }}>
        {Array.from({ length: top }).map((_, i) => (
          <Chair key={i} active={true} />
        ))}
        {!isRect && top === 1 && table.capacity < 1 && <Chair active={false} />}
      </div>

      <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
        {/* Left Side Chair */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
          {Array.from({ length: left }).map((_, i) => (
            <Chair key={i} active={true} />
          ))}
        </div>

        {/* Table Body */}
        <div
          style={{
            width: isRect ? 140 : 90,
            height: 90,
            borderRadius: 12,
            background: config.furniture,
            border: `1px solid ${config.color}33`,
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            justifyContent: 'center',
            gap: 2,
            transition: 'all 0.2s ease',
            cursor: 'default',
          }}
        >
          <div style={{ fontSize: '0.85rem', fontWeight: 700, color: config.color }}>
            Table #{table.table_number}
          </div>
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: 4,
              fontSize: '0.75rem',
              color: config.color,
              opacity: 0.8,
            }}
          >
            <Users size={12} /> {table.capacity}
          </div>
        </div>

        {/* Right Side Chair */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
          {Array.from({ length: right }).map((_, i) => (
            <Chair key={i} active={true} />
          ))}
        </div>
      </div>

      {/* Bottom Chairs */}
      <div style={{ display: 'flex', gap: 6, marginTop: 8 }}>
        {Array.from({ length: bottom }).map((_, i) => (
          <Chair key={i} active={true} />
        ))}
      </div>
    </div>
  );
}
