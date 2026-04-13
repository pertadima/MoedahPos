'use client';

import { useEffect, useState, useCallback, useRef } from 'react';
import {
  UtensilsCrossed,
  Plus,
  Pencil,
  Trash2,
  Check,
  X,
  Loader2,
  ChevronDown,
  ChevronUp,
  AlertTriangle,
  FlaskConical,
  Package,
} from 'lucide-react';
import { useAuth } from '@/lib/auth/AuthContext';
import { usePermission } from '@/hooks/usePermission';
import { menuItemsApi, categoriesApi } from '@/lib/api/store-apis';
import { productsApi } from '@/lib/api/products';
import type { MenuItem, Category, Product } from '@/types';
import { formatRp, getErrorMessage } from '@/lib/utils';

// ── Types ─────────────────────────────────────────────────────────────────────

interface IngredientInput {
  productId: string;
  productName: string;
  unit: string;
  quantity: number;
}

interface FormState {
  name: string;
  description: string;
  category_id: string;
  tax_rate: string;
  packaging_cost: string;
  overhead_cost: string;
  labor_cost: string;
  ingredients: IngredientInput[];
}

const emptyForm = (): FormState => ({
  name: '',
  description: '',
  category_id: '',
  sell_price: '',
  tax_rate: '0',
  packaging_cost: '0',
  overhead_cost: '0',
  labor_cost: '0',
  ingredients: [],
});

// ── Page ──────────────────────────────────────────────────────────────────────

export default function MenuItemsPage() {
  const { selectedStore } = useAuth();
  const { can } = usePermission();
  const [items, setItems] = useState<MenuItem[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(true);
  const [modal, setModal] = useState<{ open: boolean; mode: 'create' | 'edit'; item?: MenuItem }>({
    open: false,
    mode: 'create',
  });
  const [deleteConfirm, setDeleteConfirm] = useState<{ open: boolean; item?: MenuItem }>({
    open: false,
  });
  const [form, setForm] = useState<FormState>(emptyForm());
  const [formError, setFormError] = useState('');
  const prevHppRef = useRef<number>(0);
  const [submitting, setSubmitting] = useState(false);
  const [expanded, setExpanded] = useState<string | null>(null);
  const [toast, setToast] = useState<{ msg: string; type: 'success' | 'error' } | null>(null);
  const [productSearch, setProductSearch] = useState('');

  const storeId = selectedStore?.store_id ?? '';

  const showToast = useCallback((msg: string, type: 'success' | 'error') => {
    setToast({ msg, type });
    setTimeout(() => setToast(null), 3500);
  }, []);

  const fetchAll = useCallback(async () => {
    if (!storeId) return;
    setLoading(true);
    try {
      const [menuRes, catRes, prodRes] = await Promise.all([
        menuItemsApi.list(storeId),
        categoriesApi.list(storeId),
        productsApi.list(storeId, { page: 1, per_page: 200 }),
      ]);
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      setItems((menuRes.data as any).data ?? menuRes.data ?? []);
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      setCategories((catRes.data as any).data ?? catRes.data ?? []);

      // prodRes.data is PaginatedData<Product>, so its array is prodRes.data.data
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const paginatedProducts = prodRes.data as any;
      setProducts(paginatedProducts.data ?? []);
    } catch {
      showToast('Gagal memuat data', 'error');
    } finally {
      setLoading(false);
    }
  }, [storeId, showToast]);

  useEffect(() => {
    fetchAll();
  }, [fetchAll]);

  // ── Modal ──────────────────────────────────────────────────────────────────

  const openCreate = () => {
    setForm(emptyForm());
    setFormError('');
    setModal({ open: true, mode: 'create' });
  };

  const openEdit = (item: MenuItem) => {
    setForm({
      name: item.name,
      description: item.description,
      category_id: item.category_id ?? '',
      sell_price: String(item.sell_price),
      tax_rate: String(item.tax_rate),
      packaging_cost: String(item.packaging_cost ?? 0),
      overhead_cost: String(item.overhead_cost ?? 0),
      labor_cost: String(item.labor_cost ?? 0),
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

  const closeModal = () => {
    setModal({ open: false, mode: 'create' });
    setProductSearch('');
  };

  // ── Ingredients ────────────────────────────────────────────────────────────

  const addIngredient = (product: Product) => {
    if (form.ingredients.some(i => i.productId === product.id)) return;
    setForm(f => ({
      ...f,
      ingredients: [
        ...f.ingredients,
        { productId: product.id, productName: product.name, unit: product.unit, quantity: 1 },
      ],
    }));
    setProductSearch('');
  };

  const removeIngredient = (productId: string) => {
    setForm(f => ({ ...f, ingredients: f.ingredients.filter(i => i.productId !== productId) }));
  };

  const updateIngredientQty = (productId: string, qty: number) => {
    setForm(f => ({
      ...f,
      ingredients: f.ingredients.map(i =>
        i.productId === productId ? { ...i, quantity: qty } : i
      ),
    }));
  };

  const filteredProducts = products
    .filter(
      p =>
        p.name.toLowerCase().includes(productSearch.toLowerCase()) &&
        !form.ingredients.some(i => i.productId === p.id)
    )
    .slice(0, 8);

  // ── Submit ─────────────────────────────────────────────────────────────────

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!form.name.trim()) {
      setFormError('Nama menu wajib diisi');
      return;
    }
    setSubmitting(true);
    setFormError('');
    const payload = {
      name: form.name.trim(),
      description: form.description,
      category_id: form.category_id || undefined,
      sell_price: parseFloat(form.sell_price) || 0,
      tax_rate: parseFloat(form.tax_rate) || 0,
      packaging_cost: parseFloat(form.packaging_cost) || 0,
      overhead_cost: parseFloat(form.overhead_cost) || 0,
      labor_cost: parseFloat(form.labor_cost) || 0,
      ingredients: form.ingredients.map(i => ({ product_id: i.productId, quantity: i.quantity })),
    };
    try {
      if (modal.mode === 'create') {
        await menuItemsApi.create(storeId, payload);
        showToast('Menu berhasil ditambahkan ✓', 'success');
      } else if (modal.item) {
        await menuItemsApi.update(storeId, modal.item.id, payload);
        showToast('Menu berhasil diperbarui ✓', 'success');
      } else {
        setFormError('Menu tidak ditemukan');
        return;
      }
      closeModal();
      fetchAll();
    } catch (error) {
      setFormError(getErrorMessage(error, 'Terjadi kesalahan'));
    } finally {
      setSubmitting(false);
    }
  };

  const handleDelete = async () => {
    if (!deleteConfirm.item) return;
    try {
      await menuItemsApi.delete(storeId, deleteConfirm.item.id);
      showToast(`"${deleteConfirm.item.name}" dihapus`, 'success');
      setDeleteConfirm({ open: false });
      fetchAll();
    } catch (error) {
      showToast(getErrorMessage(error, 'Gagal menghapus'), 'error');
    }
  };

  // ── Guard ──────────────────────────────────────────────────────────────────

  if (!selectedStore)
    return (
      <div style={{ padding: 32 }}>
        <div className="empty-state card" style={{ padding: 48 }}>
          <UtensilsCrossed size={40} style={{ color: 'var(--text-3)' }} />
          <p style={{ fontWeight: 600, color: 'var(--text-2)' }}>Pilih toko terlebih dahulu</p>
        </div>
      </div>
    );

  if (selectedStore.store_type !== 'restaurant')
    return (
      <div style={{ padding: 32 }}>
        <div className="empty-state card" style={{ padding: 48 }}>
          <UtensilsCrossed size={40} style={{ color: 'var(--text-3)' }} />
          <p style={{ fontWeight: 600, color: 'var(--text-2)' }}>
            Menu hanya tersedia untuk tipe Restoran
          </p>
        </div>
      </div>
    );

  // ── Render ─────────────────────────────────────────────────────────────────

  return (
    <div className="w-full p-6">
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
          marginBottom: 24,
        }}
      >
        <div>
          <h1 className="page-title" style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
            <UtensilsCrossed size={22} style={{ color: '#fb923c' }} />
            Menu Builder
          </h1>
          <p className="page-subtitle">
            {selectedStore.store_name} · {items.length} menu item
          </p>
        </div>
        {can('products.create') && (
          <button className="btn btn-primary" onClick={openCreate} style={{ gap: 8 }}>
            <Plus size={16} /> Tambah Menu
          </button>
        )}
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
          <p style={{ fontSize: '0.85rem' }}>
            Klik &ldquo;Tambah Menu&rdquo; untuk membuat menu pertama.
          </p>
        </div>
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
          {items.map(item => {
            const isExpanded = expanded === item.id;
            return (
              <div key={item.id} className="card" style={{ padding: 0, overflow: 'hidden' }}>
                <div
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'space-between',
                    padding: '14px 18px',
                  }}
                >
                  <div
                    style={{ display: 'flex', alignItems: 'center', gap: 12, flex: 1, minWidth: 0 }}
                  >
                    <div
                      style={{
                        width: 40,
                        height: 40,
                        borderRadius: 10,
                        flexShrink: 0,
                        background:
                          'linear-gradient(135deg, rgba(251,146,60,0.2), rgba(239,68,68,0.1))',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                      }}
                    >
                      <UtensilsCrossed size={18} style={{ color: '#fb923c' }} />
                    </div>
                    <div style={{ minWidth: 0 }}>
                      <div style={{ fontWeight: 600, fontSize: '0.9rem', color: 'var(--text-1)' }}>
                        {item.name}
                      </div>
                      <div style={{ fontSize: '0.74rem', color: 'var(--text-3)' }}>
                        {item.category_name && <span>{item.category_name} · </span>}
                        {item.ingredients.length} bahan · {formatRp(item.sell_price)}
                      </div>
                    </div>
                  </div>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
                    <button
                      className="btn btn-ghost btn-sm"
                      onClick={() => setExpanded(isExpanded ? null : item.id)}
                      style={{ padding: '5px 8px' }}
                    >
                      {isExpanded ? <ChevronUp size={14} /> : <ChevronDown size={14} />}
                    </button>
                    {can('products.update') && (
                      <button
                        className="btn btn-ghost btn-sm"
                        onClick={() => openEdit(item)}
                        style={{ padding: '5px 8px' }}
                      >
                        <Pencil size={14} />
                      </button>
                    )}
                    {can('products.delete') && (
                      <button
                        className="btn btn-ghost btn-sm"
                        onClick={() => setDeleteConfirm({ open: true, item })}
                        style={{ padding: '5px 8px', color: 'rgba(239,68,68,0.7)' }}
                      >
                        <Trash2 size={14} />
                      </button>
                    )}
                  </div>
                </div>

                {/* Expanded: ingredients */}
                {isExpanded && (
                  <div
                    style={{
                      borderTop: '1px solid var(--border)',
                      padding: '12px 18px',
                      background: 'var(--bg-base)',
                    }}
                  >
                    <div
                      style={{
                        fontSize: '0.75rem',
                        fontWeight: 600,
                        color: 'var(--text-3)',
                        marginBottom: 8,
                        textTransform: 'uppercase',
                        letterSpacing: '0.05em',
                      }}
                    >
                      Komposisi Bahan
                    </div>
                    {item.ingredients.length === 0 ? (
                      <div
                        style={{ fontSize: '0.82rem', color: 'var(--text-3)', fontStyle: 'italic' }}
                      >
                        Belum ada bahan terdaftar
                      </div>
                    ) : (
                      <div style={{ display: 'flex', flexWrap: 'wrap', gap: 6 }}>
                        {item.ingredients.map(ing => (
                          <div
                            key={ing.id}
                            style={{
                              display: 'flex',
                              alignItems: 'center',
                              gap: 5,
                              padding: '4px 10px',
                              background: 'var(--bg-elevated)',
                              borderRadius: 6,
                              border: '1px solid var(--border)',
                              fontSize: '0.78rem',
                              color: 'var(--text-2)',
                            }}
                          >
                            <FlaskConical size={11} style={{ color: '#fb923c' }} />
                            {ing.product_name} — {ing.quantity} {ing.unit}
                          </div>
                        ))}
                      </div>
                    )}
                    {item.description && (
                      <div
                        style={{
                          marginTop: 8,
                          fontSize: '0.8rem',
                          color: 'var(--text-3)',
                          fontStyle: 'italic',
                        }}
                      >
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
        <div
          style={{
            position: 'fixed',
            inset: 0,
            zIndex: 1000,
            background: 'rgba(0,0,0,0.65)',
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
              maxWidth: 560,
              padding: 28,
              maxHeight: '90vh',
              overflowY: 'auto',
              animation: 'slideIn 0.2s ease',
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
              <div
                style={{
                  fontWeight: 700,
                  fontSize: '1rem',
                  display: 'flex',
                  alignItems: 'center',
                  gap: 8,
                }}
              >
                <UtensilsCrossed size={18} style={{ color: '#fb923c' }} />
                {modal.mode === 'create' ? 'Tambah Menu Item' : `Edit: ${modal.item?.name}`}
              </div>
              <button onClick={closeModal} className="btn btn-ghost btn-sm" style={{ padding: 6 }}>
                <X size={16} />
              </button>
            </div>

            <form onSubmit={handleSubmit}>
              {/* Name */}
              <div style={{ marginBottom: 14 }}>
                <label
                  style={{
                    display: 'block',
                    fontSize: '0.8rem',
                    fontWeight: 600,
                    marginBottom: 5,
                    color: 'var(--text-2)',
                  }}
                >
                  Nama Menu <span style={{ color: '#ef4444' }}>*</span>
                </label>
                <input
                  className="input"
                  placeholder="cth. Nasi Goreng Spesial"
                  autoFocus
                  value={form.name}
                  onChange={e => setForm(f => ({ ...f, name: e.target.value }))}
                />
              </div>

              <div
                style={{
                  display: 'grid',
                  gridTemplateColumns: '1fr 1fr',
                  gap: 12,
                  marginBottom: 14,
                }}
              >
                {/* Sell price */}
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
                    Harga Jual (Rp)
                  </label>
                  <input
                    className="input"
                    type="text"
                    placeholder="0"
                    onFocus={e => e.target.select()}
                    value={form.sell_price ? Number(form.sell_price).toLocaleString('id-ID') : ''}
                    onChange={e => {
                      const val = e.target.value.replace(/[^0-9]/g, '');
                      setForm(f => ({ ...f, sell_price: val }));
                    }}
                  />
                </div>
                {/* Tax */}
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
                    Pajak (%)
                  </label>
                  <input
                    className="input"
                    type="number"
                    min={0}
                    max={100}
                    placeholder="10"
                    value={form.tax_rate}
                    onChange={e => setForm(f => ({ ...f, tax_rate: e.target.value }))}
                  />
                </div>
              </div>

              {/* Category */}
              <div style={{ marginBottom: 14 }}>
                <label
                  style={{
                    display: 'block',
                    fontSize: '0.8rem',
                    fontWeight: 600,
                    marginBottom: 5,
                    color: 'var(--text-2)',
                  }}
                >
                  Kategori{' '}
                  <span style={{ fontSize: '0.72rem', color: 'var(--text-3)', fontWeight: 400 }}>
                    (opsional)
                  </span>
                </label>
                <select
                  className="input"
                  value={form.category_id}
                  onChange={e => setForm(f => ({ ...f, category_id: e.target.value }))}
                >
                  <option value="">— Tidak ada —</option>
                  {categories.map(c => (
                    <option key={c.id} value={c.id}>
                      {c.name}
                    </option>
                  ))}
                </select>
              </div>

              {/* Description */}
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
                  Deskripsi
                </label>
                <textarea
                  className="input"
                  rows={2}
                  placeholder="Deskripsi singkat menu…"
                  value={form.description}
                  onChange={e => setForm(f => ({ ...f, description: e.target.value }))}
                  style={{ resize: 'vertical', minHeight: 56 }}
                />
              </div>

              {/* Ingredients */}
              <div style={{ marginBottom: 18 }}>
                <div
                  style={{
                    fontSize: '0.8rem',
                    fontWeight: 700,
                    color: 'var(--text-1)',
                    marginBottom: 10,
                    display: 'flex',
                    alignItems: 'center',
                    gap: 6,
                  }}
                >
                  <FlaskConical size={14} style={{ color: '#fb923c' }} />
                  Komposisi Bahan
                  <span style={{ fontSize: '0.7rem', color: 'var(--text-3)', fontWeight: 400 }}>
                    — stok yang dipakai saat menu ini dijual
                  </span>
                </div>

                {/* Added ingredients */}
                {form.ingredients.length > 0 && (
                  <div
                    style={{ display: 'flex', flexDirection: 'column', gap: 6, marginBottom: 10 }}
                  >
                    {form.ingredients.map(ing => (
                      <div
                        key={ing.productId}
                        style={{
                          display: 'flex',
                          alignItems: 'center',
                          gap: 10,
                          padding: '8px 12px',
                          background: 'var(--bg-elevated)',
                          borderRadius: 8,
                          border: '1px solid var(--border)',
                        }}
                      >
                        <Package size={13} style={{ color: '#fb923c', flexShrink: 0 }} />
                        <div style={{ flex: 1, display: 'flex', flexDirection: 'column' }}>
                          <span style={{ fontSize: '0.82rem', color: 'var(--text-1)' }}>
                            {ing.productName}
                          </span>
                          {(() => {
                            const p = products.find(x => x.id === ing.productId);
                            const cost = (p?.cost_price || 0) * ing.quantity;
                            return (
                              <span style={{ fontSize: '0.7rem', color: 'var(--text-3)' }}>
                                {formatRp(p?.cost_price || 0)}/{ing.unit} = {formatRp(cost)}
                              </span>
                            );
                          })()}
                        </div>
                        <input
                          type="number"
                          min={0.01}
                          step={0.01}
                          value={ing.quantity}
                          onChange={e =>
                            updateIngredientQty(ing.productId, parseFloat(e.target.value) || 0)
                          }
                          style={{
                            width: 70,
                            background: 'var(--bg-base)',
                            border: '1px solid var(--border-md)',
                            borderRadius: 6,
                            padding: '3px 7px',
                            color: 'var(--text-1)',
                            fontSize: '0.8rem',
                            textAlign: 'right',
                          }}
                        />
                        <span style={{ fontSize: '0.75rem', color: 'var(--text-3)', minWidth: 32 }}>
                          {ing.unit}
                        </span>
                        <button
                          type="button"
                          onClick={() => removeIngredient(ing.productId)}
                          style={{
                            background: 'none',
                            border: 'none',
                            cursor: 'pointer',
                            color: '#ef4444',
                            padding: 0,
                          }}
                        >
                          <X size={14} />
                        </button>
                      </div>
                    ))}
                  </div>
                )}

                {/* Product search */}
                <div style={{ position: 'relative' }}>
                  <input
                    className="input"
                    placeholder="Cari produk/bahan untuk ditambahkan…"
                    value={productSearch}
                    onChange={e => setProductSearch(e.target.value)}
                  />
                  {productSearch && filteredProducts.length > 0 && (
                    <div
                      style={{
                        position: 'absolute',
                        top: '100%',
                        left: 0,
                        right: 0,
                        zIndex: 100,
                        background: 'var(--bg-card)',
                        border: '1px solid var(--border-md)',
                        borderRadius: 8,
                        marginTop: 4,
                        boxShadow: '0 8px 24px rgba(0,0,0,0.3)',
                        overflow: 'hidden',
                      }}
                    >
                      {filteredProducts.map(p => (
                        <button
                          key={p.id}
                          type="button"
                          onClick={() => addIngredient(p)}
                          style={{
                            display: 'flex',
                            width: '100%',
                            alignItems: 'center',
                            gap: 10,
                            padding: '9px 14px',
                            background: 'none',
                            border: 'none',
                            cursor: 'pointer',
                            textAlign: 'left',
                          }}
                          className="category-row"
                        >
                          <Package size={13} style={{ color: 'var(--text-3)', flexShrink: 0 }} />
                          <span style={{ flex: 1, fontSize: '0.82rem', color: 'var(--text-1)' }}>
                            {p.name}
                          </span>
                          <span style={{ fontSize: '0.72rem', color: 'var(--text-3)' }}>
                            {p.unit}
                          </span>
                        </button>
                      ))}
                    </div>
                  )}
                </div>

                {/* Biaya Tambahan */}
                <div style={{ marginBottom: 18 }}>
                  <div
                    style={{
                      fontSize: '0.8rem',
                      fontWeight: 700,
                      color: 'var(--text-1)',
                      marginBottom: 10,
                    }}
                  >
                    Biaya Tambahan
                  </div>
                  <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: 12 }}>
                    <div>
                      <label
                        style={{
                          display: 'block',
                          fontSize: '0.75rem',
                          fontWeight: 600,
                          marginBottom: 5,
                          color: 'var(--text-2)',
                        }}
                      >
                        Packaging
                      </label>
                      <input
                        className="input"
                        type="text"
                        placeholder="0"
                        onFocus={e => e.target.select()}
                        value={form.packaging_cost ? Number(form.packaging_cost).toLocaleString('id-ID') : ''}
                        onChange={e => {
                          const val = e.target.value.replace(/[^0-9]/g, '');
                          setForm(f => ({ ...f, packaging_cost: val }));
                        }}
                      />
                    </div>
                    <div>
                      <label
                        style={{
                          display: 'block',
                          fontSize: '0.75rem',
                          fontWeight: 600,
                          marginBottom: 5,
                          color: 'var(--text-2)',
                        }}
                      >
                        Overhead
                      </label>
                      <input
                        className="input"
                        type="text"
                        placeholder="0"
                        onFocus={e => e.target.select()}
                        value={form.overhead_cost ? Number(form.overhead_cost).toLocaleString('id-ID') : ''}
                        onChange={e => {
                          const val = e.target.value.replace(/[^0-9]/g, '');
                          setForm(f => ({ ...f, overhead_cost: val }));
                        }}
                      />
                    </div>
                    <div>
                      <label
                        style={{
                          display: 'block',
                          fontSize: '0.75rem',
                          fontWeight: 600,
                          marginBottom: 5,
                          color: 'var(--text-2)',
                        }}
                      >
                        Tenaga Kerja
                      </label>
                      <input
                        className="input"
                        type="text"
                        placeholder="0"
                        onFocus={e => e.target.select()}
                        value={form.labor_cost ? Number(form.labor_cost).toLocaleString('id-ID') : ''}
                        onChange={e => {
                          const val = e.target.value.replace(/[^0-9]/g, '');
                          setForm(f => ({ ...f, labor_cost: val }));
                        }}
                      />
                    </div>
                  </div>
                </div>

                {typeof products !== 'undefined' &&
                  (() => {
                    const ingredientCost = form.ingredients.reduce((acc, ing) => {
                      const p = products.find(x => x.id === ing.productId);
                      return acc + (p?.cost_price || 0) * ing.quantity;
                    }, 0);
                    const packagingCost = parseFloat(form.packaging_cost) || 0;
                    const overheadCost = parseFloat(form.overhead_cost) || 0;
                    const laborCost = parseFloat(form.labor_cost) || 0;
                    const totalHpp = ingredientCost + packagingCost + overheadCost + laborCost;

                    const suggestedPriceMin = totalHpp * 1.3;
                    const suggestedPriceMax = totalHpp * 1.45;
                    const currentPrice = parseFloat(form.sell_price) || 0;

                    const showSummary = totalHpp > 0;
                    if (!showSummary) return null;

                    let marginPct = 0;
                    let profit = 0;
                    if (currentPrice > 0 && currentPrice >= totalHpp) {
                      profit = currentPrice - totalHpp;
                      marginPct = (profit / totalHpp) * 100;
                    }

                    const isHppIncreased = totalHpp > prevHppRef.current && prevHppRef.current > 0;
                    prevHppRef.current = totalHpp;

                    return (
                      <div
                        style={{
                          marginTop: 12,
                          padding: 14,
                          background: 'var(--bg-elevated)',
                          borderRadius: 8,
                          border: '1px solid var(--border-md)',
                        }}
                      >
                        <style>{`
                          @keyframes bumpRed {
                            0% { transform: scale(1); color: var(--text-1); }
                            30% { transform: scale(1.1); color: #ef4444; }
                            100% { transform: scale(1); color: var(--text-1); }
                          }
                          @keyframes bumpNormal {
                            0% { transform: scale(1); }
                            30% { transform: scale(1.05); }
                            100% { transform: scale(1); }
                          }
                          .animate-hpp {
                            display: inline-block;
                          }
                        `}</style>
                        <div
                          style={{
                            marginBottom: 12,
                            paddingBottom: 10,
                            borderBottom: '1px dashed var(--border)',
                          }}
                        >
                          <div
                            style={{
                              display: 'flex',
                              justifyContent: 'space-between',
                              fontSize: '0.75rem',
                              marginBottom: 4,
                              color: 'var(--text-2)',
                            }}
                          >
                            <span>Bahan Baku</span>
                            <span>{formatRp(ingredientCost)}</span>
                          </div>
                          <div
                            style={{
                              display: 'flex',
                              justifyContent: 'space-between',
                              fontSize: '0.75rem',
                              marginBottom: 4,
                              color: 'var(--text-2)',
                            }}
                          >
                            <span>Packaging</span>
                            <span>{formatRp(packagingCost)}</span>
                          </div>
                          <div
                            style={{
                              display: 'flex',
                              justifyContent: 'space-between',
                              fontSize: '0.75rem',
                              marginBottom: 4,
                              color: 'var(--text-2)',
                            }}
                          >
                            <span>Overhead</span>
                            <span>{formatRp(overheadCost)}</span>
                          </div>
                          <div
                            style={{
                              display: 'flex',
                              justifyContent: 'space-between',
                              fontSize: '0.75rem',
                              color: 'var(--text-2)',
                            }}
                          >
                            <span>Tenaga Kerja</span>
                            <span>{formatRp(laborCost)}</span>
                          </div>
                        </div>
                        <div
                          style={{
                            display: 'flex',
                            justifyContent: 'space-between',
                            marginBottom: 10,
                            fontSize: '0.85rem',
                          }}
                        >
                          <span style={{ color: 'var(--text-1)', fontWeight: 600 }}>Total HPP</span>
                          <span
                            key={totalHpp}
                            className="animate-hpp"
                            style={{ 
                              fontWeight: 700, 
                              color: 'var(--text-1)',
                              animation: isHppIncreased ? 'bumpRed 0.8s ease' : 'bumpNormal 0.4s ease'
                            }}
                          >
                            {formatRp(totalHpp)}
                          </span>
                        </div>
                        <div
                          style={{
                            display: 'flex',
                            justifyContent: 'space-between',
                            marginBottom: 10,
                            fontSize: '0.8rem',
                          }}
                        >
                          <span style={{ color: 'var(--text-3)' }}>Saran Harga (30% - 45%)</span>
                          <span style={{ color: 'var(--brand)', fontWeight: 600 }}>
                            {formatRp(suggestedPriceMin)} - {formatRp(suggestedPriceMax)}
                          </span>
                        </div>
                        {currentPrice > 0 && (
                          <div
                            style={{
                              display: 'flex',
                              justifyContent: 'space-between',
                              marginBottom: 10,
                              fontSize: '0.8rem',
                            }}
                          >
                            <span style={{ color: 'var(--text-3)' }}>Profit saat ini (Margin)</span>
                            <span
                              style={{
                                color: marginPct >= 30 ? 'var(--brand)' : '#ef4444',
                                fontWeight: 600,
                              }}
                            >
                              {formatRp(profit)} ({marginPct.toFixed(1)}%)
                            </span>
                          </div>
                        )}
                        <div
                          style={{
                            display: 'flex',
                            gap: 8,
                            alignItems: 'center',
                            justifyContent: 'flex-end',
                            marginTop: 12,
                          }}
                        >
                          <input
                            type="text"
                            placeholder="Kustom Harga..."
                            style={{
                              width: 120,
                              textAlign: 'right',
                              fontSize: '0.8rem',
                              padding: '6px 10px',
                              borderRadius: 6,
                              border: '1px solid var(--border-md)',
                              background: 'var(--bg-base)',
                              color: 'var(--text-1)'
                            }}
                            onFocus={e => e.target.select()}
                            value={form.sell_price ? Number(form.sell_price).toLocaleString('id-ID') : ''}
                            onChange={e => {
                              const val = e.target.value.replace(/[^0-9]/g, '');
                              setForm(f => ({ ...f, sell_price: val }));
                            }}
                          />
                          <button
                            type="button"
                            className="btn btn-ghost btn-sm"
                            onClick={() =>
                              setForm(f => ({
                                ...f,
                                sell_price: String(Math.round(suggestedPriceMin)),
                              }))
                            }
                          >
                            Set Margin 30%
                          </button>
                          <button
                            type="button"
                            className="btn btn-primary btn-sm"
                            onClick={() =>
                              setForm(f => ({
                                ...f,
                                sell_price: String(Math.round(suggestedPriceMax)),
                              }))
                            }
                          >
                            Set Margin 45%
                          </button>
                        </div>
                      </div>
                    );
                  })()}
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

      {/* ── Delete Confirm ─────────────────────────────────────────────────── */}
      {deleteConfirm.open && deleteConfirm.item && (
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
            <div style={{ fontWeight: 700, fontSize: '1rem', marginBottom: 8 }}>Hapus Menu?</div>
            <div
              style={{
                fontSize: '0.85rem',
                color: 'var(--text-2)',
                marginBottom: 20,
                lineHeight: 1.6,
              }}
            >
              Menu <strong>&ldquo;{deleteConfirm.item.name}&rdquo;</strong> akan dihapus
              (soft-delete).
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
