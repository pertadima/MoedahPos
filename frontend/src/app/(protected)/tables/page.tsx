'use client';

import { useEffect, useState, useCallback, useMemo } from 'react';
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
  LayoutDashboard,
  CheckCircle2,
  Clock,
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
    label: 'Tersedia',
    color: '#047857',
    bg: '#ecfdf5',
    dot: '#10b981',
    furniture: '#c6f6d5',
  },
  occupied: {
    label: 'Sedang Makan',
    color: '#be123c',
    bg: '#fff1f2',
    dot: '#f43f5e',
    furniture: '#fed7e2',
  },
  reserved: {
    label: 'Dipesan',
    color: '#1e40af',
    bg: '#eff6ff',
    dot: '#3b82f6',
    furniture: '#dbeafe',
  },
  unavailable: {
    label: 'Tidak Tersedia',
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

  const stats = useMemo(() => {
    return {
      total: tables.length,
      available: tables.filter(t => t.status === 'available').length,
      occupied: tables.filter(t => t.status === 'occupied').length,
      reserved: tables.filter(t => t.status === 'reserved').length,
    };
  }, [tables]);

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

  return (
    <div className="w-full p-6">
      <style>{`
        @keyframes fadeInScale {
          from { opacity: 0; transform: scale(0.98) translateY(10px); }
          to { opacity: 1; transform: scale(1) translateY(0); }
        }
        .table-card-animate {
          animation: fadeInScale 0.4s cubic-bezier(0.175, 0.885, 0.32, 1.275) both;
        }
        .table-card {
          background: var(--bg-card);
          border: 1px solid var(--border);
          border-radius: 12px;
          padding: 16px;
          position: relative;
          transition: all 0.2s ease;
          cursor: pointer;
        }
        .table-card:hover {
          border-color: var(--brand);
          box-shadow: var(--shadow-md);
          transform: translateY(-2px);
        }
        .table-card.active {
          border-color: var(--brand);
          background: rgba(8, 132, 246, 0.02);
        }
        .table-hud {
          display: flex;
          align-items: center;
          gap: 8,
          margin-top: 12px;
          padding-top: 12px;
          border-top: 1px solid var(--border);
        }
      `}</style>
      {/* 1. Main Floor Plan Area */}
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
          marginBottom: 28,
        }}
      >
        <div>
          <h1 className="page-title">
            <LayoutDashboard size={20} style={{ color: 'var(--brand)' }} />
            Kelola Meja
          </h1>
          <p className="page-subtitle">
            {selectedStore.store_name} · {tables.length} meja
          </p>
        </div>
        <div style={{ display: 'flex', gap: 12, alignItems: 'center' }}>
          {can('products.create') && (
            <button className="btn btn-primary" onClick={openCreate} style={{ gap: 8 }}>
              <Plus size={16} /> Tambah Meja
            </button>
          )}
        </div>
      </div>

      {/* 2. Legend / Stats Section (Dashboard Style) */}
      {!loading && (
        <div
          style={{
            display: 'grid',
            gridTemplateColumns: 'repeat(4, 1fr)',
            gap: 16,
            marginBottom: 24,
          }}
        >
          {[
            {
              label: 'Total Meja',
              count: stats.total,
              icon: Grid3x3,
              color: 'var(--text-2)',
              bg: 'var(--bg-elevated)',
            },
            {
              label: 'Tersedia',
              count: stats.available,
              icon: CheckCircle2,
              color: '#10b981',
              bg: 'rgba(16,185,129,0.10)',
            },
            {
              label: 'Sedang Makan',
              count: stats.occupied,
              icon: Users,
              color: '#ef4444',
              bg: 'rgba(239,68,68,0.10)',
            },
            {
              label: 'Dipesan',
              count: stats.reserved,
              icon: Clock,
              color: '#3b82f6',
              bg: 'rgba(59,130,246,0.10)',
            },
          ].map((s, i) => (
            <div key={i} className="stat-card">
              <div className="stat-icon" style={{ background: s.bg }}>
                <s.icon size={22} style={{ color: s.color }} />
              </div>
              <div style={{ minWidth: 0, flex: 1 }}>
                <div className="stat-label">{s.label}</div>
                <div className="stat-val">
                  <span className="stat-number" style={{ fontSize: '1.5rem' }}>
                    {s.count}
                  </span>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Grid & Content */}
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
            gridTemplateColumns: 'repeat(4, 1fr)',
            gap: 20,
            minHeight: 400,
          }}
        >
          {tables.map((table, i) => {
            const cfg = STATUS_CONFIG[table.status as TableStatus];
            const isUpdating = statusUpdating === table.id;
            return (
              <div
                key={table.id}
                className="table-card table-card-animate"
                style={{ animationDelay: `${i * 0.05}s` }}
                onClick={() => openEdit(table)}
              >
                <div
                  style={{
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'flex-start',
                    marginBottom: 12,
                  }}
                >
                  <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                    <div
                      style={{
                        width: 36,
                        height: 36,
                        borderRadius: 8,
                        background: cfg.bg,
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        color: cfg.color,
                        fontWeight: 800,
                        fontSize: '0.9rem',
                      }}
                    >
                      {table.table_number}
                    </div>
                    <div>
                      <div style={{ fontSize: '0.85rem', fontWeight: 700, color: 'var(--text-1)' }}>
                        Meja {table.table_number}
                      </div>
                      <div
                        style={{
                          fontSize: '0.75rem',
                          color: 'var(--text-3)',
                          display: 'flex',
                          alignItems: 'center',
                          gap: 4,
                        }}
                      >
                        <Users size={12} /> {table.capacity} orang
                      </div>
                    </div>
                  </div>
                  <span
                    style={{
                      padding: '4px 8px',
                      borderRadius: 6,
                      fontSize: '0.65rem',
                      fontWeight: 700,
                      textTransform: 'uppercase',
                      letterSpacing: '0.02em',
                      background: cfg.bg,
                      color: cfg.color,
                    }}
                  >
                    {cfg.label}
                  </span>
                </div>

                <div
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'space-between',
                    marginTop: 16,
                    paddingTop: 16,
                    borderTop: '1px solid var(--border)',
                  }}
                  onClick={e => e.stopPropagation()}
                >
                  <select
                    value={table.status}
                    disabled={isUpdating}
                    onChange={e => handleStatusChange(table, e.target.value as TableStatus)}
                    className="select-minimal"
                    style={{
                      fontSize: '0.75rem',
                      height: 32,
                      width: 'auto',
                      padding: '0 28px 0 10px',
                      backgroundColor: 'var(--bg-elevated)',
                      borderRadius: 6,
                      fontWeight: 600,
                      flex: 1,
                      marginRight: 8,
                    }}
                  >
                    <option value="available">Tersedia</option>
                    <option value="occupied">Sedang Makan</option>
                    <option value="reserved">Dipesan</option>
                  </select>

                  <div style={{ display: 'flex', gap: 4 }}>
                    <button
                      className="btn btn-ghost btn-xs"
                      onClick={() => openEdit(table)}
                      style={{ width: 32, height: 32, padding: 0 }}
                    >
                      <Pencil size={14} />
                    </button>
                    <button
                      className="btn btn-ghost btn-xs"
                      onClick={() => setDeleteConfirm({ open: true, table })}
                      style={{ width: 32, height: 32, padding: 0, color: '#ef4444' }}
                    >
                      <Trash2 size={14} />
                    </button>
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      )}

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
