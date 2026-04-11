'use client';

import { useState, useEffect, useCallback } from 'react';
import {
  Users,
  Plus,
  Loader2,
  X,
  Search,
  Shield,
  Edit3,
  Archive,
  ChevronRight,
  Mail,
  Store,
  KeyRound,
  Eye,
  EyeOff,
  CheckCircle2,
  XCircle,
} from 'lucide-react';
import { usePermission } from '@/hooks/usePermission';
import { usersAdminApi, rolesApi, storesApi } from '@/lib/api/store-apis';
import { ApiError } from '@/lib/api/client';
import type { UserAdmin, Role, UserStoreAssignment, Store, PaginatedData } from '@/types';

type CreateUserResponse = { id: string };
type UsersListResponse = { data?: UserAdmin[]; meta?: { total?: number } };
type RolesResponse = { data?: Role[] };

// ── Role badge ────────────────────────────────────────────────────────────────
const ROLE_COLORS: Record<string, string> = {
  superadmin: '#8b5cf6',
  admin: '#6366f1',
  manager: '#0ea5e9',
  cashier: '#10b981',
  staff: '#f59e0b',
};

function RoleBadge({ name }: { name: string }) {
  const color = ROLE_COLORS[name] ?? '#6b7280';
  return (
    <span
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        gap: 3,
        padding: '2px 8px',
        borderRadius: 6,
        fontSize: '0.72rem',
        fontWeight: 700,
        background: `${color}20`,
        color,
      }}
    >
      {name}
    </span>
  );
}

function StatusBadge({ active }: { active: boolean }) {
  return (
    <span
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        gap: 4,
        padding: '2px 8px',
        borderRadius: 6,
        fontSize: '0.72rem',
        fontWeight: 600,
        background: active ? 'rgba(16,185,129,0.12)' : 'rgba(239,68,68,0.1)',
        color: active ? '#10b981' : '#f87171',
      }}
    >
      {active ? <CheckCircle2 size={10} /> : <XCircle size={10} />}
      {active ? 'Aktif' : 'Nonaktif'}
    </span>
  );
}

// ── Avatar ────────────────────────────────────────────────────────────────────
function Avatar({ name, size = 32 }: { name: string; size?: number }) {
  const initials = name
    .split(' ')
    .map(w => w[0])
    .join('')
    .slice(0, 2)
    .toUpperCase();
  return (
    <div
      style={{
        width: size,
        height: size,
        borderRadius: '50%',
        background: 'linear-gradient(135deg, var(--accent-in), var(--accent-em))',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        fontWeight: 800,
        fontSize: size * 0.34,
        color: '#fff',
        flexShrink: 0,
      }}
    >
      {initials}
    </div>
  );
}

// ── Password Input ────────────────────────────────────────────────────────────
function PasswordInput({
  value,
  onChange,
  placeholder = 'Password',
}: {
  value: string;
  onChange: (v: string) => void;
  placeholder?: string;
}) {
  const [show, setShow] = useState(false);
  return (
    <div style={{ position: 'relative' }}>
      <input
        className="input"
        type={show ? 'text' : 'password'}
        value={value}
        onChange={e => onChange(e.target.value)}
        placeholder={placeholder}
        style={{ paddingRight: 36 }}
      />
      <button
        type="button"
        style={{
          position: 'absolute',
          right: 10,
          top: '50%',
          transform: 'translateY(-50%)',
          background: 'none',
          border: 'none',
          cursor: 'pointer',
          color: 'var(--text-3)',
        }}
        onClick={() => setShow(s => !s)}
      >
        {show ? <EyeOff size={14} /> : <Eye size={14} />}
      </button>
    </div>
  );
}

// ── Store Assignment Manager ───────────────────────────────────────────────────
function StoreAssigner({
  value,
  onChange,
  roles,
  stores,
}: {
  value: { store_id: string; role_id: string }[];
  onChange: (v: { store_id: string; role_id: string }[]) => void;
  roles: Role[];
  stores: { id: string; name: string }[];
}) {
  const add = () => {
    if (stores.length === 0) return;
    const unused = stores.find(s => !value.some(v => v.store_id === s.id));
    if (!unused) return;
    onChange([...value, { store_id: unused.id, role_id: roles[0]?.id ?? '' }]);
  };
  const remove = (idx: number) => onChange(value.filter((_, i) => i !== idx));
  const update = (idx: number, field: 'store_id' | 'role_id', val: string) =>
    onChange(value.map((v, i) => (i === idx ? { ...v, [field]: val } : v)));

  return (
    <div>
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          marginBottom: 8,
        }}
      >
        <span style={{ fontSize: '0.78rem', fontWeight: 600, color: 'var(--text-2)' }}>
          Penugasan Toko
        </span>
        <button
          type="button"
          className="btn btn-secondary btn-sm"
          onClick={add}
          style={{ fontSize: '0.72rem' }}
        >
          <Plus size={11} /> Tambah
        </button>
      </div>
      {value.length === 0 && (
        <div
          style={{
            padding: '10px 12px',
            border: '1px dashed var(--border)',
            borderRadius: 8,
            fontSize: '0.8rem',
            color: 'var(--text-3)',
            textAlign: 'center',
          }}
        >
          Belum ada penugasan
        </div>
      )}
      {value.map((a, i) => (
        <div key={i} style={{ display: 'flex', gap: 6, marginBottom: 6, alignItems: 'center' }}>
          <select
            className="input"
            style={{ flex: 2, fontSize: '0.8rem' }}
            value={a.store_id}
            onChange={e => update(i, 'store_id', e.target.value)}
          >
            {stores.map(s => (
              <option key={s.id} value={s.id}>
                {s.name}
              </option>
            ))}
          </select>
          <select
            className="input"
            style={{ flex: 1, fontSize: '0.8rem' }}
            value={a.role_id}
            onChange={e => update(i, 'role_id', e.target.value)}
          >
            {roles.map(r => (
              <option key={r.id} value={r.id}>
                {r.name}
              </option>
            ))}
          </select>
          <button
            type="button"
            style={{
              background: 'none',
              border: 'none',
              cursor: 'pointer',
              color: 'var(--text-3)',
              padding: 4,
            }}
            onClick={() => remove(i)}
          >
            <X size={13} />
          </button>
        </div>
      ))}
    </div>
  );
}

// ── Create / Edit Modal ───────────────────────────────────────────────────────
type FormMode = 'create' | 'edit';

interface FormUser {
  name: string;
  email: string;
  password: string;
  stores: { store_id: string; role_id: string }[];
}

function UserFormModal({
  mode,
  user,
  roles,
  stores,
  onSuccess,
  onClose,
}: {
  mode: FormMode;
  user?: UserAdmin;
  roles: Role[];
  stores: { id: string; name: string }[];
  onSuccess: () => void;
  onClose: () => void;
}) {
  const [form, setForm] = useState<FormUser>({
    name: user?.name ?? '',
    email: user?.email ?? '',
    password: '',
    stores: user?.stores?.map(s => ({ store_id: s.store_id, role_id: s.role_id })) ?? [],
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async () => {
    setError('');
    if (!form.name.trim() || !form.email.trim()) {
      setError('Nama dan email wajib diisi');
      return;
    }
    if (mode === 'create' && !form.password) {
      setError('Password wajib diisi');
      return;
    }

    setLoading(true);
    try {
      if (mode === 'create') {
        const created = (await usersAdminApi.create({
          name: form.name,
          email: form.email,
          password: form.password,
          stores: form.stores,
        })) as CreateUserResponse;
        if (form.stores.length > 0) {
          await usersAdminApi.setStores(created.id, form.stores);
        }
      } else if (user) {
        await usersAdminApi.update(user.id, { name: form.name, email: form.email });
        await usersAdminApi.setStores(user.id, form.stores);
      }
      onSuccess();
    } catch (e) {
      setError(e instanceof ApiError ? e.message : 'Gagal menyimpan');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-box" style={{ maxWidth: 520 }} onClick={e => e.stopPropagation()}>
        <div
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            marginBottom: 20,
          }}
        >
          <h2 style={{ fontWeight: 800, fontSize: '1.05rem' }}>
            {mode === 'create' ? 'Tambah Pengguna' : 'Edit Pengguna'}
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
          <div>
            <label className="label">Nama</label>
            <input
              className="input"
              value={form.name}
              onChange={e => setForm(f => ({ ...f, name: e.target.value }))}
              placeholder="Nama lengkap"
            />
          </div>
          <div>
            <label className="label">Email</label>
            <input
              className="input"
              type="email"
              value={form.email}
              onChange={e => setForm(f => ({ ...f, email: e.target.value }))}
              placeholder="email@domain.com"
            />
          </div>
          {mode === 'create' && (
            <div>
              <label className="label">Password</label>
              <PasswordInput
                value={form.password}
                onChange={v => setForm(f => ({ ...f, password: v }))}
              />
            </div>
          )}

          <div style={{ borderTop: '1px solid var(--border)', paddingTop: 12 }}>
            <StoreAssigner
              value={form.stores}
              onChange={s => setForm(f => ({ ...f, stores: s }))}
              roles={roles}
              stores={stores}
            />
          </div>
        </div>

        <div style={{ display: 'flex', gap: 8, marginTop: 20 }}>
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
            {loading ? 'Menyimpan...' : mode === 'create' ? 'Buat Pengguna' : 'Simpan'}
          </button>
        </div>
      </div>
    </div>
  );
}

// ── Reset Password Modal ──────────────────────────────────────────────────────
function ResetPasswordModal({
  user,
  onSuccess,
  onClose,
}: {
  user: UserAdmin;
  onSuccess: () => void;
  onClose: () => void;
}) {
  const [pw, setPw] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleReset = async () => {
    if (!pw || pw.length < 6) {
      setError('Minimal 6 karakter');
      return;
    }
    setLoading(true);
    try {
      await usersAdminApi.resetPassword(user.id, { password: pw });
      onSuccess();
    } catch (e) {
      setError(e instanceof ApiError ? e.message : 'Gagal reset password');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-box" style={{ maxWidth: 380 }} onClick={e => e.stopPropagation()}>
        <div style={{ textAlign: 'center', padding: '8px 0 16px' }}>
          <KeyRound size={28} style={{ color: '#6366f1', marginBottom: 12 }} />
          <h2 style={{ fontWeight: 800, marginBottom: 4 }}>Reset Password</h2>
          <p style={{ fontSize: '0.82rem', color: 'var(--text-2)' }}>{user.name}</p>
        </div>
        {error && (
          <div
            style={{
              background: 'rgba(239,68,68,0.1)',
              borderRadius: 8,
              padding: '8px 12px',
              color: '#f87171',
              fontSize: '0.82rem',
              marginBottom: 12,
            }}
          >
            {error}
          </div>
        )}
        <div style={{ marginBottom: 16 }}>
          <label className="label">Password Baru</label>
          <PasswordInput value={pw} onChange={setPw} placeholder="Min. 6 karakter" />
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
            className="btn btn-primary"
            style={{ flex: 1 }}
            onClick={handleReset}
            disabled={loading}
          >
            {loading ? <Loader2 size={14} className="loading-spin" /> : <KeyRound size={14} />}
            {loading ? 'Menyimpan...' : 'Simpan'}
          </button>
        </div>
      </div>
    </div>
  );
}

// ── Deactivate Confirm ────────────────────────────────────────────────────────
function DeactivateConfirm({
  user,
  onSuccess,
  onClose,
}: {
  user: UserAdmin;
  onSuccess: () => void;
  onClose: () => void;
}) {
  const [loading, setLoading] = useState(false);
  const handle = async () => {
    setLoading(true);
    try {
      await usersAdminApi.deactivate(user.id);
      onSuccess();
    } catch (e) {
      alert(e instanceof ApiError ? e.message : 'Gagal menonaktifkan');
    } finally {
      setLoading(false);
    }
  };
  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-box" style={{ maxWidth: 400 }} onClick={e => e.stopPropagation()}>
        <div style={{ textAlign: 'center', padding: '8px 0 16px' }}>
          <Archive size={28} style={{ color: '#f59e0b', marginBottom: 12 }} />
          <h2 style={{ fontWeight: 800, marginBottom: 8 }}>Nonaktifkan Pengguna?</h2>
          <p style={{ color: 'var(--text-2)', fontSize: '0.875rem', lineHeight: 1.6 }}>
            <strong>{user.name}</strong> akan diarsipkan dan tidak dapat login. Data dan riwayat
            transaksi tetap tersimpan.
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
            onClick={handle}
            disabled={loading}
          >
            {loading ? <Loader2 size={14} className="loading-spin" /> : <Archive size={14} />}
            {loading ? 'Memproses...' : 'Ya, Nonaktifkan'}
          </button>
        </div>
      </div>
    </div>
  );
}

// ── Detail Drawer ─────────────────────────────────────────────────────────────
function DetailDrawer({
  user,
  roles: _roles,
  stores: _stores,
  onClose,
  onEdit,
  onDeactivate,
  onResetPw,
}: {
  user: UserAdmin;
  roles: Role[];
  stores: { id: string; name: string }[];
  onClose: () => void;
  onEdit: () => void;
  onDeactivate: () => void;
  onResetPw: () => void;
}) {
  return (
    <>
      <div
        onClick={onClose}
        style={{
          position: 'fixed',
          inset: 0,
          background: 'rgba(0,0,0,0.5)',
          backdropFilter: 'blur(2px)',
          zIndex: 200,
        }}
      />
      <div
        style={{
          position: 'fixed',
          top: 0,
          right: 0,
          bottom: 0,
          width: 420,
          background: 'var(--bg-card)',
          borderLeft: '1px solid var(--border)',
          zIndex: 201,
          overflow: 'hidden',
          display: 'flex',
          flexDirection: 'column',
        }}
      >
        {/* header */}
        <div
          style={{
            padding: '18px 20px',
            borderBottom: '1px solid var(--border)',
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
          }}
        >
          <div style={{ fontWeight: 800 }}>Detail Pengguna</div>
          <div style={{ display: 'flex', gap: 6 }}>
            <button className="btn btn-secondary btn-sm" onClick={onEdit}>
              <Edit3 size={13} /> Edit
            </button>
            <button className="btn btn-secondary btn-sm" onClick={onResetPw} title="Reset Password">
              <KeyRound size={13} />
            </button>
            {user.is_active && (
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
                onClick={onDeactivate}
              >
                <Archive size={12} /> Nonaktifkan
              </button>
            )}
            <button className="btn btn-ghost btn-sm" onClick={onClose}>
              <X size={15} />
            </button>
          </div>
        </div>

        {/* body */}
        <div style={{ flex: 1, overflowY: 'auto', padding: 20 }}>
          {/* profile */}
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: 14,
              background: 'var(--bg-elevated)',
              borderRadius: 12,
              padding: '16px',
              marginBottom: 16,
              border: '1px solid var(--border-md)',
            }}
          >
            <Avatar name={user.name} size={48} />
            <div>
              <div style={{ fontWeight: 800, fontSize: '1.05rem' }}>{user.name}</div>
              <div
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 4,
                  fontSize: '0.8rem',
                  color: 'var(--text-2)',
                  marginTop: 2,
                }}
              >
                <Mail size={11} /> {user.email}
              </div>
              <div style={{ marginTop: 6 }}>
                <StatusBadge active={user.is_active} />
              </div>
            </div>
          </div>

          {/* store memberships */}
          <div
            style={{
              fontSize: '0.68rem',
              fontWeight: 700,
              color: 'var(--text-3)',
              textTransform: 'uppercase',
              letterSpacing: '0.08em',
              marginBottom: 8,
              display: 'flex',
              alignItems: 'center',
              gap: 4,
            }}
          >
            <Store size={11} /> Penugasan Toko ({user.store_count})
          </div>
          {user.stores && user.stores.length > 0 ? (
            user.stores.map((s: UserStoreAssignment, i: number) => (
              <div
                key={i}
                style={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  alignItems: 'center',
                  padding: '10px 12px',
                  background: 'var(--bg-elevated)',
                  borderRadius: 8,
                  marginBottom: 6,
                  border: '1px solid var(--border)',
                }}
              >
                <div>
                  <div style={{ fontWeight: 600, fontSize: '0.88rem' }}>{s.store_name}</div>
                  <div style={{ fontSize: '0.72rem', color: 'var(--text-3)' }}>{s.store_type}</div>
                </div>
                <RoleBadge name={s.role_name} />
              </div>
            ))
          ) : (
            <div
              style={{
                padding: '10px 12px',
                border: '1px dashed var(--border)',
                borderRadius: 8,
                fontSize: '0.8rem',
                color: 'var(--text-3)',
                textAlign: 'center',
              }}
            >
              Belum ditugaskan ke toko
            </div>
          )}

          {/* permissions via roles */}
          <div
            style={{
              marginTop: 16,
              fontSize: '0.68rem',
              fontWeight: 700,
              color: 'var(--text-3)',
              textTransform: 'uppercase',
              letterSpacing: '0.08em',
              display: 'flex',
              alignItems: 'center',
              gap: 4,
            }}
          >
            <Shield size={11} /> Peran
          </div>
          <div style={{ display: 'flex', flexWrap: 'wrap', gap: 4, marginTop: 6 }}>
            {user.stores && user.stores.length > 0 ? (
              [...new Set(user.stores.map((s: UserStoreAssignment) => s.role_name))].map(r => (
                <RoleBadge key={r} name={r} />
              ))
            ) : (
              <span style={{ fontSize: '0.78rem', color: 'var(--text-3)', fontStyle: 'italic' }}>
                —
              </span>
            )}
          </div>

          {/* meta */}
          <div
            style={{
              marginTop: 20,
              padding: '12px 14px',
              background: 'var(--bg-elevated)',
              borderRadius: 8,
              border: '1px solid var(--border)',
              display: 'flex',
              flexDirection: 'column',
              gap: 4,
            }}
          >
            {[
              {
                label: 'Dibuat',
                val: new Date(user.created_at).toLocaleDateString('id-ID', { dateStyle: 'long' }),
              },
              {
                label: 'Diperbarui',
                val: new Date(user.updated_at).toLocaleDateString('id-ID', { dateStyle: 'long' }),
              },
            ].map(({ label, val }) => (
              <div
                key={label}
                style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.8rem' }}
              >
                <span style={{ color: 'var(--text-3)' }}>{label}</span>
                <span style={{ fontWeight: 500 }}>{val}</span>
              </div>
            ))}
          </div>
        </div>
      </div>
    </>
  );
}

// ── Main Page ─────────────────────────────────────────────────────────────────
export default function UsersPage() {
  const { can } = usePermission();
  const [users, setUsers] = useState<UserAdmin[]>([]);
  const [roles, setRoles] = useState<Role[]>([]);
  const [allStores, setAllStores] = useState<{ id: string; name: string }[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState('');
  const [includeInactive, setIncludeInactive] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  // Modals
  const [detail, setDetail] = useState<UserAdmin | null>(null);
  const [createOpen, setCreateOpen] = useState(false);
  const [editUser, setEditUser] = useState<UserAdmin | null>(null);
  const [deactivating, setDeactivating] = useState<UserAdmin | null>(null);
  const [resetting, setResetting] = useState<UserAdmin | null>(null);

  const PER_PAGE = 20;

  const load = useCallback(
    async (p = 1) => {
      setLoading(true);
      setError('');
      try {
        const body = (await usersAdminApi.list({
          search,
          include_inactive: includeInactive,
          page: p,
          per_page: PER_PAGE,
        })) as UsersListResponse;
        setUsers(body.data ?? []);
        setTotal(body.meta?.total ?? 0);
        setPage(p);
      } catch (e) {
        setError(e instanceof ApiError ? e.message : 'Gagal memuat pengguna');
      } finally {
        setLoading(false);
      }
    },
    [search, includeInactive]
  );

  useEffect(() => {
    load(1);
  }, [load]);

  // Fetch roles and stores once
  useEffect(() => {
    rolesApi
      .list()
      .then(body => {
        const d = body as RolesResponse;
        setRoles(d.data ?? []);
      })
      .catch(() => {});
    storesApi
      .list()
      .then(body => {
        const list = (body.data as PaginatedData<Store>).data ?? [];
        setAllStores(list.map(s => ({ id: s.id, name: s.name })));
      })
      .catch(() => {});
  }, []);

  const openDetail = async (u: UserAdmin) => {
    try {
      const user = (await usersAdminApi.get(u.id)) as UserAdmin;
      setDetail(user);
    } catch {
      setDetail(u);
    }
  };

  const refresh = () => {
    load(page);
    if (detail) openDetail(detail);
  };
  const totalPages = Math.ceil(total / PER_PAGE);

  return (
    <div className="w-full p-6">
      {/* ── Header ── */}
      <div
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
            Manajemen Pengguna
          </h1>
          <p className="page-subtitle">Kelola akun pengguna dan penugasan toko</p>
        </div>
        {can('users.create') && (
          <button className="btn btn-primary" onClick={() => setCreateOpen(true)}>
            <Plus size={16} /> Tambah Pengguna
          </button>
        )}
      </div>

      {/* ── Filters ── */}
      <div
        style={{
          display: 'flex',
          gap: 8,
          marginBottom: 16,
          flexWrap: 'wrap',
          alignItems: 'center',
        }}
      >
        <div style={{ position: 'relative', flex: 1, minWidth: 200 }}>
          <Search
            size={14}
            style={{
              position: 'absolute',
              left: 10,
              top: '50%',
              transform: 'translateY(-50%)',
              color: 'var(--text-3)',
            }}
          />
          <input
            className="input"
            style={{ paddingLeft: 32 }}
            placeholder="Cari nama atau email..."
            value={search}
            onChange={e => setSearch(e.target.value)}
            onKeyDown={e => e.key === 'Enter' && load(1)}
          />
        </div>
        <label
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: 6,
            fontSize: '0.83rem',
            cursor: 'pointer',
            userSelect: 'none',
          }}
        >
          <input
            type="checkbox"
            checked={includeInactive}
            onChange={e => setIncludeInactive(e.target.checked)}
          />
          Tampilkan nonaktif
        </label>
        <button className="btn btn-primary" onClick={() => load(1)} disabled={loading}>
          {loading ? <Loader2 size={14} className="loading-spin" /> : null}
          Cari
        </button>
      </div>

      {/* ── Summary ── */}
      <div
        style={{ display: 'grid', gridTemplateColumns: 'repeat(3,1fr)', gap: 12, marginBottom: 16 }}
      >
        {[
          { label: 'Total Pengguna', val: total, color: '#6366f1' },
          { label: 'Aktif', val: users.filter(u => u.is_active).length, color: '#10b981' },
          { label: 'Nonaktif', val: users.filter(u => !u.is_active).length, color: '#f59e0b' },
        ].map(({ label, val, color }) => (
          <div key={label} className="stat-card" style={{ padding: '14px 16px' }}>
            <div style={{ fontSize: '0.75rem', color: 'var(--text-3)', marginBottom: 4 }}>
              {label}
            </div>
            <div style={{ fontWeight: 800, fontSize: '1.4rem', color }}>{val}</div>
          </div>
        ))}
      </div>

      {/* ── Error ── */}
      {error && (
        <div
          style={{
            background: 'rgba(239,68,68,0.1)',
            border: '1px solid rgba(239,68,68,0.3)',
            borderRadius: 8,
            padding: '10px 14px',
            color: '#f87171',
            fontSize: '0.85rem',
            marginBottom: 12,
            display: 'flex',
            justifyContent: 'space-between',
          }}
        >
          {error}
          <button
            onClick={() => setError('')}
            style={{ background: 'none', border: 'none', color: '#f87171', cursor: 'pointer' }}
          >
            <X size={14} />
          </button>
        </div>
      )}

      {/* ── Table ── */}
      <div className="card" style={{ overflow: 'hidden' }}>
        {loading ? (
          <div style={{ display: 'flex', justifyContent: 'center', padding: 48 }}>
            <Loader2 size={26} className="loading-spin" style={{ color: 'var(--accent-em)' }} />
          </div>
        ) : users.length === 0 ? (
          <div className="empty-state" style={{ padding: 48 }}>
            <Users size={36} />
            <p>Tidak ada pengguna</p>
          </div>
        ) : (
          <table className="tbl">
            <thead>
              <tr>
                <th>Pengguna</th>
                <th>Email</th>
                <th>Toko</th>
                <th>Peran</th>
                <th>Status</th>
                <th style={{ width: 80 }} />
              </tr>
            </thead>
            <tbody>
              {users.map(u => (
                <tr
                  key={u.id}
                  style={{ cursor: 'pointer', opacity: u.is_active ? 1 : 0.6 }}
                  onClick={() => openDetail(u)}
                >
                  <td>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                      <Avatar name={u.name} size={30} />
                      <div style={{ fontWeight: 600, fontSize: '0.88rem' }}>{u.name}</div>
                    </div>
                  </td>
                  <td style={{ fontSize: '0.82rem', color: 'var(--text-2)' }}>{u.email}</td>
                  <td>
                    <span
                      style={{
                        display: 'inline-flex',
                        alignItems: 'center',
                        gap: 4,
                        fontSize: '0.78rem',
                        color: 'var(--text-2)',
                      }}
                    >
                      <Store size={11} /> {u.store_count} toko
                    </span>
                  </td>
                  <td>
                    {u.stores && u.stores.length > 0 ? (
                      <RoleBadge name={u.stores[0].role_name} />
                    ) : (
                      <span
                        style={{ color: 'var(--text-3)', fontStyle: 'italic', fontSize: '0.8rem' }}
                      >
                        —
                      </span>
                    )}
                  </td>
                  <td>
                    <StatusBadge active={u.is_active} />
                  </td>
                  <td onClick={e => e.stopPropagation()}>
                    <div style={{ display: 'flex', gap: 4 }}>
                      {can('users.update') && (
                        <button
                          className="btn btn-ghost btn-sm"
                          onClick={() => {
                            setEditUser(u);
                          }}
                          title="Edit"
                        >
                          <Edit3 size={13} />
                        </button>
                      )}
                      {can('users.update') && u.is_active && (
                        <button
                          title="Nonaktifkan"
                          style={{
                            display: 'flex',
                            alignItems: 'center',
                            padding: '4px 6px',
                            borderRadius: 6,
                            border: '1px solid rgba(245,158,11,0.35)',
                            background: 'rgba(245,158,11,0.08)',
                            color: '#f59e0b',
                            cursor: 'pointer',
                          }}
                          onClick={() => setDeactivating(u)}
                        >
                          <Archive size={12} />
                        </button>
                      )}
                      <button className="btn btn-ghost btn-sm" onClick={() => openDetail(u)}>
                        <ChevronRight size={14} />
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      {/* ── Pagination ── */}
      {totalPages > 1 && (
        <div
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            marginTop: 14,
          }}
        >
          <span style={{ fontSize: '0.8rem', color: 'var(--text-3)' }}>
            Hal {page} dari {totalPages} · {total} pengguna
          </span>
          <div style={{ display: 'flex', gap: 6 }}>
            <button
              className="btn btn-secondary btn-sm"
              disabled={page <= 1}
              onClick={() => load(page - 1)}
            >
              ‹
            </button>
            <button
              className="btn btn-secondary btn-sm"
              disabled={page >= totalPages}
              onClick={() => load(page + 1)}
            >
              ›
            </button>
          </div>
        </div>
      )}

      {/* ── Modals ── */}
      {createOpen && (
        <UserFormModal
          mode="create"
          roles={roles}
          stores={allStores}
          onSuccess={() => {
            setCreateOpen(false);
            load(1);
          }}
          onClose={() => setCreateOpen(false)}
        />
      )}
      {editUser && (
        <UserFormModal
          mode="edit"
          user={editUser}
          roles={roles}
          stores={allStores}
          onSuccess={() => {
            setEditUser(null);
            refresh();
          }}
          onClose={() => setEditUser(null)}
        />
      )}
      {deactivating && (
        <DeactivateConfirm
          user={deactivating}
          onSuccess={() => {
            setDeactivating(null);
            load(page);
            if (detail?.id === deactivating.id) setDetail(null);
          }}
          onClose={() => setDeactivating(null)}
        />
      )}
      {resetting && (
        <ResetPasswordModal
          user={resetting}
          onSuccess={() => setResetting(null)}
          onClose={() => setResetting(null)}
        />
      )}
      {detail && (
        <DetailDrawer
          user={detail}
          roles={roles}
          stores={allStores}
          onClose={() => setDetail(null)}
          onEdit={() => {
            setEditUser(detail);
            setDetail(null);
          }}
          onDeactivate={() => {
            setDeactivating(detail);
            setDetail(null);
          }}
          onResetPw={() => setResetting(detail)}
        />
      )}
    </div>
  );
}
