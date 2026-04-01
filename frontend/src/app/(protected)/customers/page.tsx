'use client';

import { useState, useEffect, useCallback } from 'react';
import {
  Users, Plus, Loader2, X, Search, Phone, Mail, MapPin,
  FileText, Edit3, Trash2, User, ChevronRight,
} from 'lucide-react';
import { useAuth } from '@/lib/auth/AuthContext';
import { customersApi } from '@/lib/api/store-apis';
import { formatDate } from '@/lib/utils';
import type { Customer } from '@/types';
import { ApiError } from '@/lib/api/client';

// ── Empty form ────────────────────────────────────────────────────────────────
const EMPTY = { name: '', phone: '', email: '', address: '', notes: '' };

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
    name:    initial?.name    ?? '',
    phone:   initial?.phone   ?? '',
    email:   initial?.email   ?? '',
    address: initial?.address ?? '',
    notes:   initial?.notes   ?? '',
  });
  const [saving, setSaving] = useState(false);
  const [error, setError]   = useState('');

  const set = (k: keyof typeof form) => (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) =>
    setForm(f => ({ ...f, [k]: e.target.value }));

  const handleSave = async () => {
    if (!form.name.trim()) { setError('Nama wajib diisi'); return; }
    setSaving(true); setError('');
    try {
      if (isEdit) {
        await customersApi.update(storeId, initial!.id, form);
      } else {
        await customersApi.create(storeId, form);
      }
      onSuccess();
    } catch (e) {
      setError(e instanceof ApiError ? e.message : 'Gagal menyimpan');
    } finally { setSaving(false); }
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-box" style={{ maxWidth: 460 }} onClick={e => e.stopPropagation()}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 18 }}>
          <h2 style={{ fontWeight: 800 }}>{isEdit ? 'Edit Customer' : 'Tambah Customer'}</h2>
          <button className="btn btn-ghost btn-sm" onClick={onClose}><X size={15} /></button>
        </div>

        {error && (
          <div style={{ background: 'rgba(239,68,68,0.1)', border: '1px solid rgba(239,68,68,0.3)', borderRadius: 8, padding: '8px 12px', color: '#f87171', fontSize: '0.83rem', marginBottom: 14 }}>
            {error}
          </div>
        )}

        <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
          <div className="input-group">
            <label className="input-label">Nama <span style={{ color: '#ef4444' }}>*</span></label>
            <input className="input" value={form.name} onChange={set('name')} placeholder="Nama lengkap customer" autoFocus />
          </div>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 10 }}>
            <div className="input-group">
              <label className="input-label">Telepon</label>
              <input className="input" value={form.phone} onChange={set('phone')} placeholder="08xx-xxxx-xxxx" />
            </div>
            <div className="input-group">
              <label className="input-label">Email</label>
              <input type="email" className="input" value={form.email} onChange={set('email')} placeholder="email@contoh.com" />
            </div>
          </div>
          <div className="input-group">
            <label className="input-label">Alamat</label>
            <input className="input" value={form.address} onChange={set('address')} placeholder="Alamat pengiriman" />
          </div>
          <div className="input-group">
            <label className="input-label">Catatan</label>
            <textarea className="input" value={form.notes} onChange={set('notes')} rows={2} placeholder="Preferensi, alergi, dll..." style={{ resize: 'vertical' }} />
          </div>
        </div>

        <div style={{ display: 'flex', gap: 8, marginTop: 18 }}>
          <button className="btn btn-secondary" style={{ flex: 1 }} onClick={onClose}>Batal</button>
          <button className="btn btn-primary" style={{ flex: 1 }} onClick={handleSave} disabled={saving}>
            {saving ? <Loader2 size={14} className="loading-spin" /> : null}
            {saving ? 'Menyimpan...' : isEdit ? 'Simpan Perubahan' : 'Tambah Customer'}
          </button>
        </div>
      </div>
    </div>
  );
}

// ── Delete Confirm ────────────────────────────────────────────────────────────
function DeleteConfirm({ customer, storeId, onSuccess, onClose }: { customer: Customer; storeId: string; onSuccess: () => void; onClose: () => void }) {
  const [loading, setLoading] = useState(false);
  const handleDelete = async () => {
    setLoading(true);
    try { await customersApi.delete(storeId, customer.id); onSuccess(); }
    catch (e) { alert(e instanceof ApiError ? e.message : 'Gagal menghapus'); }
    finally { setLoading(false); }
  };
  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-box" style={{ maxWidth: 380 }} onClick={e => e.stopPropagation()}>
        <div style={{ textAlign: 'center', padding: '8px 0 16px' }}>
          <Trash2 size={28} style={{ color: '#ef4444', marginBottom: 12 }} />
          <h2 style={{ fontWeight: 800, marginBottom: 8 }}>Hapus Customer?</h2>
          <p style={{ color: 'var(--text-2)', fontSize: '0.875rem', lineHeight: 1.6 }}>
            <strong>{customer.name}</strong> akan dihapus dari daftar customer. Data transaksi tetap tersimpan.
          </p>
        </div>
        <div style={{ display: 'flex', gap: 8 }}>
          <button className="btn btn-secondary" style={{ flex: 1 }} onClick={onClose} disabled={loading}>Batal</button>
          <button className="btn btn-danger" style={{ flex: 1 }} onClick={handleDelete} disabled={loading}>
            {loading ? <Loader2 size={14} className="loading-spin" /> : <Trash2 size={14} />}
            {loading ? 'Menghapus...' : 'Ya, Hapus'}
          </button>
        </div>
      </div>
    </div>
  );
}

// ── Detail Drawer ─────────────────────────────────────────────────────────────
function DetailDrawer({ customer, onClose, onEdit, onDelete }: { customer: Customer; onClose: () => void; onEdit: () => void; onDelete: () => void }) {
  const avatar = customer.name.charAt(0).toUpperCase();
  return (
    <>
      <div onClick={onClose} style={{ position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.5)', backdropFilter: 'blur(2px)', zIndex: 200 }} />
      <div style={{ position: 'fixed', top: 0, right: 0, bottom: 0, width: 400, background: 'var(--bg-card)', borderLeft: '1px solid var(--border)', zIndex: 201, overflowY: 'auto', display: 'flex', flexDirection: 'column' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '18px 20px', borderBottom: '1px solid var(--border)' }}>
          <div style={{ fontWeight: 800 }}>Detail Customer</div>
          <div style={{ display: 'flex', gap: 6 }}>
            <button className="btn btn-secondary btn-sm" onClick={onEdit}><Edit3 size={13} /> Edit</button>
            <button className="btn btn-danger btn-sm" onClick={onDelete}><Trash2 size={13} /></button>
            <button className="btn btn-ghost btn-sm" onClick={onClose}><X size={15} /></button>
          </div>
        </div>
        <div style={{ padding: 20, display: 'flex', flexDirection: 'column', gap: 18 }}>
          {/* Avatar + name */}
          <div style={{ display: 'flex', alignItems: 'center', gap: 14 }}>
            <div style={{ width: 52, height: 52, borderRadius: '50%', background: 'linear-gradient(135deg, var(--accent-in), var(--accent-em))', display: 'flex', alignItems: 'center', justifyContent: 'center', fontWeight: 800, fontSize: '1.2rem', color: '#fff', flexShrink: 0 }}>
              {avatar}
            </div>
            <div>
              <div style={{ fontWeight: 800, fontSize: '1rem' }}>{customer.name}</div>
              <div style={{ fontSize: '0.75rem', color: 'var(--text-3)', marginTop: 2 }}>Customer sejak {formatDate(customer.created_at)}</div>
            </div>
          </div>

          {/* Details */}
          <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
            {customer.phone && (
              <div style={{ display: 'flex', alignItems: 'center', gap: 10, background: 'var(--bg-elevated)', borderRadius: 10, padding: '10px 14px' }}>
                <Phone size={14} style={{ color: 'var(--accent-em)', flexShrink: 0 }} />
                <div>
                  <div style={{ fontSize: '0.68rem', color: 'var(--text-3)' }}>Telepon</div>
                  <div style={{ fontWeight: 600 }}>{customer.phone}</div>
                </div>
              </div>
            )}
            {customer.email && (
              <div style={{ display: 'flex', alignItems: 'center', gap: 10, background: 'var(--bg-elevated)', borderRadius: 10, padding: '10px 14px' }}>
                <Mail size={14} style={{ color: 'var(--accent-em)', flexShrink: 0 }} />
                <div>
                  <div style={{ fontSize: '0.68rem', color: 'var(--text-3)' }}>Email</div>
                  <div style={{ fontWeight: 600 }}>{customer.email}</div>
                </div>
              </div>
            )}
            {customer.address && (
              <div style={{ display: 'flex', alignItems: 'flex-start', gap: 10, background: 'var(--bg-elevated)', borderRadius: 10, padding: '10px 14px' }}>
                <MapPin size={14} style={{ color: 'var(--accent-em)', flexShrink: 0, marginTop: 3 }} />
                <div>
                  <div style={{ fontSize: '0.68rem', color: 'var(--text-3)' }}>Alamat</div>
                  <div style={{ fontWeight: 600 }}>{customer.address}</div>
                </div>
              </div>
            )}
            {customer.notes && (
              <div style={{ display: 'flex', alignItems: 'flex-start', gap: 10, background: 'var(--bg-elevated)', borderRadius: 10, padding: '10px 14px' }}>
                <FileText size={14} style={{ color: 'var(--accent-em)', flexShrink: 0, marginTop: 3 }} />
                <div>
                  <div style={{ fontSize: '0.68rem', color: 'var(--text-3)' }}>Catatan</div>
                  <div style={{ fontSize: '0.85rem' }}>{customer.notes}</div>
                </div>
              </div>
            )}
            {!customer.phone && !customer.email && !customer.address && !customer.notes && (
              <div style={{ textAlign: 'center', color: 'var(--text-3)', padding: '16px 0', fontSize: '0.85rem' }}>Tidak ada info tambahan</div>
            )}
          </div>
        </div>
      </div>
    </>
  );
}

// ── Main Page ─────────────────────────────────────────────────────────────────
export default function CustomersPage() {
  const { selectedStore } = useAuth();
  const [customers, setCustomers] = useState<Customer[]>([]);
  const [total, setTotal]         = useState(0);
  const [page, setPage]           = useState(1);
  const [loading, setLoading]     = useState(true);
  const [search, setSearch]       = useState('');
  const [form, setForm]           = useState<'create' | Customer | null>(null);
  const [detail, setDetail]       = useState<Customer | null>(null);
  const [deleting, setDeleting]   = useState<Customer | null>(null);

  const storeId = selectedStore?.store_id;
  const PER_PAGE = 20;

  const load = useCallback(() => {
    if (!storeId) return;
    setLoading(true);
    customersApi.list(storeId, { page, per_page: PER_PAGE, search: search || undefined })
      .then(r => {
        const d = r.data as any;
        setCustomers(d.data ?? []);
        setTotal(d.meta?.total ?? 0);
      })
      .catch(console.error)
      .finally(() => setLoading(false));
  }, [storeId, page, search]);

  useEffect(() => { load(); }, [load]);

  // Debounce search
  useEffect(() => { setPage(1); }, [search]);

  const onSuccess = () => {
    setForm(null); setDeleting(null); setDetail(null); load();
  };

  if (!selectedStore) return (
    <div style={{ padding: 32 }}>
      <div className="empty-state card" style={{ padding: 40 }}><Users size={40} /><p>Pilih toko terlebih dahulu</p></div>
    </div>
  );

  return (
    <div style={{ padding: 24 }}>
      {/* Header */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: 20 }}>
        <div>
          <h1 className="page-title">Customer</h1>
          <p className="page-subtitle">{total} customer terdaftar · {selectedStore.store_name}</p>
        </div>
        <button className="btn btn-primary" onClick={() => setForm('create')}>
          <Plus size={15} /> Tambah Customer
        </button>
      </div>

      {/* Search */}
      <div style={{ position: 'relative', marginBottom: 16, maxWidth: 360 }}>
        <Search size={15} style={{ position: 'absolute', left: 12, top: '50%', transform: 'translateY(-50%)', color: 'var(--text-3)' }} />
        <input
          className="input"
          style={{ paddingLeft: 36 }}
          placeholder="Cari nama atau telepon..."
          value={search}
          onChange={e => setSearch(e.target.value)}
        />
      </div>

      {/* Table */}
      <div className="card" style={{ overflow: 'hidden' }}>
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
            <table className="tbl">
              <thead>
                <tr>
                  <th>Nama</th>
                  <th>Telepon</th>
                  <th>Email</th>
                  <th>Alamat</th>
                  <th>Terdaftar</th>
                  <th>Aksi</th>
                </tr>
              </thead>
              <tbody>
                {customers.map(c => (
                  <tr key={c.id} style={{ cursor: 'pointer' }} onClick={() => setDetail(c)}>
                    <td>
                      <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                        <div style={{ width: 32, height: 32, borderRadius: '50%', background: 'linear-gradient(135deg, var(--accent-in), var(--accent-em))', display: 'flex', alignItems: 'center', justifyContent: 'center', fontWeight: 700, fontSize: '0.8rem', color: '#fff', flexShrink: 0 }}>
                          {c.name.charAt(0).toUpperCase()}
                        </div>
                        <span style={{ fontWeight: 600 }}>{c.name}</span>
                      </div>
                    </td>
                    <td style={{ color: 'var(--text-2)', fontSize: '0.85rem' }}>
                      {c.phone ? <span style={{ display: 'flex', alignItems: 'center', gap: 4 }}><Phone size={12} />{c.phone}</span> : '—'}
                    </td>
                    <td style={{ color: 'var(--text-2)', fontSize: '0.85rem' }}>{c.email ?? '—'}</td>
                    <td style={{ color: 'var(--text-3)', fontSize: '0.82rem', maxWidth: 180, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                      {c.address ?? '—'}
                    </td>
                    <td style={{ color: 'var(--text-3)', fontSize: '0.8rem' }}>{formatDate(c.created_at)}</td>
                    <td>
                      <div style={{ display: 'flex', gap: 4 }}>
                        <button className="btn btn-ghost btn-sm" onClick={e => { e.stopPropagation(); setForm(c); }} title="Edit">
                          <Edit3 size={13} />
                        </button>
                        <button className="btn btn-ghost btn-sm" style={{ color: 'var(--accent-rd)' }} onClick={e => { e.stopPropagation(); setDeleting(c); }} title="Hapus">
                          <Trash2 size={13} />
                        </button>
                        <button className="btn btn-ghost btn-sm" onClick={() => setDetail(c)}>
                          <ChevronRight size={14} />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>

            {/* Pagination */}
            {total > PER_PAGE && (
              <div style={{ display: 'flex', justifyContent: 'center', gap: 8, padding: '12px 16px', borderTop: '1px solid var(--border)' }}>
                <button className="btn btn-secondary btn-sm" disabled={page <= 1} onClick={() => setPage(p => p - 1)}>Sebelumnya</button>
                <span style={{ alignSelf: 'center', fontSize: '0.82rem', color: 'var(--text-2)' }}>
                  Halaman {page} dari {Math.ceil(total / PER_PAGE)}
                </span>
                <button className="btn btn-secondary btn-sm" disabled={page >= Math.ceil(total / PER_PAGE)} onClick={() => setPage(p => p + 1)}>Berikutnya</button>
              </div>
            )}
          </>
        )}
      </div>

      {/* Modals & Drawer */}
      {form === 'create' && <FormModal storeId={storeId!} onSuccess={onSuccess} onClose={() => setForm(null)} />}
      {form && form !== 'create' && <FormModal storeId={storeId!} initial={form as Customer} onSuccess={onSuccess} onClose={() => setForm(null)} />}
      {deleting && <DeleteConfirm customer={deleting} storeId={storeId!} onSuccess={onSuccess} onClose={() => setDeleting(null)} />}
      {detail && (
        <DetailDrawer
          customer={detail}
          onClose={() => setDetail(null)}
          onEdit={() => { setForm(detail); setDetail(null); }}
          onDelete={() => { setDeleting(detail); setDetail(null); }}
        />
      )}
    </div>
  );
}
