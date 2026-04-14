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
        <div style={{ display: 'flex', gap: 24 }}>
          {(Object.entries(STATUS_CONFIG) as [TableStatus, (typeof STATUS_CONFIG)['available']][])
            .filter(([k]) => k !== 'unavailable')
            .map(([key, cfg]) => (
              <div key={key} style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                <div
                  style={{
                    width: 12,
                    height: 12,
                    borderRadius: 3,
                    background: cfg.dot,
                    boxShadow: `0 0 8px ${cfg.dot}66`,
                  }}
                />
                <span
                  style={{
                    fontSize: '0.78rem',
                    fontWeight: 600,
                    color: 'var(--text-2)',
                    textTransform: 'uppercase',
                    letterSpacing: '0.05em',
                  }}
                >
                  {cfg.label}
                </span>
              </div>
            ))}
        </div>

        <div style={{ display: 'flex', gap: 16 }}>
          {[
            { label: 'Total', value: tables.length, color: 'var(--text-3)' },
            { label: 'Available', value: available, color: '#10b981' },
            { label: 'On Dine', value: occupied, color: '#ef4444' },
          ].map(stat => (
            <div
              key={stat.label}
              style={{
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'flex-end',
                padding: '4px 0',
              }}
            >
              <span
                style={{
                  fontSize: '0.65rem',
                  color: 'var(--text-3)',
                  fontWeight: 700,
                  textTransform: 'uppercase',
                  letterSpacing: '0.1em',
                }}
              >
                {stat.label}
              </span>
              <span
                style={{
                  fontSize: '1.4rem',
                  fontWeight: 800,
                  color: stat.color,
                  lineHeight: 1,
                  marginTop: 2,
                }}
              >
                {stat.value}
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
            gridTemplateColumns: 'repeat(auto-fill, minmax(180px, 1fr))',
            gap: 24,
            padding: 32,
            background: 'radial-gradient(circle, var(--border-md) 1px, transparent 1px)',
            backgroundSize: '40px 40px',
            borderRadius: 24,
            minHeight: 600,
          }}
        >
          {tables.map(table => {
            const cfg = STATUS_CONFIG[table.status as TableStatus];
            const isUpdating = statusUpdating === table.id;
            return (
              <div
                key={table.id}
                className="table-group"
                style={{
                  position: 'relative',
                  display: 'flex',
                  flexDirection: 'column',
                  alignItems: 'center',
                  padding: 20,
                  transition: 'all 0.3s ease',
                }}
              >
                {/* Visual Table */}
                <div onClick={() => openEdit(table)} style={{ cursor: 'pointer' }}>
                  <TableVisual table={table} config={cfg} />
                </div>

                {/* Floating Command HUD */}
                <div
                  className="table-hud"
                  style={{
                    position: 'absolute',
                    bottom: -10,
                    left: '50%',
                    transform: 'translateX(-50%) translateY(10px)',
                    display: 'flex',
                    alignItems: 'center',
                    gap: 8,
                    padding: '8px 12px',
                    background: 'var(--bg-card)',
                    borderRadius: 12,
                    boxShadow: 'var(--shadow-lg)',
                    border: '1px solid var(--border-md)',
                    opacity: 0,
                    visibility: 'hidden',
                    transition: 'all 0.3s cubic-bezier(0.175, 0.885, 0.32, 1.275)',
                    zIndex: 100,
                    backdropFilter: 'blur(12px)',
                  }}
                >
                  <select
                    value={table.status}
                    disabled={isUpdating}
                    onChange={e => handleStatusChange(table, e.target.value as TableStatus)}
                    className="select-minimal"
                    style={{
                      fontSize: '0.7rem',
                      height: 28,
                      padding: '0 8px',
                      background: 'var(--bg-elevated)',
                      borderRadius: 6,
                      fontWeight: 700,
                    }}
                  >
                    <option value="available">🟢 Available</option>
                    <option value="occupied">🔴 On Dine</option>
                    <option value="reserved">🔵 Reserved</option>
                  </select>

                  <div style={{ width: 1, height: 16, background: 'var(--border-md)' }} />

                  <div style={{ display: 'flex', gap: 4 }}>
                    <button
                      className="btn btn-ghost btn-xs"
                      onClick={() => openEdit(table)}
                      style={{ width: 26, height: 26, padding: 0 }}
                      title="Edit Table"
                    >
                      <Pencil size={12} />
                    </button>
                    <button
                      className="btn btn-ghost btn-xs"
                      onClick={() => setDeleteConfirm({ open: true, table })}
                      style={{ width: 26, height: 26, padding: 0, color: '#ef4444' }}
                      title="Delete Table"
                    >
                      <Trash2 size={12} />
                    </button>
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
  const isCircle = table.capacity < 4;

  // Chair distribution
  let top = 0,
    bottom = 0,
    left = 0,
    right = 0;
  if (isRect) {
    top = Math.ceil(table.capacity / 2);
    bottom = table.capacity - top;
  } else if (isCircle) {
    top = 1;
    bottom = table.capacity > 1 ? 1 : 0;
    left = table.capacity > 2 ? 1 : 0;
  } else {
    top = 1;
    bottom = 1;
    left = 1;
    right = 1;
  }

  const Chair = ({ active, position = 'top' }: { active: boolean; position?: string }) => {
    const isHorizontal = position === 'left' || position === 'right';
    return (
      <div
        className="chair-icon"
        style={{
          width: isHorizontal ? 10 : 16,
          height: isHorizontal ? 16 : 10,
          position: 'relative',
          display: 'flex',
          flexDirection: position === 'top' ? 'column' : position === 'bottom' ? 'column-reverse' : position === 'left' ? 'row' : 'row-reverse',
          alignItems: 'center',
          gap: 0,
        }}
      >
        {/* Chair Seat */}
        <div
          style={{
            width: isHorizontal ? 8 : 14,
            height: isHorizontal ? 14 : 8,
            borderRadius: 3,
            background: active ? config.furniture : 'var(--bg-elevated)',
            border: `1px solid ${active ? config.color : 'var(--border-md)'}`,
            zIndex: 2,
          }}
        />
        {/* Chair Backrest */}
        <div
          style={{
            width: isHorizontal ? 3 : 10,
            height: isHorizontal ? 10 : 3,
            borderRadius: 2,
            background: active ? config.color : 'var(--text-3)',
            opacity: active ? 0.6 : 0.2,
            zIndex: 1,
          }}
        />
      </div>
    );
  };

  return (
    <div
      style={{
        position: 'relative',
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        padding: '24px 0',
      }}
    >
      {/* Top Chairs */}
      <div style={{ display: 'flex', gap: 6, marginBottom: 8 }}>
        {Array.from({ length: top }).map((_, i) => (
          <Chair key={i} active={true} position="top" />
        ))}
      </div>

      <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
        {/* Left Side Chair */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
          {Array.from({ length: left }).map((_, i) => (
            <Chair key={i} active={true} position="left" />
          ))}
        </div>

        {/* Table Body */}
        <div
          className="table-furniture"
          style={{
            width: isRect ? 140 : 100,
            height: 100,
            borderRadius: isCircle ? '50%' : 16,
            background: `linear-gradient(135deg, ${config.furniture} 0%, ${config.furniture}bb 100%)`,
            border: `1.5px solid ${config.color}22`,
            boxShadow: `inset 0 0 10px ${config.color}11`,
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            justifyContent: 'center',
            gap: 2,
            position: 'relative',
            transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
          }}
        >
          <div style={{ fontSize: '0.9rem', fontWeight: 800, color: config.color }}>
            #{table.table_number}
          </div>
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: 4,
              fontSize: '0.75rem',
              color: config.color,
              opacity: 0.7,
              fontWeight: 600,
            }}
          >
            <Users size={12} /> {table.capacity}
          </div>
        </div>

        {/* Right Side Chair */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
          {Array.from({ length: right }).map((_, i) => (
            <Chair key={i} active={true} position="right" />
          ))}
        </div>
      </div>

      {/* Bottom Chairs */}
      <div style={{ display: 'flex', gap: 6, marginTop: 8 }}>
        {Array.from({ length: bottom }).map((_, i) => (
          <Chair key={i} active={true} position="bottom" />
        ))}
      </div>
    </div>
  );
}
