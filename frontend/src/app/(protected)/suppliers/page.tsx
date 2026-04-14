'use client';

import { useEffect, useState, useCallback } from 'react';
import { Users, Truck, Plus, Pencil, Trash2, Loader2, X, Search } from 'lucide-react';
import { usePermission } from '@/hooks/usePermission';
import { suppliersApi } from '@/lib/api/store-apis';
import type { Supplier } from '@/types';
import { formatDate } from '@/lib/utils';
import { ApiError } from '@/lib/api/client';

type SupplierListResponse = { data?: Supplier[] };

export default function SuppliersPage() {
  const { can } = usePermission();
  const [suppliers, setSuppliers] = useState<Supplier[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState('');
  const [showModal, setShowModal] = useState(false);
  const [editing, setEditing] = useState<Supplier | null>(null);
  const [form, setForm] = useState({
    name: '',
    contact_name: '',
    phone: '',
    email: '',
    address: '',
  });
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');

  const load = useCallback(() => {
    setLoading(true);
    suppliersApi
      .list({ per_page: 100, search: search || undefined })
      .then(r => setSuppliers((r.data as SupplierListResponse).data ?? []))
      .catch(console.error)
      .finally(() => setLoading(false));
  }, [search]);
  useEffect(() => {
    load();
  }, [load]);

  const openCreate = () => {
    setEditing(null);
    setForm({ name: '', contact_name: '', phone: '', email: '', address: '' });
    setError('');
    setShowModal(true);
  };
  const openEdit = (s: Supplier) => {
    setEditing(s);
    setForm({
      name: s.name,
      contact_name: s.contact_name,
      phone: s.phone,
      email: s.email,
      address: s.address,
    });
    setError('');
    setShowModal(true);
  };

  const handleSave = async () => {
    setSaving(true);
    setError('');
    try {
      if (editing) await suppliersApi.update(editing.id, form);
      else await suppliersApi.create(form);
      setShowModal(false);
      load();
    } catch (e) {
      setError(e instanceof ApiError ? e.message : 'Gagal menyimpan');
    } finally {
      setSaving(false);
    }
  };
  const handleDelete = async (id: string) => {
    if (!confirm('Hapus supplier ini?')) return;
    await suppliersApi.delete(id).catch(console.error);
    load();
  };

  const f =
    (k: keyof typeof form) => (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) =>
      setForm(d => ({ ...d, [k]: e.target.value }));

  return (
    <div className="w-full p-6">
      <style>{`
        @keyframes fadeInScale {
          from { opacity: 0; transform: scale(0.98) translateY(12px); }
          to { opacity: 1; transform: scale(1) translateY(0); }
        }
        .reveal-animate {
          animation: fadeInScale 0.5s cubic-bezier(0.165, 0.84, 0.44, 1) both;
        }
      `}</style>
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
            <Truck size={22} style={{ color: 'var(--accent-em)' }} />
            Supplier
          </h1>
          <p className="page-subtitle">Kelola daftar pemasok dan vendor</p>
        </div>
        {can('suppliers.create') && (
          <button className="btn btn-primary" onClick={openCreate}>
            <Plus size={15} /> Tambah Supplier
          </button>
        )}
      </div>

      <div
        className="reveal-animate"
        style={{ position: 'relative', maxWidth: 360, marginBottom: 16, animationDelay: '0.1s' }}
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
          placeholder="Cari supplier..."
          value={search}
          onChange={e => setSearch(e.target.value)}
        />
      </div>

      <div className="card reveal-animate" style={{ overflow: 'hidden', animationDelay: '0.15s' }}>
        {loading ? (
          <div style={{ display: 'flex', justifyContent: 'center', padding: 40 }}>
            <Loader2 size={24} className="loading-spin" style={{ color: 'var(--accent-em)' }} />
          </div>
        ) : suppliers.length === 0 ? (
          <div className="empty-state">
            <Users size={32} />
            <p>Belum ada supplier</p>
          </div>
        ) : (
          <table className="tbl">
            <thead>
              <tr>
                <th>Nama</th>
                <th>Kontak</th>
                <th>Telepon</th>
                <th>Email</th>
                <th>Status</th>
                <th>Dibuat</th>
                <th />
              </tr>
            </thead>
            <tbody>
              {suppliers.map((s, i) => (
                <tr
                  key={s.id}
                  className="reveal-animate"
                  style={{ animationDelay: `${0.2 + i * 0.02}s` }}
                >
                  <td style={{ fontWeight: 600 }}>{s.name}</td>
                  <td style={{ color: 'var(--text-2)' }}>{s.contact_name || '–'}</td>
                  <td
                    style={{ color: 'var(--text-2)', fontFamily: 'monospace', fontSize: '0.82rem' }}
                  >
                    {s.phone || '–'}
                  </td>
                  <td style={{ color: 'var(--text-2)', fontSize: '0.82rem' }}>{s.email || '–'}</td>
                  <td>
                    <span className={`badge ${s.is_active ? 'badge-green' : 'badge-gray'}`}>
                      {s.is_active ? 'Aktif' : 'Nonaktif'}
                    </span>
                  </td>
                  <td style={{ color: 'var(--text-3)', fontSize: '0.8rem' }}>
                    {formatDate(s.created_at)}
                  </td>
                  <td>
                    <div style={{ display: 'flex', gap: 4 }}>
                      {can('suppliers.update') && (
                        <button className="btn btn-ghost btn-sm" onClick={() => openEdit(s)}>
                          <Pencil size={13} />
                        </button>
                      )}
                      {can('suppliers.delete') && (
                        <button
                          className="btn btn-ghost btn-sm"
                          style={{ color: 'var(--accent-rd)' }}
                          onClick={() => handleDelete(s.id)}
                        >
                          <Trash2 size={13} />
                        </button>
                      )}
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      {showModal && (
        <div className="modal-overlay" onClick={() => setShowModal(false)}>
          <div className="modal-box" style={{ maxWidth: 420 }} onClick={e => e.stopPropagation()}>
            <div
              style={{
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
                marginBottom: 18,
              }}
            >
              <h2 style={{ fontWeight: 700 }}>{editing ? 'Edit Supplier' : 'Tambah Supplier'}</h2>
              <button className="btn btn-ghost btn-sm" onClick={() => setShowModal(false)}>
                <X size={15} />
              </button>
            </div>
            {error && (
              <div
                style={{
                  background: 'rgba(239,68,68,0.12)',
                  borderRadius: 8,
                  padding: '8px 12px',
                  color: '#f87171',
                  fontSize: '0.83rem',
                  marginBottom: 14,
                  border: '1px solid rgba(239,68,68,0.3)',
                }}
              >
                {error}
              </div>
            )}
            <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
              {[
                ['name', 'Nama Supplier'],
                ['contact_name', 'Nama Kontak'],
                ['phone', 'Telepon'],
                ['email', 'Email'],
                ['address', 'Alamat'],
              ].map(([k, l]) => (
                <div key={k} className="input-group">
                  <label className="input-label">{l}</label>
                  <input
                    className="input"
                    type={k === 'email' ? 'email' : 'text'}
                    value={form[k as keyof typeof form]}
                    onChange={f(k as keyof typeof form)}
                  />
                </div>
              ))}
            </div>
            <div style={{ display: 'flex', gap: 8, marginTop: 20 }}>
              <button
                className="btn btn-secondary"
                style={{ flex: 1 }}
                onClick={() => setShowModal(false)}
              >
                Batal
              </button>
              <button
                className="btn btn-primary"
                style={{ flex: 1 }}
                disabled={saving}
                onClick={handleSave}
              >
                {saving ? <Loader2 size={15} className="loading-spin" /> : null}
                {saving ? 'Menyimpan...' : 'Simpan'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
