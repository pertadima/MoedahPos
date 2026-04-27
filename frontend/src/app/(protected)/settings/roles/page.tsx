'use client';

import { useState, useEffect, useCallback, useMemo } from 'react';
import { ShieldCheck, Plus, Loader2, X, Edit3, Trash2, Check } from 'lucide-react';
import Portal from '@/components/ui/Portal';
import { rolesApi } from '@/lib/api/store-apis';
import { ApiError } from '@/lib/api/client';
import type { Role, Permission } from '@/types';

// ── Translations ─────────────────────────────────────────────────────────────

const MODULE_LABELS: Record<string, string> = {
  penjualan: 'Riwayat Penjualan',
  kasir: 'Kasir / POS',
  inventory: 'Stok & Inventori',
  pembelian: 'Pembelian & Supplier',
  keuangan: 'Keuangan & Biaya',
  settings: 'Pengaturan Sistem',
};

const PERM_LABELS: Record<string, string> = {
  'penjualan:read': 'Lihat riwayat penjualan',
  'penjualan:write': 'Buat catatan penjualan',
  'penjualan:update': 'Ubah catatan penjualan',
  'penjualan:delete': 'Hapus/Batalkan penjualan',
  'kasir:read': 'Akses antarmuka kasir',
  'kasir:write': 'Proses transaksi baru',
  'kasir:update': 'Ubah keranjang aktif',
  'kasir:delete': 'Kosongkan keranjang aktif',
  'inventory:read': 'Lihat stok dan produk',
  'inventory:write': 'Tambah produk/stok baru',
  'inventory:update': 'Ubah detail produk',
  'inventory:delete': 'Hapus produk',
  'pembelian:read': 'Lihat pesanan pembelian',
  'pembelian:write': 'Buat pesanan pembelian',
  'pembelian:update': 'Ubah pesanan pembelian',
  'pembelian:delete': 'Hapus pesanan pembelian',
  'keuangan:read': 'Lihat catatan keuangan',
  'keuangan:write': 'Tambah pemasukan/pengeluaran',
  'keuangan:update': 'Ubah catatan keuangan',
  'keuangan:delete': 'Hapus catatan keuangan',
  'settings:read': 'Lihat pengaturan dan pengguna',
  'settings:write': 'Tambah pengguna/peran',
  'settings:update': 'Ubah pengaturan/pengguna',
  'settings:delete': 'Hapus pengguna/peran',
};

const ROLE_DESC_LABELS: Record<string, string> = {
  'Full system access across all stores': 'Akses sistem penuh di semua toko',
  'Full access within assigned stores': 'Akses penuh dalam toko yang ditugaskan',
  'Process transactions only': 'Hanya memproses transaksi',
  'Manage products, stock, and reports': 'Kelola produk, stok, dan laporan',
  'View-only access': 'Akses lihat saja',
};

// ── Permissions Checkbox Matrix Modal ────────────────────────────────────────

function RoleFormDrawer({
  mode,
  role,
  allPermissions,
  onSuccess,
  onClose,
}: {
  mode: 'create' | 'edit';
  role?: Role;
  allPermissions: Permission[];
  onSuccess: () => void;
  onClose: () => void;
}) {
  const [form, setForm] = useState({
    name: role?.name ?? '',
    description: role?.description ?? '',
  });
  // Map of selected permission IDs
  const [selectedPerms, setSelectedPerms] = useState<Set<string>>(
    new Set(role?.permissions?.map(p => p.id) ?? [])
  );
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  // Group permissions by module
  const permsByModule = useMemo(() => {
    const grouped: Record<string, Permission[]> = {};
    for (const p of allPermissions) {
      const mod = p.name.split(':')[0] || 'Lainnya';
      if (!grouped[mod]) grouped[mod] = [];
      grouped[mod].push(p);
    }
    return grouped;
  }, [allPermissions]);

  const togglePermission = (id: string) => {
    const next = new Set(selectedPerms);
    if (next.has(id)) next.delete(id);
    else next.add(id);
    setSelectedPerms(next);
  };

  const toggleModule = (module: string) => {
    const modPerms = permsByModule[module] ?? [];
    const allSelected = modPerms.every(p => selectedPerms.has(p.id));
    const next = new Set(selectedPerms);

    if (allSelected) {
      modPerms.forEach(p => next.delete(p.id));
    } else {
      modPerms.forEach(p => next.add(p.id));
    }
    setSelectedPerms(next);
  };

  const handleSubmit = async () => {
    setError('');
    if (!form.name.trim()) {
      setError('Nama peran wajib diisi');
      return;
    }

    setLoading(true);
    try {
      const payload = {
        name: form.name,
        description: form.description,
        permission_ids: Array.from(selectedPerms),
      };

      if (mode === 'create') {
        await rolesApi.create(payload);
      } else if (role) {
        await rolesApi.update(role.id, payload);
      }
      onSuccess();
    } catch (e) {
      setError(e instanceof ApiError ? e.message : 'Gagal menyimpan');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Portal>
      <div
        style={{
          position: 'fixed',
          inset: 0,
          zIndex: 5000,
          display: 'flex',
          justifyContent: 'flex-end',
        }}
        onClick={onClose}
      >
        <style>{`
          @keyframes slideInRight {
            from { transform: translateX(100%); }
            to { transform: translateX(0); }
          }
        `}</style>
        {/* Backdrop */}
        <div
          style={{
            position: 'absolute',
            inset: 0,
            background: 'rgba(0,0,0,0.4)',
            backdropFilter: 'blur(3px)',
          }}
        />
        {/* Drawer Content */}
        <div
          className="card"
          style={{
            position: 'relative',
            width: '100%',
            maxWidth: 580,
            height: '100%',
            borderRadius: 0,
            padding: 0,
            display: 'flex',
            flexDirection: 'column',
            boxShadow: '-8px 0 32px rgba(0,0,0,0.15)',
            animation: 'slideInRight 0.3s cubic-bezier(0.16, 1, 0.3, 1)',
          }}
          onClick={e => e.stopPropagation()}
        >
          {/* Header */}
          <div
            style={{
              padding: '20px 24px',
              borderBottom: '1px solid var(--border)',
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
              background: 'var(--bg-card)',
            }}
          >
            <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
              <ShieldCheck size={22} style={{ color: 'var(--accent-em)' }} />
              <h2 style={{ fontWeight: 800, fontSize: '1.2rem', margin: 0 }}>
                {mode === 'create' ? 'Tambah Peran' : 'Edit Peran'}
              </h2>
            </div>
            <button onClick={onClose} className="btn btn-ghost btn-sm" style={{ padding: 6 }}>
              <X size={20} />
            </button>
          </div>

          {/* Body */}
          <div style={{ flex: 1, overflowY: 'auto', padding: '24px' }}>
            {error && (
              <div
                style={{
                  background: 'rgba(239,68,68,0.12)',
                  border: '1px solid rgba(239,68,68,0.3)',
                  borderRadius: 10,
                  padding: '12px 16px',
                  color: '#f87171',
                  fontSize: '0.85rem',
                  marginBottom: 20,
                }}
              >
                {error}
              </div>
            )}

            <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
              <div>
                <label className="label">Nama Peran</label>
                <input
                  className="input"
                  value={form.name}
                  onChange={e => setForm(f => ({ ...f, name: e.target.value }))}
                  placeholder="Misal: Manager, Kasir"
                />
              </div>
              <div>
                <label className="label">Deskripsi</label>
                <textarea
                  className="input"
                  value={form.description}
                  onChange={e => setForm(f => ({ ...f, description: e.target.value }))}
                  placeholder="Deskripsi singkat mengenai peran ini"
                  style={{ resize: 'vertical', minHeight: 80 }}
                />
              </div>

              <div style={{ marginTop: 10 }}>
                <label className="label" style={{ marginBottom: 12 }}>
                  Hak Akses (Permissions)
                </label>
                <div
                  style={{
                    border: '1px solid var(--border)',
                    borderRadius: 12,
                    background: 'var(--bg-elevated)',
                    overflow: 'hidden',
                  }}
                >
                  {Object.entries(permsByModule).map(([mod, perms]) => {
                    const allSelected = perms.every(p => selectedPerms.has(p.id));
                    const someSelected = perms.some(p => selectedPerms.has(p.id));

                    return (
                      <div key={mod} style={{ borderBottom: '1px solid var(--border)' }}>
                        <div
                          style={{
                            padding: '12px 16px',
                            background: 'rgba(0,0,0,0.02)',
                            display: 'flex',
                            alignItems: 'center',
                            gap: 10,
                            cursor: 'pointer',
                            transition: 'background 0.2s',
                          }}
                          className="module-header"
                          onClick={() => toggleModule(mod)}
                        >
                          <div
                            className="checkbox-container"
                            style={{
                              display: 'flex',
                              alignItems: 'center',
                              justifyContent: 'center',
                              width: 18,
                              height: 18,
                              borderRadius: 4,
                              border: `1px solid ${allSelected || someSelected ? 'var(--primary)' : 'var(--border-hover)'}`,
                              background: allSelected
                                ? 'var(--primary)'
                                : someSelected
                                  ? 'rgba(99,102,241,0.15)'
                                  : 'var(--bg-card)',
                              transition: 'all 0.2s',
                            }}
                          >
                            {allSelected && <Check size={12} color="#fff" strokeWidth={3} />}
                            {!allSelected && someSelected && (
                              <div
                                style={{
                                  width: 8,
                                  height: 2,
                                  background: 'var(--primary)',
                                  borderRadius: 1,
                                }}
                              />
                            )}
                          </div>
                          <span style={{ fontWeight: 700, fontSize: '0.88rem' }}>
                            {MODULE_LABELS[mod] || mod}
                          </span>
                        </div>
                        <div
                          style={{
                            padding: '12px 16px 16px 44px',
                            display: 'grid',
                            gridTemplateColumns: '1fr',
                            gap: 12,
                          }}
                        >
                          {perms.map(p => {
                            const isSelected = selectedPerms.has(p.id);
                            return (
                              <div
                                key={p.id}
                                className="perm-item"
                                style={{
                                  display: 'flex',
                                  alignItems: 'center',
                                  gap: 10,
                                  cursor: 'pointer',
                                }}
                                onClick={() => togglePermission(p.id)}
                              >
                                <div
                                  style={{
                                    width: 18,
                                    height: 18,
                                    borderRadius: 4,
                                    border: `1px solid ${isSelected ? 'var(--primary)' : 'var(--border-hover)'}`,
                                    background: isSelected ? 'var(--primary)' : 'var(--bg-card)',
                                    display: 'flex',
                                    alignItems: 'center',
                                    justifyContent: 'center',
                                    flexShrink: 0,
                                    transition: 'all 0.2s',
                                  }}
                                >
                                  {isSelected && <Check size={12} color="#fff" strokeWidth={3} />}
                                </div>
                                <span
                                  style={{
                                    fontSize: '0.85rem',
                                    color: isSelected ? 'var(--text-1)' : 'var(--text-2)',
                                    fontWeight: isSelected ? 600 : 400,
                                  }}
                                >
                                  {PERM_LABELS[p.name] || p.description || p.name}
                                </span>
                              </div>
                            );
                          })}
                        </div>
                      </div>
                    );
                  })}
                </div>
              </div>
            </div>
          </div>

          {/* Footer */}
          <div
            style={{
              padding: '20px 24px',
              borderTop: '1px solid var(--border)',
              display: 'flex',
              gap: 12,
              background: 'var(--bg-card)',
            }}
          >
            <button
              className="btn btn-secondary"
              style={{ flex: 1, height: 44 }}
              onClick={onClose}
              disabled={loading}
            >
              Batal
            </button>
            <button
              className="btn btn-primary"
              style={{ flex: 1, height: 44 }}
              onClick={handleSubmit}
              disabled={loading}
            >
              {loading ? <Loader2 size={16} className="loading-spin" /> : null}
              {loading ? 'Menyimpan...' : mode === 'create' ? 'Buat Peran' : 'Simpan Peran'}
            </button>
          </div>
        </div>
      </div>
    </Portal>
  );
}

// ── Main Page ─────────────────────────────────────────────────────────────────

export default function RolesPage() {
  const [roles, setRoles] = useState<Role[]>([]);
  const [permissions, setPermissions] = useState<Permission[]>([]);
  const [loading, setLoading] = useState(true);
  const [modalMode, setModalMode] = useState<'create' | 'edit' | null>(null);
  const [editingRole, setEditingRole] = useState<Role | undefined>();

  const loadData = useCallback(async () => {
    setLoading(true);
    try {
      const [rolesRes, permsRes] = await Promise.all([rolesApi.list(), rolesApi.listPermissions()]);
      setRoles(rolesRes.data ?? []);
      setPermissions(permsRes.data ?? []);
    } catch (e) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadData();
  }, [loadData]);

  const handleDelete = async (role: Role) => {
    if (role.name === 'superadmin') {
      alert('Role superadmin tidak dapat dihapus');
      return;
    }
    if (!confirm(`Hapus peran ${role.name}?`)) return;
    try {
      await rolesApi.delete(role.id);
      loadData();
    } catch (e) {
      alert(e instanceof ApiError ? e.message : 'Gagal menghapus');
    }
  };

  return (
    <div className="w-full p-6">
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
            <ShieldCheck size={22} style={{ color: 'var(--accent-em)' }} />
            Peran & Hak Akses
          </h1>
          <p className="page-subtitle">Kelola akses modular untuk sistem POS Anda.</p>
        </div>
        <button
          className="btn btn-primary"
          onClick={() => {
            setEditingRole(undefined);
            setModalMode('create');
          }}
        >
          <Plus size={16} /> Tambah Peran
        </button>
      </div>

      <div className="card reveal-animate" style={{ padding: 0, animationDelay: '0.1s' }}>
        {loading ? (
          <div style={{ textAlign: 'center', padding: '40px 0', color: 'var(--text-3)' }}>
            <Loader2 size={24} className="loading-spin mx-auto mb-2" />
            <p>Memuat peran...</p>
          </div>
        ) : (
          <div className="tbl-container">
            <table className="tbl">
              <thead>
                <tr>
                  <th style={{ width: '25%' }}>Nama Peran</th>
                  <th style={{ width: '35%' }}>Deskripsi</th>
                  <th style={{ width: '20%' }}>Akses Terpilih</th>
                  <th style={{ width: '20%', textAlign: 'right' }}>Aksi</th>
                </tr>
              </thead>
              <tbody>
                {roles.length === 0 ? (
                  <tr>
                    <td colSpan={4} className="text-center py-6 text-gray-500">
                      Belum ada peran
                    </td>
                  </tr>
                ) : (
                  roles.map(role => (
                    <tr key={role.id}>
                      <td>
                        <div style={{ fontWeight: 600 }}>{role.name}</div>
                      </td>
                      <td>
                        <span style={{ fontSize: '0.85rem', color: 'var(--text-2)' }}>
                          {ROLE_DESC_LABELS[role.description] || role.description || '-'}
                        </span>
                      </td>
                      <td>
                        <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
                          <span
                            style={{
                              background: 'rgba(16,185,129,0.1)',
                              color: '#10b981',
                              padding: '2px 8px',
                              borderRadius: 12,
                              fontSize: '0.75rem',
                              fontWeight: 700,
                            }}
                          >
                            {role.permissions?.length ?? 0}
                          </span>
                          <span style={{ fontSize: '0.8rem', color: 'var(--text-3)' }}>Izin</span>
                        </div>
                      </td>
                      <td style={{ textAlign: 'right' }}>
                        <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 6 }}>
                          <button
                            className="btn btn-secondary btn-sm"
                            onClick={() => {
                              setEditingRole(role);
                              setModalMode('edit');
                            }}
                            disabled={role.name === 'superadmin'}
                          >
                            <Edit3 size={14} />
                          </button>
                          <button
                            className="btn btn-ghost btn-sm"
                            style={{ color: '#ef4444' }}
                            onClick={() => handleDelete(role)}
                            disabled={role.name === 'superadmin'}
                          >
                            <Trash2 size={14} />
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {modalMode && (
        <RoleFormDrawer
          mode={modalMode}
          role={editingRole}
          allPermissions={permissions}
          onSuccess={() => {
            setModalMode(null);
            loadData();
          }}
          onClose={() => setModalMode(null)}
        />
      )}
    </div>
  );
}
