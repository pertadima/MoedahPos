'use client';

import { useEffect, useState, useCallback } from 'react';
import { Wallet, Plus, Loader2, X, Calendar, Edit2, Trash2 } from 'lucide-react';
import { useAuth } from '@/lib/auth/AuthContext';
import { expensesApi, recurringExpensesApi } from '@/lib/api/store-apis';
import { formatRp, formatDate, parseNumberInput, formatNumberInput } from '@/lib/utils';
import type { Expense, ExpenseCategory, RecurringExpense } from '@/types';

function ExpensesPage() {
  const { selectedStore } = useAuth();
  const [activeTab, setActiveTab] = useState<'riwayat' | 'rutin'>('riwayat');
  const [expenses, setExpenses] = useState<Expense[]>([]);
  const [categories, setCategories] = useState<ExpenseCategory[]>([]);
  const [loading, setLoading] = useState(true);

  const [filter, setFilter] = useState({
    category_id: '',
    date_from: '',
    date_to: '',
  });

  const [meta, setMeta] = useState({ page: 1, per_page: 20, total: 0, total_pages: 0 });

  const loadExpenses = useCallback(async () => {
    if (!selectedStore) return;
    setLoading(true);
    try {
      const res = await expensesApi.list(selectedStore.store_id, {
        ...filter,
        page: meta.page,
        per_page: meta.per_page,
      });
      setExpenses(res.data.data);
      setMeta(res.data.meta);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  }, [selectedStore, filter, meta.page, meta.per_page]);

  const loadCategories = useCallback(async () => {
    try {
      const res = await expensesApi.listCategories();
      setCategories(res.data || []);
    } catch (err) {
      console.error(err);
    }
  }, []);

  useEffect(() => {
    loadCategories();
  }, [loadCategories]);

  useEffect(() => {
    loadExpenses();
  }, [loadExpenses]);

  const [showModal, setShowModal] = useState(false);
  const [editTarget, setEditTarget] = useState<Expense | null>(null);

  const handleDelete = async (id: string) => {
    if (!selectedStore) return;
    if (!confirm('Hapus pengeluaran ini?')) return;
    try {
      await expensesApi.delete(selectedStore.store_id, id);
      loadExpenses();
    } catch (err: unknown) {
      alert('Gagal menghapus pengeluaran');
      console.error(err);
    }
  };

  const handleMarkPaid = async (id: string) => {
    if (!selectedStore) return;
    try {
      await expensesApi.updateStatus(selectedStore.store_id, id, { payment_status: 'paid' });
      loadExpenses();
    } catch (err) {
      alert('Gagal mengubah status pengeluaran');
      console.error(err);
    }
  };

  const handleFilterChange = (key: string, value: string) => {
    setFilter(prev => ({ ...prev, [key]: value }));
    setMeta(prev => ({ ...prev, page: 1 }));
  };

  return (
    <div style={{ padding: '24px 32px', maxWidth: '100%', margin: '0 auto' }}>
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          marginBottom: 24,
        }}
      >
        <div>
          <h1
            style={{
              fontSize: '1.4rem',
              fontWeight: 700,
              margin: '0 0 4px',
              color: 'var(--text-1)',
            }}
          >
            Pengeluaran Operasional
          </h1>
          <p style={{ margin: 0, fontSize: '0.85rem', color: 'var(--text-3)' }}>
            Catat dan kelola semua pengeluaran selain stok produk.
          </p>
        </div>
       {activeTab !== 'rutin' && (
          <button
            onClick={() => {
              setEditTarget(null);
              setShowModal(true);
            }}
            className="btn btn-primary"
            style={{ display: 'flex', alignItems: 'center', gap: 6 }}
            >
            <Plus size={16} /> Tambah Pengeluaran
          </button>
        )}
        
      </div>

      <div
        style={{
          display: 'flex',
          gap: 16,
          marginBottom: 24,
          borderBottom: '1px solid var(--border)',
        }}
      >
        <button
          onClick={() => setActiveTab('riwayat')}
          style={{
            background: 'none',
            border: 'none',
            padding: '8px 16px',
            fontSize: '1rem',
            fontWeight: activeTab === 'riwayat' ? 600 : 400,
            color: activeTab === 'riwayat' ? 'var(--primary)' : 'var(--text-3)',
            borderBottom:
              activeTab === 'riwayat' ? '2px solid var(--primary)' : '2px solid transparent',
            cursor: 'pointer',
          }}
        >
          Riwayat Pengeluaran
        </button>
        <button
          onClick={() => setActiveTab('rutin')}
          style={{
            background: 'none',
            border: 'none',
            padding: '8px 16px',
            fontSize: '1rem',
            fontWeight: activeTab === 'rutin' ? 600 : 400,
            color: activeTab === 'rutin' ? 'var(--primary)' : 'var(--text-3)',
            borderBottom:
              activeTab === 'rutin' ? '2px solid var(--primary)' : '2px solid transparent',
            cursor: 'pointer',
          }}
        >
          Pengeluaran Rutin
        </button>
      </div>

      {activeTab === 'riwayat' ? (
        <>
          <div
            style={{
              background: 'var(--bg-card)',
              padding: 16,
              borderRadius: 12,
              border: '1px solid var(--border)',
              marginBottom: 20,
              display: 'flex',
              gap: 16,
              alignItems: 'flex-end',
              flexWrap: 'wrap',
            }}
          >
            <div style={{ flex: '1 1 180px' }}>
              <label className="label">Kategori</label>
              <select
                className="input"
                style={{ width: '100%', height: 38 }}
                value={filter.category_id}
                onChange={e => handleFilterChange('category_id', e.target.value)}
              >
                <option value="">Semua Kategori</option>
                {categories.map(c => (
                  <option key={c.id} value={c.id}>
                    {c.name}
                  </option>
                ))}
              </select>
            </div>
            <div style={{ flex: '1 1 150px' }}>
              <label className="label">Dari Tanggal</label>
              <input
                type="date"
                className="input"
                style={{ width: '100%', height: 38 }}
                value={filter.date_from}
                onChange={e => handleFilterChange('date_from', e.target.value)}
              />
            </div>
            <div style={{ flex: '1 1 150px' }}>
              <label className="label">Sampai Tanggal</label>
              <input
                type="date"
                className="input"
                style={{ width: '100%', height: 38 }}
                value={filter.date_to}
                onChange={e => handleFilterChange('date_to', e.target.value)}
              />
            </div>
            <button
              onClick={() => setFilter({ category_id: '', date_from: '', date_to: '' })}
              className="btn btn-secondary"
              style={{ height: 38 }}
            >
              Reset
            </button>
          </div>

          {loading && expenses.length === 0 ? (
            <div style={{ textAlign: 'center', padding: 40, color: 'var(--text-3)' }}>
              <Loader2 size={32} className="loading-spin" style={{ margin: '0 auto 12px' }} />
              Memuat data pengeluaran...
            </div>
          ) : expenses.length === 0 ? (
            <div
              style={{
                textAlign: 'center',
                padding: '60px 20px',
                background: 'var(--bg-card)',
                borderRadius: 12,
                border: '1px dashed var(--border)',
              }}
            >
              <Wallet size={48} style={{ color: 'var(--text-4)', margin: '0 auto 16px' }} />
              <h3 style={{ margin: '0 0 8px', fontSize: '1.1rem', color: 'var(--text-1)' }}>
                Belum Ada Pengeluaran
              </h3>
              <p style={{ margin: 0, color: 'var(--text-3)', fontSize: '0.9rem' }}>
                {filter.category_id || filter.date_from
                  ? 'Tidak ada data yang cocok dengan filter.'
                  : 'Mulai catat sewa, gaji, atau tagihan utilitas Anda.'}
              </p>
            </div>
          ) : (
            <div
              style={{
                background: 'var(--bg-card)',
                borderRadius: 12,
                border: '1px solid var(--border)',
                overflow: 'hidden',
              }}
            >
              <div style={{ overflowX: 'auto' }}>
                <table className="tbl" style={{ width: '100%', minWidth: 600 }}>
                  <thead>
                    <tr>
                      <th>Tanggal</th>
                      <th>Kategori</th>
                      <th>Catatan</th>
                      <th>Status</th>
                      <th style={{ textAlign: 'right' }}>Jumlah</th>
                      <th style={{ textAlign: 'right', width: 100 }}>Aksi</th>
                    </tr>
                  </thead>
                  <tbody>
                    {expenses.map(exp => (
                      <tr key={exp.id}>
                        <td>{formatDate(exp.expense_date)}</td>
                        <td>
                          <span
                            style={{
                              background: 'var(--bg-hover)',
                              padding: '4px 8px',
                              borderRadius: 6,
                              fontSize: '0.75rem',
                              fontWeight: 600,
                            }}
                          >
                            {exp.category_name}
                          </span>
                        </td>
                        <td style={{ color: 'var(--text-2)' }}>{exp.notes || '—'}</td>
                        <td>
                          {exp.payment_status === 'unpaid' ? (
                            <span
                              style={{
                                background: '#fee2e2',
                                color: '#dc2626',
                                padding: '4px 8px',
                                borderRadius: 6,
                                fontSize: '0.75rem',
                                fontWeight: 600,
                              }}
                            >
                              Belum Dibayar
                            </span>
                          ) : exp.payment_status === 'paid' ? (
                            <span
                              style={{
                                background: '#dcfce7',
                                color: '#16a34a',
                                padding: '4px 8px',
                                borderRadius: 6,
                                fontSize: '0.75rem',
                                fontWeight: 600,
                              }}
                            >
                              Lunas
                            </span>
                          ) : (
                            <span
                              style={{
                                background: 'var(--bg-hover)',
                                color: 'var(--text-3)',
                                padding: '4px 8px',
                                borderRadius: 6,
                                fontSize: '0.75rem',
                                fontWeight: 600,
                              }}
                            >
                              Dibatalkan
                            </span>
                          )}
                        </td>
                        <td style={{ textAlign: 'right', fontWeight: 600, color: '#dc2626' }}>
                          {formatRp(exp.amount)}
                        </td>
                        <td
                          style={{
                            textAlign: 'right',
                            gap: 8,
                            display: 'flex',
                            justifyContent: 'flex-end',
                            alignItems: 'center',
                          }}
                        >
                          {exp.payment_status === 'unpaid' && (
                            <button
                              onClick={() => handleMarkPaid(exp.id)}
                              className="btn btn-secondary btn-sm"
                              style={{ fontSize: '0.75rem' }}
                            >
                              Tandai Lunas
                            </button>
                          )}
                          <button
                            onClick={() => {
                              setEditTarget(exp);
                              setShowModal(true);
                            }}
                            style={{
                              background: 'none',
                              border: 'none',
                              cursor: 'pointer',
                              color: 'var(--text-3)',
                              marginRight: 10,
                            }}
                          >
                            <Edit2 size={16} />
                          </button>
                          <button
                            onClick={() => handleDelete(exp.id)}
                            style={{
                              background: 'none',
                              border: 'none',
                              cursor: 'pointer',
                              color: '#dc2626',
                            }}
                          >
                            <Trash2 size={16} />
                          </button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>

              {meta.total_pages > 1 && (
                <div
                  style={{
                    padding: 16,
                    borderTop: '1px solid var(--border)',
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center',
                  }}
                >
                  <div style={{ fontSize: '0.85rem', color: 'var(--text-3)' }}>
                    Hal {meta.page} dari {meta.total_pages} ({meta.total} data)
                  </div>
                  <div style={{ display: 'flex', gap: 8 }}>
                    <button
                      className="btn btn-secondary btn-sm"
                      disabled={meta.page <= 1}
                      onClick={() => setMeta(prev => ({ ...prev, page: prev.page - 1 }))}
                    >
                      Prev
                    </button>
                    <button
                      className="btn btn-secondary btn-sm"
                      disabled={meta.page >= meta.total_pages}
                      onClick={() => setMeta(prev => ({ ...prev, page: prev.page + 1 }))}
                    >
                      Next
                    </button>
                  </div>
                </div>
              )}
            </div>
          )}
        </>
      ) : (
        <RecurringExpensesView categories={categories} />
      )}

      {showModal && activeTab === 'riwayat' && (
        <ExpenseModal
          categories={categories}
          expense={editTarget}
          onClose={() => setShowModal(false)}
          onSuccess={() => {
            setShowModal(false);
            loadExpenses();
          }}
        />
      )}
    </div>
  );
}

function ExpenseModal({
  categories,
  expense,
  onClose,
  onSuccess,
}: {
  categories: ExpenseCategory[];
  expense: Expense | null;
  onClose: () => void;
  onSuccess: () => void;
}) {
  const { selectedStore } = useAuth();
  const [saving, setSaving] = useState(false);

  const [form, setForm] = useState({
    category_id: expense?.category_id || (categories.length > 0 ? categories[0].id : ''),
    amount: expense ? formatNumberInput(expense.amount) : '',
    expense_date: expense?.expense_date || new Date().toISOString().slice(0, 10),
    notes: expense?.notes || '',
    payment_status: expense?.payment_status || 'paid',
  });

  const [err, setErr] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedStore) return;

    const amt = parseNumberInput(form.amount);
    if (!amt || amt <= 0) {
      setErr('Jumlah pengeluaran tidak valid');
      return;
    }
    if (!form.category_id) {
      setErr('Pilih kategori');
      return;
    }

    setSaving(true);
    setErr('');
    try {
      const payload = {
        ...form,
        amount: amt,
      };

      if (expense) {
        await expensesApi.update(selectedStore.store_id, expense.id, payload);
      } else {
        await expensesApi.create(selectedStore.store_id, payload);
      }
      onSuccess();
    } catch (error: any) {
      setErr(error.message || 'Gagal menyimpan pengeluaran');
    } finally {
      setSaving(false);
    }
  };

  return (
    <div
      style={{
        position: 'fixed',
        inset: 0,
        zIndex: 100,
        background: 'rgba(0,0,0,0.5)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
      }}
    >
      <div
        style={{
          background: 'var(--bg-card)',
          padding: 24,
          borderRadius: 12,
          width: 400,
          maxWidth: '90%',
        }}
      >
        <div
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            marginBottom: 20,
          }}
        >
          <h2 style={{ margin: 0, fontSize: '1.2rem' }}>
            {expense ? 'Ubah Pengeluaran' : 'Tambah Pengeluaran'}
          </h2>
          <button
            onClick={onClose}
            style={{
              background: 'none',
              border: 'none',
              cursor: 'pointer',
              color: 'var(--text-3)',
            }}
          >
            <X size={20} />
          </button>
        </div>

        {err && (
          <div style={{ color: '#dc2626', fontSize: '0.85rem', marginBottom: 16 }}>{err}</div>
        )}

        <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
          <div>
            <label className="label">Kategori</label>
            <select
              className="input"
              style={{ width: '100%' }}
              value={form.category_id}
              onChange={e => setForm({ ...form, category_id: e.target.value })}
              required
            >
              {categories.map(c => (
                <option key={c.id} value={c.id}>
                  {c.name}
                </option>
              ))}
            </select>
          </div>

          <div>
            <label className="label">Jumlah (Rp)</label>
            <input
              type="text"
              className="input"
              style={{ width: '100%', fontSize: '1.1rem', fontWeight: 600, color: '#dc2626' }}
              value={form.amount}
              onChange={e =>
                setForm({ ...form, amount: formatNumberInput(parseNumberInput(e.target.value)) })
              }
              placeholder="0"
              required
            />
          </div>

          <div>
            <label className="label">Tanggal</label>
            <input
              type="date"
              className="input"
              style={{ width: '100%' }}
              value={form.expense_date}
              onChange={e => setForm({ ...form, expense_date: e.target.value })}
              required
            />
          </div>

          <div>
            <label className="label">Status Pembayaran</label>
            <select
              className="input"
              style={{ width: '100%' }}
              value={form.payment_status}
              onChange={e => setForm({ ...form, payment_status: e.target.value as any })}
            >
              <option value="paid">Lunas</option>
              <option value="unpaid">Belum Dibayar</option>
            </select>
          </div>

          <div>
            <label className="label">Catatan (Opsional)</label>
            <textarea
              className="input"
              style={{ width: '100%', height: 80, resize: 'vertical' }}
              value={form.notes}
              onChange={e => setForm({ ...form, notes: e.target.value })}
              placeholder="Rincian pengeluaran..."
            />
          </div>

          <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 12, marginTop: 8 }}>
            <button type="button" className="btn btn-secondary" onClick={onClose}>
              Batal
            </button>
            <button type="submit" className="btn btn-primary" disabled={saving}>
              {saving ? <Loader2 size={16} className="loading-spin" /> : 'Simpan'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

function RecurringExpensesView({ categories }: { categories: ExpenseCategory[] }) {
  const { selectedStore } = useAuth();
  const [routines, setRoutines] = useState<RecurringExpense[]>([]);
  const [loading, setLoading] = useState(true);
  const [showModal, setShowModal] = useState(false);
  const [editTarget, setEditTarget] = useState<RecurringExpense | null>(null);

  const loadData = useCallback(async () => {
    if (!selectedStore) return;
    setLoading(true);
    try {
      const res = await recurringExpensesApi.list(selectedStore.store_id);
      setRoutines(res.data.data);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  }, [selectedStore]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  const handleDelete = async (id: string) => {
    if (!selectedStore) return;
    if (!confirm('Hapus pengaturan pengeluaran rutin ini?')) return;
    try {
      await recurringExpensesApi.delete(selectedStore.store_id, id);
      loadData();
    } catch (err) {
      alert('Gagal menghapus rutin');
      console.error(err);
    }
  };

  const getIntervalLabel = (interval: string, val: number) => {
    if (val === 1) {
      switch (interval) {
        case 'daily':
          return 'Harian';
        case 'weekly':
          return 'Mingguan';
        case 'monthly':
          return 'Bulanan';
        case 'yearly':
          return 'Tahunan';
      }
    }
    return `Setiap ${val} ${interval === 'daily' ? 'Hari' : interval === 'weekly' ? 'Minggu' : interval === 'monthly' ? 'Bulan' : 'Tahun'}`;
  };

  return (
    <>
      <div style={{ marginBottom: 16 }}>
        <button
          onClick={() => {
            setEditTarget(null);
            setShowModal(true);
          }}
          className="btn btn-secondary"
          style={{ display: 'flex', alignItems: 'center', gap: 6 }}
        >
          <Plus size={16} /> Buat Templat Rutin
        </button>
      </div>

      {loading && routines.length === 0 ? (
        <div style={{ textAlign: 'center', padding: 40, color: 'var(--text-3)' }}>
          <Loader2 size={32} className="loading-spin" style={{ margin: '0 auto 12px' }} />
          Memuat data rutin...
        </div>
      ) : routines.length === 0 ? (
        <div
          style={{
            textAlign: 'center',
            padding: '60px 20px',
            background: 'var(--bg-card)',
            borderRadius: 12,
            border: '1px dashed var(--border)',
          }}
        >
          <Calendar size={48} style={{ color: 'var(--text-4)', margin: '0 auto 16px' }} />
          <h3 style={{ margin: '0 0 8px', fontSize: '1.1rem', color: 'var(--text-1)' }}>
            Belum Ada Pengeluaran Rutin
          </h3>
          <p style={{ margin: 0, color: 'var(--text-3)', fontSize: '0.9rem' }}>
            Atur tagihan rutin seperti cicilan, sewa ruko, atau gaji karyawan di sini.
          </p>
        </div>
      ) : (
        <div
          style={{
            background: 'var(--bg-card)',
            borderRadius: 12,
            border: '1px solid var(--border)',
            overflow: 'hidden',
          }}
        >
          <div style={{ overflowX: 'auto' }}>
            <table className="tbl" style={{ width: '100%', minWidth: 700 }}>
              <thead>
                <tr>
                  <th>Nama Templat</th>
                  <th>Kategori</th>
                  <th>Siklus</th>
                  <th>Run Berikutnya</th>
                  <th>Status</th>
                  <th style={{ textAlign: 'right' }}>Jumlah</th>
                  <th style={{ textAlign: 'right', width: 100 }}>Aksi</th>
                </tr>
              </thead>
              <tbody>
                {routines.map(rt => (
                  <tr key={rt.id}>
                    <td>
                      <div style={{ fontWeight: 600 }}>{rt.name}</div>
                      <div style={{ fontSize: '0.8rem', color: 'var(--text-3)' }}>
                        Mulai: {formatDate(rt.start_date)}
                      </div>
                    </td>
                    <td>
                      <span
                        style={{
                          background: 'var(--bg-hover)',
                          padding: '4px 8px',
                          borderRadius: 6,
                          fontSize: '0.75rem',
                          fontWeight: 600,
                        }}
                      >
                        {rt.category_name}
                      </span>
                    </td>
                    <td>{getIntervalLabel(rt.interval, rt.interval_value)}</td>
                    <td style={{ color: 'var(--text-1)', fontWeight: 500 }}>
                      {formatDate(rt.next_run_date)}
                    </td>
                    <td>
                      {rt.is_active ? (
                        <span style={{ color: '#16a34a', fontSize: '0.85rem', fontWeight: 600 }}>
                          Aktif
                        </span>
                      ) : (
                        <span style={{ color: '#9ca3af', fontSize: '0.85rem', fontWeight: 600 }}>
                          Tidak Aktif
                        </span>
                      )}
                    </td>
                    <td style={{ textAlign: 'right', fontWeight: 600, color: '#dc2626' }}>
                      {formatRp(rt.amount)}
                    </td>
                    <td style={{ textAlign: 'right' }}>
                      <button
                        onClick={() => {
                          setEditTarget(rt);
                          setShowModal(true);
                        }}
                        style={{
                          background: 'none',
                          border: 'none',
                          cursor: 'pointer',
                          color: 'var(--text-3)',
                          marginRight: 10,
                        }}
                      >
                        <Edit2 size={16} />
                      </button>
                      <button
                        onClick={() => handleDelete(rt.id)}
                        style={{
                          background: 'none',
                          border: 'none',
                          cursor: 'pointer',
                          color: '#dc2626',
                        }}
                      >
                        <Trash2 size={16} />
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {showModal && (
        <RecurringExpenseModal
          categories={categories}
          expense={editTarget}
          onClose={() => setShowModal(false)}
          onSuccess={() => {
            setShowModal(false);
            loadData();
          }}
        />
      )}
    </>
  );
}

function RecurringExpenseModal({
  categories,
  expense,
  onClose,
  onSuccess,
}: {
  categories: ExpenseCategory[];
  expense: RecurringExpense | null;
  onClose: () => void;
  onSuccess: () => void;
}) {
  const { selectedStore } = useAuth();
  const [saving, setSaving] = useState(false);

  const [form, setForm] = useState({
    name: expense?.name || '',
    category_id: expense?.category_id || (categories.length > 0 ? categories[0].id : ''),
    amount: expense ? formatNumberInput(expense.amount) : '',
    interval: expense?.interval || 'monthly',
    interval_value: expense?.interval_value ? String(expense.interval_value) : '1',
    start_date: expense?.start_date || new Date().toISOString().slice(0, 10),
    end_date: expense?.end_date || '',
    is_active: expense ? expense.is_active : true,
    notes: expense?.notes || '',
  });

  const [err, setErr] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedStore) return;

    const amt = parseNumberInput(form.amount);
    if (!amt || amt <= 0) {
      setErr('Jumlah tidak valid');
      return;
    }
    const intVal = parseInt(form.interval_value, 10);
    if (isNaN(intVal) || intVal < 1) {
      setErr('Nilai siklus tidak valid');
      return;
    }

    setSaving(true);
    setErr('');
    try {
      const payload = {
        ...form,
        amount: amt,
        interval_value: intVal,
        end_date: form.end_date || undefined,
      };

      if (expense) {
        await recurringExpensesApi.update(selectedStore.store_id, expense.id, payload);
      } else {
        await recurringExpensesApi.create(selectedStore.store_id, payload);
      }
      onSuccess();
    } catch (error: any) {
      setErr(error.message || 'Gagal menyimpan templat rutin');
    } finally {
      setSaving(false);
    }
  };

  return (
    <div
      style={{
        position: 'fixed',
        inset: 0,
        zIndex: 100,
        background: 'rgba(0,0,0,0.5)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
      }}
    >
      <div
        style={{
          background: 'var(--bg-card)',
          padding: 24,
          borderRadius: 12,
          width: 500,
          maxWidth: '90%',
          maxHeight: '90vh',
          overflowY: 'auto',
        }}
      >
        <div
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            marginBottom: 20,
          }}
        >
          <h2 style={{ margin: 0, fontSize: '1.2rem' }}>
            {expense ? 'Ubah Pengeluaran Rutin' : 'Buat Templat Rutin'}
          </h2>
          <button
            onClick={onClose}
            style={{
              background: 'none',
              border: 'none',
              cursor: 'pointer',
              color: 'var(--text-3)',
            }}
          >
            <X size={20} />
          </button>
        </div>

        {err && (
          <div
            style={{
              background: '#fee2e2',
              color: '#b91c1c',
              padding: 12,
              borderRadius: 6,
              fontSize: '0.85rem',
              marginBottom: 16,
            }}
          >
            {err}
          </div>
        )}

        <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
          <div>
            <label className="label">Nama / Judul</label>
            <input
              type="text"
              className="input"
              style={{ width: '100%' }}
              value={form.name}
              onChange={e => setForm({ ...form, name: e.target.value })}
              placeholder="Contoh: Sewa Ruko Bulan"
              required
            />
          </div>

          <div style={{ display: 'flex', gap: 16 }}>
            <div style={{ flex: 1 }}>
              <label className="label">Kategori</label>
              <select
                className="input"
                style={{ width: '100%' }}
                value={form.category_id}
                onChange={e => setForm({ ...form, category_id: e.target.value })}
                required
              >
                {categories.map(c => (
                  <option key={c.id} value={c.id}>
                    {c.name}
                  </option>
                ))}
              </select>
            </div>

            <div style={{ flex: 1 }}>
              <label className="label">Jumlah (Rp)</label>
              <input
                type="text"
                className="input"
                style={{ width: '100%', fontSize: '1.1rem', fontWeight: 600, color: '#dc2626' }}
                value={form.amount}
                onChange={e =>
                  setForm({ ...form, amount: formatNumberInput(parseNumberInput(e.target.value)) })
                }
                placeholder="0"
                required
              />
            </div>
          </div>

          <div style={{ display: 'flex', gap: 16 }}>
            <div style={{ flex: 1 }}>
              <label className="label">Siklus</label>
              <select
                className="input"
                style={{ width: '100%' }}
                value={form.interval}
                onChange={e => setForm({ ...form, interval: e.target.value as any })}
                required
              >
                <option value="daily">Harian</option>
                <option value="weekly">Mingguan</option>
                <option value="monthly">Bulanan</option>
                <option value="yearly">Tahunan</option>
              </select>
            </div>
            <div style={{ flex: 1 }}>
              <label className="label">Interval Siklus (Angka)</label>
              <input
                type="number"
                className="input"
                style={{ width: '100%' }}
                value={form.interval_value}
                onChange={e => setForm({ ...form, interval_value: e.target.value })}
                min="1"
                required
              />
            </div>
          </div>

          <div style={{ display: 'flex', gap: 16 }}>
            <div style={{ flex: 1 }}>
              <label className="label">Tanggal Mulai Run</label>
              <input
                type="date"
                className="input"
                style={{ width: '100%' }}
                value={form.start_date}
                onChange={e => setForm({ ...form, start_date: e.target.value })}
                required
              />
            </div>
            <div style={{ flex: 1 }}>
              <label className="label">Tanggal Berakhir (Opsional)</label>
              <input
                type="date"
                className="input"
                style={{ width: '100%' }}
                value={form.end_date}
                onChange={e => setForm({ ...form, end_date: e.target.value })}
              />
            </div>
          </div>

          {expense && (
            <div>
              <label style={{ display: 'flex', alignItems: 'center', gap: 8, cursor: 'pointer' }}>
                <input
                  type="checkbox"
                  checked={form.is_active}
                  onChange={e => setForm({ ...form, is_active: e.target.checked })}
                />
                Templat Aktif
              </label>
              <p style={{ margin: '4px 0 0', fontSize: '0.8rem', color: 'var(--text-3)' }}>
                Jika tidak aktif, sistem tidak akan membuat tagihan ini secara otomatis.
              </p>
            </div>
          )}

          <div>
            <label className="label">Catatan (Opsional)</label>
            <textarea
              className="input"
              style={{ width: '100%', height: 80, resize: 'vertical' }}
              value={form.notes}
              onChange={e => setForm({ ...form, notes: e.target.value })}
            />
          </div>

          <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 12, marginTop: 8 }}>
            <button type="button" className="btn btn-secondary" onClick={onClose}>
              Batal
            </button>
            <button type="submit" className="btn btn-primary" disabled={saving}>
              {saving ? <Loader2 size={16} className="loading-spin" /> : 'Simpan Templat'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

export default ExpensesPage;
