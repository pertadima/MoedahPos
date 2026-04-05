'use client';

import { useEffect, useState, useCallback } from 'react';
import {
  Wallet,
  Plus,
  Loader2,
  X,
  CreditCard,
  Calendar,
  MoreVertical,
  Edit2,
  Trash2,
  Tag,
  AlignLeft,
} from 'lucide-react';
import { useAuth } from '@/lib/auth/AuthContext';
import { expensesApi } from '@/lib/api/store-apis';
import { formatRp, formatDate, parseNumberInput, formatNumberInput } from '@/lib/utils';
import type { Expense, ExpenseCategory } from '@/types';

function ExpensesPage() {
  const { selectedStore } = useAuth();
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

  const handleFilterChange = (key: string, value: string) => {
    setFilter(prev => ({ ...prev, [key]: value }));
    setMeta(prev => ({ ...prev, page: 1 }));
  };

  return (
    <div style={{ padding: '24px 32px', maxWidth: 1200, margin: '0 auto' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <div>
          <h1 style={{ fontSize: '1.4rem', fontWeight: 700, margin: '0 0 4px', color: 'var(--text-1)' }}>
            Pengeluaran Operasional
          </h1>
          <p style={{ margin: 0, fontSize: '0.85rem', color: 'var(--text-3)' }}>
             Catat dan kelola semua pengeluaran selain stok produk.
          </p>
        </div>
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
      </div>

      <div style={{
        background: 'var(--bg-card)',
        padding: 16,
        borderRadius: 12,
        border: '1px solid var(--border)',
        marginBottom: 20,
        display: 'flex',
        gap: 16,
        alignItems: 'flex-end',
        flexWrap: 'wrap'
      }}>
        <div style={{ flex: '1 1 180px' }}>
          <label className="label">Kategori</label>
          <select 
            className="input" 
            style={{ width: '100%', height: 38 }}
            value={filter.category_id}
            onChange={e => handleFilterChange('category_id', e.target.value)}
          >
            <option value="">Semua Kategori</option>
            {categories.map(c => <option key={c.id} value={c.id}>{c.name}</option>)}
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
        <div style={{
          textAlign: 'center', padding: '60px 20px', background: 'var(--bg-card)',
          borderRadius: 12, border: '1px dashed var(--border)'
        }}>
          <Wallet size={48} style={{ color: 'var(--text-4)', margin: '0 auto 16px' }} />
          <h3 style={{ margin: '0 0 8px', fontSize: '1.1rem', color: 'var(--text-1)' }}>Belum Ada Pengeluaran</h3>
          <p style={{ margin: 0, color: 'var(--text-3)', fontSize: '0.9rem' }}>
            {filter.category_id || filter.date_from ? 'Tidak ada data yang cocok dengan filter.' : 'Mulai catat sewa, gaji, atau tagihan utilitas Anda.'}
          </p>
        </div>
      ) : (
        <div style={{ background: 'var(--bg-card)', borderRadius: 12, border: '1px solid var(--border)', overflow: 'hidden' }}>
          <div style={{ overflowX: 'auto' }}>
            <table className="tbl" style={{ width: '100%', minWidth: 600 }}>
              <thead>
                <tr>
                  <th>Tanggal</th>
                  <th>Kategori</th>
                  <th>Catatan</th>
                  <th style={{ textAlign: 'right' }}>Jumlah</th>
                  <th style={{ textAlign: 'right', width: 100 }}>Aksi</th>
                </tr>
              </thead>
              <tbody>
                {expenses.map(exp => (
                  <tr key={exp.id}>
                    <td>{formatDate(exp.expense_date)}</td>
                    <td>
                      <span style={{ 
                        background: 'var(--bg-hover)', 
                        padding: '4px 8px', 
                        borderRadius: 6, 
                        fontSize: '0.75rem',
                        fontWeight: 600
                      }}>
                        {exp.category_name}
                      </span>
                    </td>
                    <td style={{ color: 'var(--text-2)' }}>{exp.notes || '—'}</td>
                    <td style={{ textAlign: 'right', fontWeight: 600, color: '#dc2626' }}>
                      {formatRp(exp.amount)}
                    </td>
                    <td style={{ textAlign: 'right' }}>
                      <button
                        onClick={() => {
                          setEditTarget(exp);
                          setShowModal(true);
                        }}
                        style={{ background: 'none', border: 'none', cursor: 'pointer', color: 'var(--text-3)', marginRight: 10 }}
                      >
                        <Edit2 size={16} />
                      </button>
                      <button
                        onClick={() => handleDelete(exp.id)}
                        style={{ background: 'none', border: 'none', cursor: 'pointer', color: '#dc2626' }}
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
            <div style={{ padding: 16, borderTop: '1px solid var(--border)', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
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

      {showModal && (
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
  onSuccess 
}: { 
  categories: ExpenseCategory[], 
  expense: Expense | null, 
  onClose: () => void, 
  onSuccess: () => void 
}) {
  const { selectedStore } = useAuth();
  const [saving, setSaving] = useState(false);
  
  const [form, setForm] = useState({
    category_id: expense?.category_id || (categories.length > 0 ? categories[0].id : ''),
    amount: expense ? formatNumberInput(expense.amount) : '',
    expense_date: expense?.expense_date || new Date().toISOString().slice(0, 10),
    notes: expense?.notes || ''
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
        amount: amt
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
    <div style={{ position: 'fixed', inset: 0, zIndex: 100, background: 'rgba(0,0,0,0.5)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
      <div style={{ background: 'var(--bg-card)', padding: 24, borderRadius: 12, width: 400, maxWidth: '90%' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 20 }}>
          <h2 style={{ margin: 0, fontSize: '1.2rem' }}>{expense ? 'Ubah Pengeluaran' : 'Tambah Pengeluaran'}</h2>
          <button onClick={onClose} style={{ background: 'none', border: 'none', cursor: 'pointer', color: 'var(--text-3)' }}>
            <X size={20} />
          </button>
        </div>
        
        {err && <div style={{ color: '#dc2626', fontSize: '0.85rem', marginBottom: 16 }}>{err}</div>}

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
              {categories.map(c => <option key={c.id} value={c.id}>{c.name}</option>)}
            </select>
          </div>
          
          <div>
            <label className="label">Jumlah (Rp)</label>
            <input 
              type="text" 
              className="input" 
              style={{ width: '100%', fontSize: '1.1rem', fontWeight: 600, color: '#dc2626' }}
              value={form.amount}
              onChange={e => setForm({ ...form, amount: formatNumberInput(parseNumberInput(e.target.value)) })}
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
            <button type="button" className="btn btn-secondary" onClick={onClose}>Batal</button>
            <button type="submit" className="btn btn-primary" disabled={saving}>
              {saving ? <Loader2 size={16} className="loading-spin" /> : 'Simpan'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

export default ExpensesPage;
