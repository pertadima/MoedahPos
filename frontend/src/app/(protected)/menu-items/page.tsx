'use client';

import { useEffect, useState, useCallback } from 'react';
import {
  UtensilsCrossed, Plus, Pencil, Trash2, Check, X, Loader2,
  ChevronDown, ChevronUp, AlertTriangle, FlaskConical, Package,
} from 'lucide-react';
import { useAuth } from '@/lib/auth/AuthContext';
import { menuItemsApi, categoriesApi } from '@/lib/api/store-apis';
import { productsApi } from '@/lib/api/products';
import type { MenuItem, Category, Product } from '@/types';
import { formatRp } from '@/lib/utils';

// ── Types ─────────────────────────────────────────────────────────────────────

interface IngredientInput { productId: string; productName: string; unit: string; quantity: number }

interface FormState {
  name: string;
  description: string;
  category_id: string;
  sell_price: string;
  tax_rate: string;
  ingredients: IngredientInput[];
}

const emptyForm = (): FormState => ({
  name: '', description: '', category_id: '', sell_price: '', tax_rate: '0', ingredients: [],
});

// ── Page ──────────────────────────────────────────────────────────────────────

export default function MenuItemsPage() {
  const { selectedStore } = useAuth();
  const [items, setItems] = useState<MenuItem[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(true);
  const [modal, setModal] = useState<{ open: boolean; mode: 'create' | 'edit'; item?: MenuItem }>({ open: false, mode: 'create' });
  const [deleteConfirm, setDeleteConfirm] = useState<{ open: boolean; item?: MenuItem }>({ open: false });
  const [form, setForm] = useState<FormState>(emptyForm());
  const [formError, setFormError] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [expanded, setExpanded] = useState<string | null>(null);
  const [toast, setToast] = useState<{ msg: string; type: 'success' | 'error' } | null>(null);
  const [productSearch, setProductSearch] = useState('');

  const storeId = selectedStore?.store_id ?? '';

  const fetchAll = useCallback(async () => {
    if (!storeId) return;
    setLoading(true);
    try {
      const [menuRes, catRes, prodRes] = await Promise.all([
        menuItemsApi.list(storeId),
        categoriesApi.list(storeId),
        productsApi.list(storeId, { page: 1, per_page: 200 }),
      ]);
      setItems((menuRes.data as any).data ?? menuRes.data ?? []);
      setCategories((catRes.data as any).data ?? catRes.data ?? []);
      setProducts((prodRes.data as any).data?.data ?? []);
    } catch {
      showToast('Gagal memuat data', 'error');
    } finally {
      setLoading(false);
    }
  }, [storeId]);

  useEffect(() => { fetchAll(); }, [fetchAll]);

  const showToast = (msg: string, type: 'success' | 'error') => {
    setToast({ msg, type });
    setTimeout(() => setToast(null), 3500);
  };

  // ── Modal ──────────────────────────────────────────────────────────────────

  const openCreate = () => {
    setForm(emptyForm());
    setFormError('');
    setModal({ open: true, mode: 'create' });
  };

  const openEdit = (item: MenuItem) => {
    setForm({
      name: item.name, description: item.description,
      category_id: item.category_id ?? '',
      sell_price: String(item.sell_price),
      tax_rate: String(item.tax_rate),
      ingredients: item.ingredients.map(i => ({
        productId: i.product_id,
        productName: i.product_name,
        unit: i.unit,
        quantity: i.quantity,
      })),
    });
    setFormError('');
    setModal({ open: true, mode: 'edit', item });
  };

  const closeModal = () => { setModal({ open: false, mode: 'create' }); setProductSearch(''); };

  // ── Ingredients ────────────────────────────────────────────────────────────

  const addIngredient = (product: Product) => {
    if (form.ingredients.some(i => i.productId === product.id)) return;
    setForm(f => ({
      ...f,
      ingredients: [...f.ingredients, { productId: product.id, productName: product.name, unit: product.unit, quantity: 1 }],
    }));
    setProductSearch('');
  };

  const removeIngredient = (productId: string) => {
    setForm(f => ({ ...f, ingredients: f.ingredients.filter(i => i.productId !== productId) }));
  };

  const updateIngredientQty = (productId: string, qty: number) => {
    setForm(f => ({
      ...f,
      ingredients: f.ingredients.map(i => i.productId === productId ? { ...i, quantity: qty } : i),
    }));
  };

  const filteredProducts = products.filter(p =>
    p.name.toLowerCase().includes(productSearch.toLowerCase()) &&
    !form.ingredients.some(i => i.productId === p.id)
  ).slice(0, 8);

  // ── Submit ─────────────────────────────────────────────────────────────────

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!form.name.trim()) { setFormError('Nama menu wajib diisi'); return; }
    setSubmitting(true); setFormError('');
    const payload = {
      name: form.name.trim(),
      description: form.description,
      category_id: form.category_id || undefined,
      sell_price: parseFloat(form.sell_price) || 0,
      tax_rate: parseFloat(form.tax_rate) || 0,
      ingredients: form.ingredients.map(i => ({ product_id: i.productId, quantity: i.quantity })),
    };
    try {
      if (modal.mode === 'create') {
        await menuItemsApi.create(storeId, payload);
        showToast('Menu berhasil ditambahkan ✓', 'success');
      } else {
        await menuItemsApi.update(storeId, modal.item!.id, payload);
        showToast('Menu berhasil diperbarui ✓', 'success');
      }
      closeModal(); fetchAll();
    } catch (err: any) {
      setFormError(err?.response?.data?.message ?? 'Terjadi kesalahan');
    } finally {
      setSubmitting(false);
    }
  };

  const handleDelete = async () => {
    if (!deleteConfirm.item) return;
    try {
      await menuItemsApi.delete(storeId, deleteConfirm.item.id);
      showToast(`"${deleteConfirm.item.name}" dihapus`, 'success');
      setDeleteConfirm({ open: false }); fetchAll();
    } catch (err: any) {
      showToast(err?.response?.data?.message ?? 'Gagal menghapus', 'error');
    }
  };

  // ── Guard ──────────────────────────────────────────────────────────────────

  if (!selectedStore) return (
    <div style={{ padding: 32 }}>
      <div className="empty-state card" style={{ padding: 48 }}>
        <UtensilsCrossed size={40} style={{ color: 'var(--text-3)' }} />
        <p style={{ fontWeight: 600, color: 'var(--text-2)' }}>Pilih toko terlebih dahulu</p>
      </div>
    </div>
  );

  if (selectedStore.store_type !== 'restaurant') return (
    <div style={{ padding: 32 }}>
      <div className="empty-state card" style={{ padding: 48 }}>
        <UtensilsCrossed size={40} style={{ color: 'var(--text-3)' }} />
        <p style={{ fontWeight: 600, color: 'var(--text-2)' }}>Menu hanya tersedia untuk tipe Restoran</p>
      </div>
    </div>
  );

  // ── Render ─────────────────────────────────────────────────────────────────

  return (
    <div style={{ padding: 24, maxWidth: 900, margin: '0 auto' }}>
      {toast && (
        <div style={{
          position: 'fixed', top: 20, right: 20, zIndex: 9999,
          background: toast.type === 'success' ? 'rgba(16,185,129,0.15)' : 'rgba(239,68,68,0.15)',
          border: `1px solid ${toast.type === 'success' ? '#10b981' : '#ef4444'}`,
          color: toast.type === 'success' ? '#10b981' : '#ef4444',
          padding: '12px 20px', borderRadius: 10, fontWeight: 600, fontSize: '0.85rem',
          backdropFilter: 'blur(12px)', animation: 'slideIn 0.2s ease',
        }}>{toast.msg}</div>
      )}

      {/* Header */}
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 24 }}>
        <div>
          <h1 className="page-title" style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
            <UtensilsCrossed size={22} style={{ color: '#fb923c' }} />
            Menu Builder
          </h1>
          <p className="page-subtitle">{selectedStore.store_name} · {items.length} menu item</p>
        </div>
        <button className="btn btn-primary" onClick={openCreate} style={{ gap: 8 }}>
          <Plus size={16} /> Tambah Menu
        </button>
      </div>

      {/* Menu list */}
      {loading ? (
        <div style={{ display: 'flex', justifyContent: 'center', padding: 60 }}>
          <Loader2 size={28} className="loading-spin" style={{ color: 'var(--accent-em)' }} />
        </div>
      ) : items.length === 0 ? (
        <div className="empty-state card" style={{ padding: 60 }}>
          <UtensilsCrossed size={48} style={{ color: 'var(--text-3)' }} />
          <p style={{ fontWeight: 600, color: 'var(--text-2)' }}>Belum ada menu</p>
          <p style={{ fontSize: '0.85rem' }}>Klik "Tambah Menu" untuk membuat menu pertama.</p>
        </div>
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
          {items.map(item => {
            const isExpanded = expanded === item.id;
            return (
              <div key={item.id} className="card" style={{ padding: 0, overflow: 'hidden' }}>
                <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '14px 18px' }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 12, flex: 1, minWidth: 0 }}>
                    <div style={{
                      width: 40, height: 40, borderRadius: 10, flexShrink: 0,
                      background: 'linear-gradient(135deg, rgba(251,146,60,0.2), rgba(239,68,68,0.1))',
                      display: 'flex', alignItems: 'center', justifyContent: 'center',
                    }}>
                      <UtensilsCrossed size={18} style={{ color: '#fb923c' }} />
                    </div>
                    <div style={{ minWidth: 0 }}>
                      <div style={{ fontWeight: 600, fontSize: '0.9rem', color: 'var(--text-1)' }}>{item.name}</div>
                      <div style={{ fontSize: '0.74rem', color: 'var(--text-3)' }}>
                        {item.category_name && <span>{item.category_name} · </span>}
                        {item.ingredients.length} bahan · {formatRp(item.sell_price)}
                      </div>
                    </div>
                  </div>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
                    <button className="btn btn-ghost btn-sm" onClick={() => setExpanded(isExpanded ? null : item.id)} style={{ padding: '5px 8px' }}>
                      {isExpanded ? <ChevronUp size={14} /> : <ChevronDown size={14} />}
                    </button>
                    <button className="btn btn-ghost btn-sm" onClick={() => openEdit(item)} style={{ padding: '5px 8px' }}>
                      <Pencil size={14} />
                    </button>
                    <button className="btn btn-ghost btn-sm" onClick={() => setDeleteConfirm({ open: true, item })} style={{ padding: '5px 8px', color: 'rgba(239,68,68,0.7)' }}>
                      <Trash2 size={14} />
                    </button>
                  </div>
                </div>

                {/* Expanded: ingredients */}
                {isExpanded && (
                  <div style={{ borderTop: '1px solid var(--border)', padding: '12px 18px', background: 'var(--bg-base)' }}>
                    <div style={{ fontSize: '0.75rem', fontWeight: 600, color: 'var(--text-3)', marginBottom: 8, textTransform: 'uppercase', letterSpacing: '0.05em' }}>
                      Komposisi Bahan
                    </div>
                    {item.ingredients.length === 0 ? (
                      <div style={{ fontSize: '0.82rem', color: 'var(--text-3)', fontStyle: 'italic' }}>Belum ada bahan terdaftar</div>
                    ) : (
                      <div style={{ display: 'flex', flexWrap: 'wrap', gap: 6 }}>
                        {item.ingredients.map(ing => (
                          <div key={ing.id} style={{
                            display: 'flex', alignItems: 'center', gap: 5, padding: '4px 10px',
                            background: 'var(--bg-elevated)', borderRadius: 6, border: '1px solid var(--border)',
                            fontSize: '0.78rem', color: 'var(--text-2)',
                          }}>
                            <FlaskConical size={11} style={{ color: '#fb923c' }} />
                            {ing.product_name} — {ing.quantity} {ing.unit}
                          </div>
                        ))}
                      </div>
                    )}
                    {item.description && (
                      <div style={{ marginTop: 8, fontSize: '0.8rem', color: 'var(--text-3)', fontStyle: 'italic' }}>
                        {item.description}
                      </div>
                    )}
                  </div>
                )}
              </div>
            );
          })}
        </div>
      )}

      {/* ── Create / Edit Modal ─────────────────────────────────────────────── */}
      {modal.open && (
        <div style={{
          position: 'fixed', inset: 0, zIndex: 1000, background: 'rgba(0,0,0,0.65)',
          backdropFilter: 'blur(4px)', display: 'flex', alignItems: 'center', justifyContent: 'center', padding: 16,
        }}>
          <div className="card" style={{ width: '100%', maxWidth: 560, padding: 28, maxHeight: '90vh', overflowY: 'auto', animation: 'slideIn 0.2s ease' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 20 }}>
              <div style={{ fontWeight: 700, fontSize: '1rem', display: 'flex', alignItems: 'center', gap: 8 }}>
                <UtensilsCrossed size={18} style={{ color: '#fb923c' }} />
                {modal.mode === 'create' ? 'Tambah Menu Item' : `Edit: ${modal.item?.name}`}
              </div>
              <button onClick={closeModal} className="btn btn-ghost btn-sm" style={{ padding: 6 }}><X size={16} /></button>
            </div>

            <form onSubmit={handleSubmit}>
              {/* Name */}
              <div style={{ marginBottom: 14 }}>
                <label style={{ display: 'block', fontSize: '0.8rem', fontWeight: 600, marginBottom: 5, color: 'var(--text-2)' }}>
                  Nama Menu <span style={{ color: '#ef4444' }}>*</span>
                </label>
                <input className="input" placeholder="cth. Nasi Goreng Spesial" autoFocus
                  value={form.name} onChange={e => setForm(f => ({ ...f, name: e.target.value }))} />
              </div>

              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12, marginBottom: 14 }}>
                {/* Sell price */}
                <div>
                  <label style={{ display: 'block', fontSize: '0.8rem', fontWeight: 600, marginBottom: 5, color: 'var(--text-2)' }}>
                    Harga Jual (Rp)
                  </label>
                  <input className="input" type="number" min={0} placeholder="30000"
                    value={form.sell_price} onChange={e => setForm(f => ({ ...f, sell_price: e.target.value }))} />
                </div>
                {/* Tax */}
                <div>
                  <label style={{ display: 'block', fontSize: '0.8rem', fontWeight: 600, marginBottom: 5, color: 'var(--text-2)' }}>
                    Pajak (%)
                  </label>
                  <input className="input" type="number" min={0} max={100} placeholder="10"
                    value={form.tax_rate} onChange={e => setForm(f => ({ ...f, tax_rate: e.target.value }))} />
                </div>
              </div>

              {/* Category */}
              <div style={{ marginBottom: 14 }}>
                <label style={{ display: 'block', fontSize: '0.8rem', fontWeight: 600, marginBottom: 5, color: 'var(--text-2)' }}>
                  Kategori <span style={{ fontSize: '0.72rem', color: 'var(--text-3)', fontWeight: 400 }}>(opsional)</span>
                </label>
                <select className="input" value={form.category_id} onChange={e => setForm(f => ({ ...f, category_id: e.target.value }))}>
                  <option value="">— Tidak ada —</option>
                  {categories.map(c => <option key={c.id} value={c.id}>{c.name}</option>)}
                </select>
              </div>

              {/* Description */}
              <div style={{ marginBottom: 18 }}>
                <label style={{ display: 'block', fontSize: '0.8rem', fontWeight: 600, marginBottom: 5, color: 'var(--text-2)' }}>
                  Deskripsi
                </label>
                <textarea className="input" rows={2} placeholder="Deskripsi singkat menu…"
                  value={form.description} onChange={e => setForm(f => ({ ...f, description: e.target.value }))}
                  style={{ resize: 'vertical', minHeight: 56 }} />
              </div>

              {/* Ingredients */}
              <div style={{ marginBottom: 18 }}>
                <div style={{ fontSize: '0.8rem', fontWeight: 700, color: 'var(--text-1)', marginBottom: 10, display: 'flex', alignItems: 'center', gap: 6 }}>
                  <FlaskConical size={14} style={{ color: '#fb923c' }} />
                  Komposisi Bahan
                  <span style={{ fontSize: '0.7rem', color: 'var(--text-3)', fontWeight: 400 }}>— stok yang dipakai saat menu ini dijual</span>
                </div>

                {/* Added ingredients */}
                {form.ingredients.length > 0 && (
                  <div style={{ display: 'flex', flexDirection: 'column', gap: 6, marginBottom: 10 }}>
                    {form.ingredients.map(ing => (
                      <div key={ing.productId} style={{
                        display: 'flex', alignItems: 'center', gap: 10, padding: '8px 12px',
                        background: 'var(--bg-elevated)', borderRadius: 8, border: '1px solid var(--border)',
                      }}>
                        <Package size={13} style={{ color: '#fb923c', flexShrink: 0 }} />
                        <span style={{ flex: 1, fontSize: '0.82rem', color: 'var(--text-1)' }}>{ing.productName}</span>
                        <input
                          type="number" min={0.01} step={0.01}
                          value={ing.quantity}
                          onChange={e => updateIngredientQty(ing.productId, parseFloat(e.target.value) || 0)}
                          style={{ width: 70, background: 'var(--bg-base)', border: '1px solid var(--border-md)', borderRadius: 6, padding: '3px 7px', color: 'var(--text-1)', fontSize: '0.8rem', textAlign: 'right' }}
                        />
                        <span style={{ fontSize: '0.75rem', color: 'var(--text-3)', minWidth: 32 }}>{ing.unit}</span>
                        <button type="button" onClick={() => removeIngredient(ing.productId)} style={{ background: 'none', border: 'none', cursor: 'pointer', color: '#ef4444', padding: 0 }}>
                          <X size={14} />
                        </button>
                      </div>
                    ))}
                  </div>
                )}

                {/* Product search */}
                <div style={{ position: 'relative' }}>
                  <input className="input" placeholder="Cari produk/bahan untuk ditambahkan…"
                    value={productSearch} onChange={e => setProductSearch(e.target.value)} />
                  {productSearch && filteredProducts.length > 0 && (
                    <div style={{
                      position: 'absolute', top: '100%', left: 0, right: 0, zIndex: 100,
                      background: 'var(--bg-card)', border: '1px solid var(--border-md)', borderRadius: 8,
                      marginTop: 4, boxShadow: '0 8px 24px rgba(0,0,0,0.3)', overflow: 'hidden',
                    }}>
                      {filteredProducts.map(p => (
                        <button key={p.id} type="button" onClick={() => addIngredient(p)}
                          style={{ display: 'flex', width: '100%', alignItems: 'center', gap: 10, padding: '9px 14px', background: 'none', border: 'none', cursor: 'pointer', textAlign: 'left' }}
                          className="category-row"
                        >
                          <Package size={13} style={{ color: 'var(--text-3)', flexShrink: 0 }} />
                          <span style={{ flex: 1, fontSize: '0.82rem', color: 'var(--text-1)' }}>{p.name}</span>
                          <span style={{ fontSize: '0.72rem', color: 'var(--text-3)' }}>{p.unit}</span>
                        </button>
                      ))}
                    </div>
                  )}
                </div>
              </div>

              {formError && (
                <div style={{ background: 'rgba(239,68,68,0.12)', border: '1px solid rgba(239,68,68,0.3)', borderRadius: 8, padding: '10px 14px', marginBottom: 14, fontSize: '0.82rem', color: '#ef4444' }}>
                  {formError}
                </div>
              )}

              <div style={{ display: 'flex', gap: 10, justifyContent: 'flex-end' }}>
                <button type="button" className="btn btn-ghost" onClick={closeModal}>Batal</button>
                <button type="submit" className="btn btn-primary" disabled={submitting} style={{ gap: 8 }}>
                  {submitting ? <><Loader2 size={14} className="loading-spin" /> Menyimpan...</> : <><Check size={14} /> Simpan</>}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* ── Delete Confirm ─────────────────────────────────────────────────── */}
      {deleteConfirm.open && deleteConfirm.item && (
        <div style={{ position: 'fixed', inset: 0, zIndex: 1000, background: 'rgba(0,0,0,0.6)', backdropFilter: 'blur(4px)', display: 'flex', alignItems: 'center', justifyContent: 'center', padding: 16 }}>
          <div className="card" style={{ width: '100%', maxWidth: 380, padding: 28, textAlign: 'center', animation: 'slideIn 0.2s ease' }}>
            <div style={{ width: 48, height: 48, borderRadius: '50%', background: 'rgba(239,68,68,0.15)', display: 'flex', alignItems: 'center', justifyContent: 'center', margin: '0 auto 14px' }}>
              <AlertTriangle size={22} style={{ color: '#ef4444' }} />
            </div>
            <div style={{ fontWeight: 700, fontSize: '1rem', marginBottom: 8 }}>Hapus Menu?</div>
            <div style={{ fontSize: '0.85rem', color: 'var(--text-2)', marginBottom: 20, lineHeight: 1.6 }}>
              Menu <strong>"{deleteConfirm.item.name}"</strong> akan dihapus (soft-delete).
            </div>
            <div style={{ display: 'flex', gap: 10 }}>
              <button className="btn btn-ghost" onClick={() => setDeleteConfirm({ open: false })} style={{ flex: 1 }}>Batal</button>
              <button className="btn" onClick={handleDelete} style={{ flex: 1, gap: 8, background: 'rgba(239,68,68,0.15)', color: '#ef4444', border: '1px solid rgba(239,68,68,0.3)' }}>
                <Trash2 size={14} /> Hapus
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
