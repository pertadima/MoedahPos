'use client';

import { useState, useEffect, useCallback, useMemo } from 'react';
import { ShieldCheck, Plus, Loader2, X, Edit3, Trash2, Check } from 'lucide-react';
import Portal from '@/components/ui/Portal';
import { rolesApi } from '@/lib/api/store-apis';
import { ApiError } from '@/lib/api/client';
import type { Role, Permission } from '@/types';

// ── Permissions Checkbox Matrix Modal ────────────────────────────────────────

function RoleFormModal({
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
      if (!grouped[p.module]) grouped[p.module] = [];
      grouped[p.module].push(p);
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
      <div className="modal-overlay" style={{ zIndex: 5000 }} onClick={onClose}>
        <div className="modal-box" style={{ maxWidth: 640 }} onClick={e => e.stopPropagation()}>
          <div
            style={{
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
              marginBottom: 20,
            }}
          >
            <h2
              style={{
                fontWeight: 800,
                fontSize: '1.05rem',
                display: 'flex',
                alignItems: 'center',
                gap: 8,
              }}
            >
              <ShieldCheck size={18} style={{ color: 'var(--primary)' }} />
              {mode === 'create' ? 'Tambah Peran' : 'Edit Peran'}
            </h2>
            <button className="btn btn-ghost btn-sm" onClick={onClose}>
              <X size={16} />
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
            <div style={{ display: 'flex', gap: 12 }}>
              <div style={{ flex: 1 }}>
                <label className="label">Nama Peran</label>
                <input
                  className="input"
                  value={form.name}
                  onChange={e => setForm(f => ({ ...f, name: e.target.value }))}
                  placeholder="Misal: Manager, Kasir"
                />
              </div>
            </div>
            <div>
              <label className="label">Deskripsi</label>
              <textarea
                className="input"
                value={form.description}
                onChange={e => setForm(f => ({ ...f, description: e.target.value }))}
                placeholder="Deskripsi singkat mengenai peran ini"
                style={{ resize: 'vertical', minHeight: 60 }}
              />
            </div>

            <div style={{ marginTop: 8 }}>
              <label className="label">Hak Akses (Permissions)</label>
              <div
                style={{
                  maxHeight: 320,
                  overflowY: 'auto',
                  border: '1px solid var(--border)',
                  borderRadius: 8,
                  background: 'var(--bg-elevated)',
                }}
              >
                {Object.entries(permsByModule).map(([mod, perms]) => {
                  const allSelected = perms.every(p => selectedPerms.has(p.id));
                  const someSelected = perms.some(p => selectedPerms.has(p.id));

                  return (
                    <div key={mod} style={{ borderBottom: '1px solid var(--border)' }}>
                      <div
                        style={{
                          padding: '10px 12px',
                          background: 'rgba(0,0,0,0.02)',
                          display: 'flex',
                          alignItems: 'center',
                          gap: 8,
                          cursor: 'pointer',
                          fontWeight: 700,
                          fontSize: '0.85rem',
                        }}
                        onClick={() => toggleModule(mod)}
                      >
                        <div
                          style={{
                            width: 16,
                            height: 16,
                            borderRadius: 4,
                            border: '1px solid var(--border-hover)',
                            background: allSelected
                              ? 'var(--primary)'
                              : someSelected
                                ? 'var(--text-3)'
                                : 'transparent',
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'center',
                          }}
                        >
                          {(allSelected || someSelected) && <Check size={12} color="#fff" />}
                        </div>
                        <span style={{ textTransform: 'capitalize' }}>{mod}</span>
                      </div>
                      <div
                        style={{
                          padding: '8px 12px 12px 36px',
                          display: 'grid',
                          gridTemplateColumns: '1fr 1fr',
                          gap: 8,
                        }}
                      >
                        {perms.map(p => {
                          const isSelected = selectedPerms.has(p.id);
                          return (
                            <div
                              key={p.id}
                              style={{
                                display: 'flex',
                                alignItems: 'center',
                                gap: 6,
                                cursor: 'pointer',
                              }}
                              onClick={() => togglePermission(p.id)}
                            >
                              <div
                                style={{
                                  width: 14,
                                  height: 14,
                                  borderRadius: 3,
                                  border: '1px solid var(--border-hover)',
                                  background: isSelected ? 'var(--primary)' : 'transparent',
                                  display: 'flex',
                                  alignItems: 'center',
                                  justifyContent: 'center',
                                  flexShrink: 0,
                                }}
                              >
                                {isSelected && <Check size={10} color="#fff" />}
                              </div>
                              <span
                                style={{
                                  fontSize: '0.8rem',
                                  color: isSelected ? 'var(--text-1)' : 'var(--text-2)',
                                }}
                              >
                                {p.description || `${p.module}:${p.action}`}
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

          <div style={{ display: 'flex', gap: 8, marginTop: 24 }}>
            <button
              className="btn btn-secondary"
              style={{ flex: 1 }}
              onClick={onClose}
              disabled={loading}
            >
              Batal
            </button>
            <button
              className="btn btn-primary"
              style={{ flex: 1 }}
              onClick={handleSubmit}
              disabled={loading}
            >
              {loading ? <Loader2 size={14} className="loading-spin" /> : null}
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
    <div className="page-container">
      <div className="page-header">
        <div>
          <h1 className="page-title">Peran & Hak Akses</h1>
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

      <div className="card">
        {loading ? (
          <div style={{ textAlign: 'center', padding: '40px 0', color: 'var(--text-3)' }}>
            <Loader2 size={24} className="loading-spin mx-auto mb-2" />
            <p>Memuat peran...</p>
          </div>
        ) : (
          <div className="table-responsive">
            <table className="table">
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
                          {role.description || '-'}
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
        <RoleFormModal
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
