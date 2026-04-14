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
  sell_price: string;
  use_global_tax: boolean;
  tax_percentage: string;
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
  use_global_tax: true,
  tax_percentage: '',
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
  const [sidebar, setSidebar] = useState<{
    open: boolean;
    mode: 'create' | 'edit';
    item?: MenuItem;
  }>({
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
      const menuData = (menuRes.data as unknown as { data: MenuItem[] }).data ?? menuRes.data ?? [];
      const catData = (catRes.data as unknown as { data: Category[] }).data ?? catRes.data ?? [];
      const prodData = (prodRes.data as unknown as { data: Product[] }).data ?? [];

      setItems(menuData);
      setCategories(catData);
      setProducts(prodData);
    } catch {
      showToast('Gagal memuat data', 'error');
    } finally {
      setLoading(false);
    }
  }, [storeId, showToast]);

  useEffect(() => {
    fetchAll();
  }, [fetchAll]);

  // ── Sidebar Logic ──────────────────────────────────────────────────────────

  const openCreate = () => {
    setForm(emptyForm());
    setFormError('');
    setSidebar({ open: true, mode: 'create' });
  };

  const openEdit = (item: MenuItem) => {
    setForm({
      name: item.name,
      description: item.description,
      category_id: item.category_id ?? '',
      sell_price: String(item.sell_price),
      use_global_tax: item.use_global_tax ?? true,
      tax_percentage:
        item.tax_percentage !== null && item.tax_percentage !== undefined
          ? String(item.tax_percentage)
          : '',
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
    setSidebar({ open: true, mode: 'edit', item });
  };

  const closeSidebar = () => {
    setSidebar({ open: false, mode: 'create' });
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
      use_global_tax: form.use_global_tax,
      tax_percentage: form.use_global_tax ? null : parseFloat(form.tax_percentage) || 0,
      packaging_cost: parseFloat(form.packaging_cost) || 0,
      overhead_cost: parseFloat(form.overhead_cost) || 0,
      labor_cost: parseFloat(form.labor_cost) || 0,
      ingredients: form.ingredients.map(i => ({ product_id: i.productId, quantity: i.quantity })),
    };
    try {
      if (sidebar.mode === 'create') {
        await menuItemsApi.create(storeId, payload);
        showToast('Menu berhasil ditambahkan ✓', 'success');
      } else if (sidebar.item) {
        await menuItemsApi.update(storeId, sidebar.item.id, payload);
        showToast('Menu berhasil diperbarui ✓', 'success');
      } else {
        setFormError('Menu tidak ditemukan');
        return;
      }
      closeSidebar();
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

      {/* ── Create / Edit Sidebar ─────────────────────────────────────────────── */}
      {sidebar.open && (
        <div
          style={{
            position: 'fixed',
            inset: 0,
            zIndex: 1000,
            display: 'flex',
            justifyContent: 'flex-end',
          }}
          onClick={closeSidebar}
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
          {/* Sidebar drawer content */}
          <div
            className="card"
            style={{
              position: 'relative',
              width: '100%',
              maxWidth: 560,
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
                {sidebar.mode === 'create' ? 'Tambah Menu Item' : `Edit: ${sidebar.item?.name}`}
              </div>
              <button
                onClick={closeSidebar}
                className="btn btn-ghost btn-sm"
                style={{ padding: 6 }}
              >
                <X size={16} />
              </button>
            </div>

            {/* Scrollable Form Body */}
            <div style={{ flex: 1, overflowY: 'auto', padding: '24px' }}>
              <form
                id="menu-form"
                onSubmit={handleSubmit}
                style={{ display: 'flex', flexDirection: 'column', gap: 20 }}
              >
                {/* Name */}
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

                <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
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
                        display: 'flex',
                        alignItems: 'center',
                        gap: 8,
                        fontSize: '0.8rem',
                        fontWeight: 600,
                        marginBottom: form.use_global_tax ? 15 : 5,
                        color: 'var(--text-2)',
                        cursor: 'pointer',
                      }}
                    >
                      <input
                        type="checkbox"
                        checked={form.use_global_tax}
                        onChange={e => setForm(f => ({ ...f, use_global_tax: e.target.checked }))}
                      />
                      Gunakan PPN default toko
                    </label>
                    {!form.use_global_tax && (
                      <>
                        <label
                          style={{
                            display: 'block',
                            fontSize: '0.8rem',
                            fontWeight: 600,
                            marginBottom: 5,
                            color: 'var(--text-2)',
                          }}
                        >
                          Custom PPN (%)
                        </label>
                        <input
                          className="input"
                          type="number"
                          min={0}
                          max={100}
                          placeholder="10"
                          value={form.tax_percentage}
                          onChange={e => setForm(f => ({ ...f, tax_percentage: e.target.value }))}
                        />
                      </>
                    )}
                  </div>
                </div>

                {/* Category */}
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
                <div>
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
                      style={{
                        display: 'flex',
                        flexDirection: 'column',
                        gap: 10,
                        marginBottom: 14,
                      }}
                    >
                      {form.ingredients.map(ing => (
                        <div
                          key={ing.productId}
                          style={{
                            display: 'flex',
                            alignItems: 'center',
                            gap: 12,
                            padding: '10px 14px',
                            background: 'var(--bg-elevated)',
                            borderRadius: 10,
                            border: '1px solid var(--border)',
                          }}
                        >
                          <Package size={14} style={{ color: '#fb923c', flexShrink: 0 }} />
                          <div style={{ flex: 1, display: 'flex', flexDirection: 'column' }}>
                            <span
                              style={{
                                fontSize: '0.85rem',
                                color: 'var(--text-1)',
                                fontWeight: 600,
                              }}
                            >
                              {ing.productName}
                            </span>
                            {(() => {
                              const p = products.find(x => x.id === ing.productId);
                              const cost = (p?.cost_price || 0) * ing.quantity;
                              return (
                                <span style={{ fontSize: '0.72rem', color: 'var(--text-3)' }}>
                                  {formatRp(p?.cost_price || 0)}/{ing.unit} = {formatRp(cost)}
                                </span>
                              );
                            })()}
                          </div>
                          <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
                            <input
                              type="number"
                              min={0.01}
                              step={0.01}
                              value={ing.quantity}
                              onChange={e =>
                                updateIngredientQty(ing.productId, parseFloat(e.target.value) || 0)
                              }
                              style={{
                                width: 64,
                                background: 'var(--bg-base)',
                                border: '1px solid var(--border-md)',
                                borderRadius: 6,
                                padding: '4px 6px',
                                color: 'var(--text-1)',
                                fontSize: '0.8rem',
                                textAlign: 'right',
                              }}
                            />
                            <span
                              style={{ fontSize: '0.75rem', color: 'var(--text-3)', minWidth: 32 }}
                            >
                              {ing.unit}
                            </span>
                          </div>
                          <button
                            type="button"
                            onClick={() => removeIngredient(ing.productId)}
                            className="btn btn-ghost btn-sm"
                            style={{ color: '#ef4444', padding: 4 }}
                          >
                            <X size={15} />
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
                          borderRadius: 10,
                          marginTop: 6,
                          boxShadow: '0 10px 30px rgba(0,0,0,0.25)',
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
                              gap: 12,
                              padding: '10px 16px',
                              background: 'none',
                              border: 'none',
                              cursor: 'pointer',
                              textAlign: 'left',
                            }}
                            className="category-row hover:bg-muted"
                          >
                            <Package size={14} style={{ color: 'var(--text-3)', flexShrink: 0 }} />
                            <span style={{ flex: 1, fontSize: '0.85rem', color: 'var(--text-1)' }}>
                              {p.name}
                            </span>
                            <span style={{ fontSize: '0.75rem', color: 'var(--text-3)' }}>
                              {p.unit}
                            </span>
                          </button>
                        ))}
                      </div>
                    )}
                  </div>
                </div>

                {/* Biaya Tambahan */}
                <div>
                  <div
                    style={{
                      fontSize: '0.85rem',
                      fontWeight: 700,
                      color: 'var(--text-1)',
                      marginBottom: 12,
                    }}
                  >
                    Biaya Tambahan
                  </div>
                  <div className="grid grid-cols-3 gap-3">
                    <div>
                      <label
                        style={{
                          display: 'block',
                          fontSize: '0.72rem',
                          fontWeight: 600,
                          marginBottom: 4,
                          color: 'var(--text-2)',
                        }}
                      >
                        Packaging (Rp)
                      </label>
                      <input
                        className="input"
                        type="text"
                        placeholder="0"
                        onFocus={e => e.target.select()}
                        value={
                          form.packaging_cost
                            ? Number(form.packaging_cost).toLocaleString('id-ID')
                            : ''
                        }
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
                          fontSize: '0.72rem',
                          fontWeight: 600,
                          marginBottom: 4,
                          color: 'var(--text-2)',
                        }}
                      >
                        Overhead (Rp)
                      </label>
                      <input
                        className="input"
                        type="text"
                        placeholder="0"
                        onFocus={e => e.target.select()}
                        value={
                          form.overhead_cost
                            ? Number(form.overhead_cost).toLocaleString('id-ID')
                            : ''
                        }
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
                          fontSize: '0.72rem',
                          fontWeight: 600,
                          marginBottom: 4,
                          color: 'var(--text-2)',
                        }}
                      >
                        Tenaga Kerja (Rp)
                      </label>
                      <input
                        className="input"
                        type="text"
                        placeholder="0"
                        onFocus={e => e.target.select()}
                        value={
                          form.labor_cost ? Number(form.labor_cost).toLocaleString('id-ID') : ''
                        }
                        onChange={e => {
                          const val = e.target.value.replace(/[^0-9]/g, '');
                          setForm(f => ({ ...f, labor_cost: val }));
                        }}
                      />
                    </div>
                  </div>
                </div>

                {/* HPP Summary Card */}
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

                    prevHppRef.current = totalHpp;

                    return (
                      <div
                        style={{
                          marginTop: 10,
                          padding: 18,
                          background: 'rgba(8, 132, 246, 0.04)',
                          borderRadius: 12,
                          border: '1px solid rgba(8, 132, 246, 0.12)',
                        }}
                      >
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
                              fontSize: '0.78rem',
                              marginBottom: 6,
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
                              fontSize: '0.78rem',
                              marginBottom: 6,
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
                              fontSize: '0.78rem',
                              marginBottom: 6,
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
                              fontSize: '0.78rem',
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
                            marginBottom: 12,
                            fontSize: '0.9rem',
                          }}
                        >
                          <span style={{ color: 'var(--text-1)', fontWeight: 700 }}>Total HPP</span>
                          <span
                            key={totalHpp}
                            style={{
                              fontWeight: 800,
                              color: 'var(--text-1)',
                              fontSize: '1rem',
                            }}
                          >
                            {formatRp(totalHpp)}
                          </span>
                        </div>
                        <div
                          style={{
                            display: 'flex',
                            justifyContent: 'space-between',
                            marginBottom: 12,
                            fontSize: '0.82rem',
                          }}
                        >
                          <span style={{ color: 'var(--text-3)' }}>Saran Harga (30-45%)</span>
                          <span style={{ color: 'var(--brand)', fontWeight: 700 }}>
                            {formatRp(suggestedPriceMin)} - {formatRp(suggestedPriceMax)}
                          </span>
                        </div>
                        {currentPrice > 0 && (
                          <div
                            style={{
                              display: 'flex',
                              justifyContent: 'space-between',
                              marginBottom: 14,
                              fontSize: '0.82rem',
                            }}
                          >
                            <span style={{ color: 'var(--text-3)' }}>Profit (Margin)</span>
                            <span
                              style={{
                                color: marginPct >= 30 ? 'var(--brand)' : '#ef4444',
                                fontWeight: 700,
                              }}
                            >
                              {formatRp(profit)} ({marginPct.toFixed(1)}%)
                            </span>
                          </div>
                        )}
                        <div
                          style={{
                            display: 'flex',
                            flexDirection: 'column',
                            gap: 8,
                          }}
                        >
                          <div className="flex gap-2">
                            <div
                              style={{
                                flex: 1,
                                display: 'flex',
                                alignItems: 'center',
                                background: 'var(--bg-base)',
                                border: '1px solid var(--border-md)',
                                borderRadius: 8,
                                padding: '0 12px',
                              }}
                            >
                              <span
                                style={{
                                  fontSize: '0.75rem',
                                  color: 'var(--text-3)',
                                  fontWeight: 600,
                                  whiteSpace: 'nowrap',
                                }}
                              >
                                Harga Jual (Rp)
                              </span>
                              <input
                                type="text"
                                placeholder="0"
                                style={{
                                  flex: 1,
                                  textAlign: 'right',
                                  fontSize: '0.8rem',
                                  padding: '8px 0',
                                  background: 'transparent',
                                  border: 'none',
                                  outline: 'none',
                                  color: 'var(--text-1)',
                                }}
                                onFocus={e => e.target.select()}
                                value={
                                  form.sell_price
                                    ? Number(form.sell_price).toLocaleString('id-ID')
                                    : ''
                                }
                                onChange={e => {
                                  const val = e.target.value.replace(/[^0-9]/g, '');
                                  setForm(f => ({ ...f, sell_price: val }));
                                }}
                              />
                            </div>
                          </div>
                          <div className="flex gap-2">
                            <button
                              type="button"
                              className="btn btn-secondary flex-1 text-xs"
                              onClick={() =>
                                setForm(f => ({
                                  ...f,
                                  sell_price: String(Math.round(suggestedPriceMin)),
                                }))
                              }
                            >
                              Set 30%
                            </button>
                            <button
                              type="button"
                              className="btn btn-primary flex-1 text-xs"
                              onClick={() =>
                                setForm(f => ({
                                  ...f,
                                  sell_price: String(Math.round(suggestedPriceMax)),
                                }))
                              }
                            >
                              Set 45%
                            </button>
                          </div>
                        </div>
                      </div>
                    );
                  })()}

                {formError && (
                  <div
                    style={{
                      background: 'rgba(239,68,68,0.12)',
                      border: '1px solid rgba(239,68,68,0.3)',
                      borderRadius: 10,
                      padding: '12px 16px',
                      fontSize: '0.85rem',
                      color: '#ef4444',
                    }}
                  >
                    {formError}
                  </div>
                )}
              </form>
            </div>

            {/* Footer */}
            <div
              style={{
                padding: '20px 24px',
                borderTop: '1px solid var(--border)',
                display: 'flex',
                gap: 12,
                justifyContent: 'flex-end',
                background: 'var(--bg-card)',
              }}
            >
              <button type="button" className="btn btn-ghost" onClick={closeSidebar}>
                Batal
              </button>
              <button
                type="submit"
                form="menu-form"
                className="btn btn-primary"
                disabled={submitting}
                style={{ minWidth: 120 }}
              >
                {submitting ? (
                  <>
                    <Loader2 size={16} className="loading-spin mr-2" /> Menyimpan...
                  </>
                ) : (
                  <>
                    <Check size={16} className="mr-2" /> Simpan Menu
                  </>
                )}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Delete Confirmation Modal */}
      {deleteConfirm.open && (
        <div
          style={{
            position: 'fixed',
            inset: 0,
            zIndex: 1100,
            background: 'rgba(0,0,0,0.6)',
            backdropFilter: 'blur(4px)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            padding: 20,
          }}
        >
          <div
            className="card"
            style={{ width: '100%', maxWidth: 400, padding: 32, textAlign: 'center' }}
          >
            <div
              style={{
                width: 64,
                height: 64,
                borderRadius: '50%',
                background: 'rgba(239,68,68,0.1)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                margin: '0 auto 20px',
              }}
            >
              <Trash2 size={32} style={{ color: '#ef4444' }} />
            </div>
            <h3
              style={{
                fontSize: '1.2rem',
                fontWeight: 700,
                marginBottom: 8,
                color: 'var(--text-1)',
              }}
            >
              Hapus Menu?
            </h3>
            <p style={{ fontSize: '0.9rem', color: 'var(--text-3)', marginBottom: 28 }}>
              Anda akan menghapus &ldquo;{deleteConfirm.item?.name}&rdquo;. Tindakan ini tidak dapat
              dibatalkan.
            </p>
            <div style={{ display: 'flex', gap: 12 }}>
              <button
                className="btn btn-ghost flex-1"
                onClick={() => setDeleteConfirm({ open: false })}
              >
                Batal
              </button>
              <button
                className="btn btn-primary flex-1"
                style={{ background: '#ef4444' }}
                onClick={handleDelete}
              >
                Ya, Hapus
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
