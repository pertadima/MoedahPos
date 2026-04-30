'use client';

import { useState, useEffect, useCallback } from 'react';
import {
  Users,
  Plus,
  Loader2,
  X,
  Search,
  Phone,
  Mail,
  MapPin,
  FileText,
  Edit3,
  Archive,
  ChevronRight,
  TrendingUp,
  TrendingDown,
  RotateCcw,
  SlidersHorizontal,
  History,
  ChevronLeft,
} from 'lucide-react';
import { useAuth } from '@/lib/auth/AuthContext';
import { usePermission } from '@/hooks/usePermission';
import { customersApi, loyaltyApi } from '@/lib/api/store-apis';
import Portal from '@/components/ui/Portal';
import { formatDate } from '@/lib/utils';
import type { Customer, LoyaltyLedgerEntry, PaginatedData } from '@/types';
import { ApiError } from '@/lib/api/client';

// ── Empty form ────────────────────────────────────────────────────────────────

// ── Customer Form Modal ───────────────────────────────────────────────────────
interface FormModalProps {
  storeId: string;
  initial?: Customer | null;
  onSuccess: () => void;
  onClose: () => void;
}

function FormModal({ storeId, initial, onSuccess, onClose }: FormModalProps) {
  const isEdit = !!initial;
  const [form, setForm] = useState({
    name: initial?.name ?? '',
    phone: initial?.phone ?? '',
    email: initial?.email ?? '',
    address: initial?.address ?? '',
    notes: initial?.notes ?? '',
  });
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');

  const set =
    (k: keyof typeof form) => (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) =>
      setForm(f => ({ ...f, [k]: e.target.value }));

  const handleSave = async () => {
    if (!form.name.trim()) {
      setError('Nama wajib diisi');
      return;
    }
    setSaving(true);
    setError('');
    try {
      if (isEdit && initial) {
        await customersApi.update(storeId, initial.id, form);
      } else {
        await customersApi.create(storeId, form);
      }
      onSuccess();
    } catch (e) {
      setError(e instanceof ApiError ? e.message : 'Gagal menyimpan');
    } finally {
      setSaving(false);
    }
  };

  return (
    <Portal>
      <div className="modal-overlay" style={{ zIndex: 5000 }} onClick={onClose}>
        <div className="modal-box" style={{ maxWidth: 460 }} onClick={e => e.stopPropagation()}>
          <div
            style={{
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
              marginBottom: 18,
            }}
          >
            <h2 style={{ fontWeight: 800 }}>{isEdit ? 'Edit Customer' : 'Tambah Customer'}</h2>
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
                Nama <span style={{ color: '#ef4444' }}>*</span>
              </label>
              <input
                className="input"
                value={form.name}
                onChange={set('name')}
                placeholder="Nama lengkap customer"
                autoFocus
              />
            </div>
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 10 }}>
              <div className="input-group">
                <label className="input-label">Telepon</label>
                <input
                  className="input"
                  value={form.phone}
                  onChange={set('phone')}
                  placeholder="08xx-xxxx-xxxx"
                />
              </div>
              <div className="input-group">
                <label className="input-label">Email</label>
                <input
                  type="email"
                  className="input"
                  value={form.email}
                  onChange={set('email')}
                  placeholder="email@contoh.com"
                />
              </div>
            </div>
            <div className="input-group">
              <label className="input-label">Alamat</label>
              <input
                className="input"
                value={form.address}
                onChange={set('address')}
                placeholder="Alamat pengiriman"
              />
            </div>
            <div className="input-group">
              <label className="input-label">Catatan</label>
              <textarea
                className="input"
                value={form.notes}
                onChange={set('notes')}
                rows={2}
                placeholder="Preferensi, alergi, dll..."
                style={{ resize: 'vertical' }}
              />
            </div>
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
              {saving ? 'Menyimpan...' : isEdit ? 'Simpan Perubahan' : 'Tambah Customer'}
            </button>
          </div>
        </div>
      </div>
    </Portal>
  );
}

// ── Delete Confirm ────────────────────────────────────────────────────────────
function DeleteConfirm({
  customer,
  storeId,
  onSuccess,
  onClose,
}: {
  customer: Customer;
  storeId: string;
  onSuccess: () => void;
  onClose: () => void;
}) {
  const [loading, setLoading] = useState(false);
  const handleDelete = async () => {
    setLoading(true);
    try {
      await customersApi.delete(storeId, customer.id);
      onSuccess();
    } catch (e) {
      alert(e instanceof ApiError ? e.message : 'Gagal menonaktifkan');
    } finally {
      setLoading(false);
    }
  };
  return (
    <Portal>
      <div className="modal-overlay" style={{ zIndex: 5000 }} onClick={onClose}>
        <div className="modal-box" style={{ maxWidth: 400 }} onClick={e => e.stopPropagation()}>
          <div style={{ textAlign: 'center', padding: '8px 0 16px' }}>
            <Archive size={28} style={{ color: '#f59e0b', marginBottom: 12 }} />
            <h2 style={{ fontWeight: 800, marginBottom: 8 }}>Nonaktifkan Customer?</h2>
            <p style={{ color: 'var(--text-2)', fontSize: '0.875rem', lineHeight: 1.6 }}>
              <strong>{customer.name}</strong> akan diarsipkan dan tidak muncul di daftar customer
              maupun pencarian kasir. Data transaksi tetap tersimpan dan customer dapat dipulihkan
              dari database jika diperlukan.
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
    </Portal>
  );
}

// ── Riwayat Poin (Point History) ──────────────────────────────────────────────
function RiwayatPoin({ storeId, customerId }: { storeId: string; customerId: string }) {
  const [entries, setEntries] = useState<LoyaltyLedgerEntry[]>([]);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const PER_PAGE = 10;

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setLoading(true);
    loyaltyApi
      .getHistoryPaginated(storeId, customerId, page, PER_PAGE)
      .then(r => {
        setEntries(r.data ?? []);
        setTotal(r.meta?.total ?? 0);
      })
      .catch(console.error)
      .finally(() => setLoading(false));
  }, [storeId, customerId, page]);

  const typeIcon = (type: LoyaltyLedgerEntry['type']) => {
    switch (type) {
      case 'EARN':
        return <TrendingUp size={12} style={{ color: '#10b981' }} />;
      case 'SPEND':
        return <TrendingDown size={12} style={{ color: '#f59e0b' }} />;
      case 'VOID':
        return <RotateCcw size={12} style={{ color: '#f87171' }} />;
      case 'ADJUST':
        return <SlidersHorizontal size={12} style={{ color: '#a78bfa' }} />;
    }
  };

  const typeLabel: Record<LoyaltyLedgerEntry['type'], string> = {
    EARN: 'Earn',
    SPEND: 'Redeem',
    VOID: 'Void',
    ADJUST: 'Adjust',
  };

  const totalPages = Math.ceil(total / PER_PAGE);

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
      {/* Header */}
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: 6,
          fontWeight: 700,
          fontSize: '0.8rem',
        }}
      >
        <History size={14} style={{ color: 'var(--accent-em)' }} />
        Riwayat Poin
        {total > 0 && <span style={{ color: 'var(--text-3)', fontWeight: 400 }}>({total})</span>}
      </div>

      {loading ? (
        <div style={{ display: 'flex', justifyContent: 'center', padding: 16 }}>
          <Loader2 size={18} className="loading-spin" style={{ color: 'var(--accent-em)' }} />
        </div>
      ) : entries.length === 0 ? (
        <div
          style={{
            textAlign: 'center',
            color: 'var(--text-3)',
            fontSize: '0.8rem',
            padding: '12px 0',
          }}
        >
          Belum ada riwayat poin
        </div>
      ) : (
        <>
          {entries.map(e => (
            <div
              key={e.id}
              style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
                background: 'var(--bg-elevated)',
                borderRadius: 8,
                padding: '8px 12px',
                fontSize: '0.78rem',
              }}
            >
              {/* Left: icon + type + date */}
              <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                <div
                  style={{
                    width: 24,
                    height: 24,
                    borderRadius: '50%',
                    background: 'var(--bg-card)',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    flexShrink: 0,
                  }}
                >
                  {typeIcon(e.type)}
                </div>
                <div>
                  <div style={{ fontWeight: 600 }}>{typeLabel[e.type]}</div>
                  <div style={{ color: 'var(--text-3)', fontSize: '0.7rem' }}>
                    {new Date(e.created_at).toLocaleDateString('id-ID', {
                      day: '2-digit',
                      month: 'short',
                      year: 'numeric',
                    })}
                  </div>
                  {e.transaction_id && (
                    <div
                      style={{
                        color: 'var(--text-3)',
                        fontSize: '0.68rem',
                        fontFamily: 'monospace',
                        overflow: 'hidden',
                        textOverflow: 'ellipsis',
                        maxWidth: 130,
                        whiteSpace: 'nowrap',
                      }}
                      title={`TXN: ${e.transaction_id}`}
                    >
                      TXN: {e.transaction_id.slice(0, 8)}…
                    </div>
                  )}
                </div>
              </div>

              {/* Right: delta + snapshot */}
              <div style={{ textAlign: 'right' }}>
                <div
                  style={{
                    fontWeight: 700,
                    fontSize: '0.85rem',
                    color:
                      e.points_delta > 0
                        ? '#10b981'
                        : e.points_delta < 0
                          ? '#f87171'
                          : 'var(--text-2)',
                  }}
                >
                  {e.points_delta > 0 ? '+' : ''}
                  {e.points_delta.toLocaleString('id-ID')} pts
                </div>
                {e.balance_snapshot !== undefined && (
                  <div style={{ color: 'var(--text-3)', fontSize: '0.7rem' }}>
                    Saldo: {e.balance_snapshot.toLocaleString('id-ID')}
                  </div>
                )}
              </div>
            </div>
          ))}

          {/* Pagination */}
          {totalPages > 1 && (
            <div style={{ display: 'flex', justifyContent: 'center', gap: 8, marginTop: 4 }}>
              <button
                className="btn btn-ghost btn-sm"
                disabled={page <= 1}
                onClick={() => setPage(p => p - 1)}
              >
                <ChevronLeft size={13} />
              </button>
              <span style={{ alignSelf: 'center', fontSize: '0.75rem', color: 'var(--text-3)' }}>
                {page} / {totalPages}
              </span>
              <button
                className="btn btn-ghost btn-sm"
                disabled={page >= totalPages}
                onClick={() => setPage(p => p + 1)}
              >
                <ChevronRight size={13} />
              </button>
            </div>
          )}
        </>
      )}
    </div>
  );
}

// ── Detail Drawer ─────────────────────────────────────────────────────────────
function DetailDrawer({
  customer,
  storeId,
  onClose,
  onEdit,
  onDelete,
  canEdit,
  canDelete,
}: {
  customer: Customer;
  storeId: string;
  onClose: () => void;
  onEdit: () => void;
  onDelete: () => void;
  canEdit?: boolean;
  canDelete?: boolean;
}) {
  const avatar = customer.name.charAt(0).toUpperCase();
  return (
    <Portal>
      <div
        onClick={onClose}
        style={{
          position: 'fixed',
          inset: 0,
          background: 'rgba(0,0,0,0.5)',
          backdropFilter: 'blur(2px)',
          zIndex: 5000,
        }}
      />
      <div
        style={{
          position: 'fixed',
          top: 0,
          right: 0,
          bottom: 0,
          width: 400,
          background: 'var(--bg-card)',
          borderLeft: '1px solid var(--border)',
          zIndex: 5001,
          overflowY: 'auto',
          display: 'flex',
          flexDirection: 'column',
        }}
      >
        <div
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            padding: '18px 20px',
            borderBottom: '1px solid var(--border)',
          }}
        >
          <div style={{ fontWeight: 800 }}>Detail Customer</div>
          <div style={{ display: 'flex', gap: 6 }}>
            {canEdit !== false && (
              <button className="btn btn-secondary btn-sm" onClick={onEdit}>
                <Edit3 size={13} /> Edit
              </button>
            )}
            {canDelete !== false && (
              <button
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 4,
                  padding: '5px 10px',
                  borderRadius: 6,
                  border: '1px solid rgba(245,158,11,0.4)',
                  background: 'rgba(245,158,11,0.1)',
                  color: '#f59e0b',
                  fontSize: '0.78rem',
                  fontWeight: 600,
                  cursor: 'pointer',
                }}
                onClick={onDelete}
              >
                <Archive size={12} /> Nonaktifkan
              </button>
            )}
            <button className="btn btn-ghost btn-sm" onClick={onClose}>
              <X size={15} />
            </button>
          </div>
        </div>
        <div style={{ padding: 20, display: 'flex', flexDirection: 'column', gap: 18 }}>
          {/* Avatar + name */}
          <div style={{ display: 'flex', alignItems: 'center', gap: 14 }}>
            <div
              style={{
                width: 52,
                height: 52,
                borderRadius: '50%',
                background: 'linear-gradient(135deg, var(--accent-in), var(--accent-em))',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                fontWeight: 800,
                fontSize: '1.2rem',
                color: '#fff',
                flexShrink: 0,
              }}
            >
              {avatar}
            </div>
            <div>
              <div style={{ fontWeight: 800, fontSize: '1rem' }}>{customer.name}</div>
              <div style={{ fontSize: '0.75rem', color: 'var(--text-3)', marginTop: 2 }}>
                Customer sejak {formatDate(customer.created_at)}
              </div>
            </div>
          </div>

          {/* Details */}
          <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
            {customer.phone && (
              <div
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 10,
                  background: 'var(--bg-elevated)',
                  borderRadius: 10,
                  padding: '10px 14px',
                }}
              >
                <Phone size={14} style={{ color: 'var(--accent-em)', flexShrink: 0 }} />
                <div>
                  <div style={{ fontSize: '0.68rem', color: 'var(--text-3)' }}>Telepon</div>
                  <div style={{ fontWeight: 600 }}>{customer.phone}</div>
                </div>
              </div>
            )}
            {customer.email && (
              <div
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 10,
                  background: 'var(--bg-elevated)',
                  borderRadius: 10,
                  padding: '10px 14px',
                }}
              >
                <Mail size={14} style={{ color: 'var(--accent-em)', flexShrink: 0 }} />
                <div>
                  <div style={{ fontSize: '0.68rem', color: 'var(--text-3)' }}>Email</div>
                  <div style={{ fontWeight: 600 }}>{customer.email}</div>
                </div>
              </div>
            )}
            {customer.address && (
              <div
                style={{
                  display: 'flex',
                  alignItems: 'flex-start',
                  gap: 10,
                  background: 'var(--bg-elevated)',
                  borderRadius: 10,
                  padding: '10px 14px',
                }}
              >
                <MapPin
                  size={14}
                  style={{ color: 'var(--accent-em)', flexShrink: 0, marginTop: 3 }}
                />
                <div>
                  <div style={{ fontSize: '0.68rem', color: 'var(--text-3)' }}>Alamat</div>
                  <div style={{ fontWeight: 600 }}>{customer.address}</div>
                </div>
              </div>
            )}
            {customer.notes && (
              <div
                style={{
                  display: 'flex',
                  alignItems: 'flex-start',
                  gap: 10,
                  background: 'var(--bg-elevated)',
                  borderRadius: 10,
                  padding: '10px 14px',
                }}
              >
                <FileText
                  size={14}
                  style={{ color: 'var(--accent-em)', flexShrink: 0, marginTop: 3 }}
                />
                <div>
                  <div style={{ fontSize: '0.68rem', color: 'var(--text-3)' }}>Catatan</div>
                  <div style={{ fontSize: '0.85rem' }}>{customer.notes}</div>
                </div>
              </div>
            )}
            {!customer.phone && !customer.email && !customer.address && !customer.notes && (
              <div
                style={{
                  textAlign: 'center',
                  color: 'var(--text-3)',
                  padding: '16px 0',
                  fontSize: '0.85rem',
                }}
              >
                Tidak ada info tambahan
              </div>
            )}
          </div>

          {/* ── Loyalty Point History ── */}
          <div
            style={{
              borderTop: '1px solid var(--border)',
              paddingTop: 14,
            }}
          >
            <RiwayatPoin storeId={storeId} customerId={customer.id} />
          </div>
        </div>
      </div>
    </Portal>
  );
}

// ── Main Page ─────────────────────────────────────────────────────────────────
export default function CustomersPage() {
  const { selectedStore } = useAuth();
  const { can } = usePermission();
  const [customers, setCustomers] = useState<Customer[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState('');
  const [form, setForm] = useState<'create' | Customer | null>(null);
  const [detail, setDetail] = useState<Customer | null>(null);
  const [deleting, setDeleting] = useState<Customer | null>(null);

  const storeId = selectedStore?.store_id;
  const PER_PAGE = 20;

  const load = useCallback(() => {
    if (!storeId) return;
    setLoading(true);
    customersApi
      .list(storeId, { page, per_page: PER_PAGE, search: search || undefined })
      .then(r => {
        const d = r.data as PaginatedData<Customer>;
        setCustomers(d.data ?? []);
        setTotal(d.meta?.total ?? 0);
      })
      .catch(console.error)
      .finally(() => setLoading(false));
  }, [storeId, page, search]);

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    load();
  }, [load]);

  // Debounce search — reset page when search changes
  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setPage(1);
  }, [search]);

  const onSuccess = () => {
    setForm(null);
    setDeleting(null);
    setDetail(null);
    load();
  };

  if (!selectedStore)
    return (
      <div style={{ padding: 32 }}>
        <div className="empty-state card" style={{ padding: 40 }}>
          <Users size={40} />
          <p>Pilih toko terlebih dahulu</p>
        </div>
      </div>
    );

  return (
    <div className="w-full p-6">
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
            <Users size={22} style={{ color: 'var(--accent-em)' }} />
            Customer
          </h1>
          <p className="page-subtitle">
            {total} customer terdaftar · {selectedStore.store_name}
          </p>
        </div>
        {can('sales.create') && (
          <button className="btn btn-primary" onClick={() => setForm('create')}>
            <Plus size={15} /> Tambah Customer
          </button>
        )}
      </div>

      {/* Search */}
      <div
        className="reveal-animate"
        style={{ position: 'relative', marginBottom: 16, maxWidth: 360, animationDelay: '0.1s' }}
      >
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
          style={{ paddingLeft: 36 }}
          placeholder="Cari nama atau telepon..."
          value={search}
          onChange={e => setSearch(e.target.value)}
        />
      </div>

      {/* Table */}
      <div className="card reveal-animate" style={{ padding: 0, animationDelay: '0.2s' }}>
        {loading ? (
          <div style={{ display: 'flex', justifyContent: 'center', padding: 40 }}>
            <Loader2 size={24} className="loading-spin" style={{ color: 'var(--accent-em)' }} />
          </div>
        ) : customers.length === 0 ? (
          <div className="empty-state">
            <Users size={36} />
            <p>{search ? `Tidak ada hasil untuk "${search}"` : 'Belum ada customer'}</p>
          </div>
        ) : (
          <>
            <div className="tbl-container">
              <table className="tbl">
                <thead>
                  <tr>
                    <th>Nama</th>
                    <th>Telepon</th>
                    <th>Email</th>
                    <th>Alamat</th>
                    <th>Loyalty</th>
                    <th>Terdaftar</th>
                    <th>Aksi</th>
                  </tr>
                </thead>
                <tbody>
                  {customers.map((c, i) => (
                    <tr
                      key={c.id}
                      className="reveal-animate"
                      style={{
                        cursor: 'pointer',
                        animationDelay: `${0.25 + i * 0.02}s`,
                      }}
                      onClick={() => setDetail(c)}
                    >
                      <td>
                        <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                          <div
                            style={{
                              width: 32,
                              height: 32,
                              borderRadius: '50%',
                              background:
                                'linear-gradient(135deg, var(--accent-in), var(--accent-em))',
                              display: 'flex',
                              alignItems: 'center',
                              justifyContent: 'center',
                              fontWeight: 700,
                              fontSize: '0.8rem',
                              color: '#fff',
                              flexShrink: 0,
                            }}
                          >
                            {c.name.charAt(0).toUpperCase()}
                          </div>
                          <span style={{ fontWeight: 600 }}>{c.name}</span>
                        </div>
                      </td>
                      <td style={{ color: 'var(--text-2)', fontSize: '0.85rem' }}>
                        {c.phone ? (
                          <span style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
                            <Phone size={12} />
                            {c.phone}
                          </span>
                        ) : (
                          '—'
                        )}
                      </td>
                      <td style={{ color: 'var(--text-2)', fontSize: '0.85rem' }}>
                        {c.email ?? '—'}
                      </td>
                      <td
                        style={{
                          color: 'var(--text-3)',
                          fontSize: '0.82rem',
                          maxWidth: 180,
                          overflow: 'hidden',
                          textOverflow: 'ellipsis',
                          whiteSpace: 'nowrap',
                        }}
                      >
                        {c.address ?? '—'}
                      </td>
                      <td>
                        <div
                          style={{
                            display: 'inline-flex',
                            alignItems: 'center',
                            gap: 4,
                            background: 'rgba(251,191,36,0.15)',
                            border: '1px solid rgba(251,191,36,0.3)',
                            borderRadius: 6,
                            padding: '2px 8px',
                            color: '#f59e0b',
                            fontSize: '0.75rem',
                            fontWeight: 600,
                          }}
                        >
                          <svg
                            width="12"
                            height="12"
                            viewBox="0 0 24 24"
                            fill="#f59e0b"
                            stroke="#f59e0b"
                            strokeWidth="2"
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            style={{ marginTop: -1 }}
                          >
                            <polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2" />
                          </svg>
                          {(c.loyalty_balance ?? 0).toLocaleString('id-ID')} pts
                          {c.loyalty_tier && (
                            <span
                              style={{ fontWeight: 400, color: 'var(--text-3)', marginLeft: 2 }}
                            >
                              · {c.loyalty_tier.name}
                            </span>
                          )}
                        </div>
                      </td>
                      <td style={{ color: 'var(--text-3)', fontSize: '0.8rem' }}>
                        {formatDate(c.created_at)}
                      </td>
                      <td>
                        <div style={{ display: 'flex', gap: 4 }}>
                          {can('sales.create') && (
                            <button
                              className="btn btn-ghost btn-sm"
                              onClick={e => {
                                e.stopPropagation();
                                setForm(c);
                              }}
                              title="Edit"
                            >
                              <Edit3 size={13} />
                            </button>
                          )}
                          {can('sales.create') && (
                            <button
                              title="Nonaktifkan"
                              style={{
                                display: 'flex',
                                alignItems: 'center',
                                gap: 3,
                                padding: '4px 6px',
                                borderRadius: 6,
                                border: '1px solid rgba(245,158,11,0.35)',
                                background: 'rgba(245,158,11,0.08)',
                                color: '#f59e0b',
                                cursor: 'pointer',
                                fontSize: '0.72rem',
                              }}
                              onClick={e => {
                                e.stopPropagation();
                                setDeleting(c);
                              }}
                            >
                              <Archive size={12} />
                            </button>
                          )}
                          <button className="btn btn-ghost btn-sm" onClick={() => setDetail(c)}>
                            <ChevronRight size={14} />
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            {/* Pagination */}
            {total > PER_PAGE && (
              <div
                style={{
                  display: 'flex',
                  justifyContent: 'center',
                  gap: 8,
                  padding: '12px 16px',
                  borderTop: '1px solid var(--border)',
                }}
              >
                <button
                  className="btn btn-secondary btn-sm"
                  disabled={page <= 1}
                  onClick={() => setPage(p => p - 1)}
                >
                  Sebelumnya
                </button>
                <span style={{ alignSelf: 'center', fontSize: '0.82rem', color: 'var(--text-2)' }}>
                  Halaman {page} dari {Math.ceil(total / PER_PAGE)}
                </span>
                <button
                  className="btn btn-secondary btn-sm"
                  disabled={page >= Math.ceil(total / PER_PAGE)}
                  onClick={() => setPage(p => p + 1)}
                >
                  Berikutnya
                </button>
              </div>
            )}
          </>
        )}
      </div>

      {/* Modals & Drawer */}
      {form === 'create' &&
        (storeId ? (
          <FormModal storeId={storeId} onSuccess={onSuccess} onClose={() => setForm(null)} />
        ) : null)}
      {form &&
        form !== 'create' &&
        (storeId ? (
          <FormModal
            storeId={storeId}
            initial={form as Customer}
            onSuccess={onSuccess}
            onClose={() => setForm(null)}
          />
        ) : null)}
      {deleting &&
        (storeId ? (
          <DeleteConfirm
            customer={deleting}
            storeId={storeId}
            onSuccess={onSuccess}
            onClose={() => setDeleting(null)}
          />
        ) : null)}
      {detail && storeId && (
        <DetailDrawer
          customer={detail}
          storeId={storeId}
          onClose={() => setDetail(null)}
          onEdit={() => {
            setForm(detail);
            setDetail(null);
          }}
          onDelete={() => {
            setDeleting(detail);
            setDetail(null);
          }}
          canEdit={can('sales.create')}
          canDelete={can('sales.create')}
        />
      )}
    </div>
  );
}
