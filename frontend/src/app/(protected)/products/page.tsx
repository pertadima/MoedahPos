'use client';

import { useEffect, useState, useCallback } from 'react';
import {
  Search,
  Plus,
  Package,
  Pencil,
  Trash2,
  Loader2,
  X,
  ChevronLeft,
  ChevronRight,
  Check,
} from 'lucide-react';
import { useAuth } from '@/lib/auth/AuthContext';
import { usePermission } from '@/hooks/usePermission';
import { productsApi } from '@/lib/api/products';
import { stockApi } from '@/lib/api/store-apis';
import { formatRp } from '@/lib/utils';
import type { Product, Category, PaginatedData } from '@/types';
import { ApiError } from '@/lib/api/client';

export default function ProductsPage() {
  const { selectedStore } = useAuth();
  const { can } = usePermission();
  const [products, setProducts] = useState<Product[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState('');
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [sidebar, setSidebar] = useState<{
    open: boolean;
    mode: 'create' | 'edit';
    item?: Product;
  }>({
    open: false,
    mode: 'create',
  });
  const [formData, setFormData] = useState({
    name: '',
    sku: '',
    sell_price: '',
    cost_price: '',
    unit: 'pcs',
    use_global_tax: true,
    tax_percentage: '',
    category_id: '',
    initial_qty: '0',
  });
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');
  const [minStock, setMinStock] = useState('0');

  const storeId = selectedStore?.store_id;
  const isRestaurant = selectedStore?.store_type === 'restaurant';
  const productLabel = isRestaurant ? 'Bahan Baku' : 'Produk';
  const productLabelLower = productLabel.toLowerCase();

  const load = useCallback(() => {
    if (!storeId) return;
    setLoading(true);
    productsApi
      .list(storeId, { page, per_page: 15, search: search || undefined })
      .then(res => {
        const d = res.data as PaginatedData<Product>;
        setProducts(d.data ?? []);
        setTotalPages(d.meta?.total_pages ?? 1);
      })
      .catch(console.error)
      .finally(() => setLoading(false));
  }, [page, search, storeId]);

  useEffect(() => {
    if (storeId) productsApi.listCategories(storeId).then(r => setCategories(r.data as Category[]));
  }, [storeId]);
  useEffect(() => {
    load();
  }, [load]);

  const openCreate = () => {
    setFormData({
      name: '',
      sku: '',
      sell_price: '',
      cost_price: '',
      unit: 'pcs',
      use_global_tax: true,
      tax_percentage: '',
      category_id: '',
      initial_qty: '0',
    });
    setMinStock('0');
    setError('');
    setSidebar({ open: true, mode: 'create' });
  };
  const openEdit = (p: Product) => {
    setFormData({
      name: p.name,
      sku: p.sku,
      sell_price: String(p.sell_price),
      cost_price: String(p.cost_price),
      unit: p.unit,
      use_global_tax: p.use_global_tax ?? true,
      tax_percentage:
        p.tax_percentage !== null && p.tax_percentage !== undefined ? String(p.tax_percentage) : '',
      category_id: p.category_id ?? '',
      initial_qty: '0',
    });
    setError('');
    setSidebar({ open: true, mode: 'edit', item: p });
    // Load current min stock for this product
    if (storeId) {
      stockApi
        .levels(storeId)
        .then(res => {
          const levels = res.data as import('@/types').StockLevel[];
          const level = levels.find(l => l.product_id === p.id);
          setMinStock(level ? String(level.min_quantity) : '0');
        })
        .catch(() => setMinStock('0'));
    }
  };

  const handleSave = async () => {
    if (!storeId) return;
    setSaving(true);
    setError('');
    try {
      const payload = {
        name: formData.name,
        sku: formData.sku,
        sell_price: +formData.sell_price,
        cost_price: +formData.cost_price,
        unit: formData.unit,
        use_global_tax: formData.use_global_tax,
        tax_percentage: formData.use_global_tax ? null : +formData.tax_percentage || 0,
        category_id: formData.category_id || undefined,
        initial_qty: +formData.initial_qty,
        is_active: true,
      };
      let savedProduct: Product;
      if (sidebar.mode === 'edit' && sidebar.item) {
        const res = await productsApi.update(storeId, sidebar.item.id, payload);
        savedProduct = res.data as Product;
      } else {
        const res = await productsApi.create(storeId, payload);
        savedProduct = res.data as Product;
      }
      // Save min stock via stock API (upsert)
      const minVal = parseFloat(minStock);
      if (!isNaN(minVal) && minVal >= 0) {
        await stockApi
          .setMin(storeId, { product_id: savedProduct.id, min_quantity: minVal })
          .catch(() => {}); // non-blocking; stock level row may not exist yet for brand new products
      }
      setSidebar({ open: false, mode: 'create' });
      load();
    } catch (e) {
      setError(e instanceof ApiError ? e.message : 'Gagal menyimpan');
    } finally {
      setSaving(false);
    }
  };

  const closeSidebar = () => {
    setSidebar({ open: false, mode: 'create' });
  };

  const handleDelete = async (id: string) => {
    if (!storeId || !confirm('Hapus produk ini?')) return;
    await productsApi.delete(storeId, id).catch(console.error);
    load();
  };

  const f =
    (k: keyof typeof formData) => (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) =>
      setFormData(d => ({ ...d, [k]: e.target.value }));

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
            <Package size={22} style={{ color: 'var(--accent-em)' }} />
            {productLabel}
          </h1>
          <p className="page-subtitle">
            Kelola katalog {productLabelLower} {selectedStore?.store_name}
          </p>
        </div>
        {can('products.create') && (
          <button className="btn btn-primary" onClick={openCreate}>
            <Plus size={15} /> Tambah {productLabel}
          </button>
        )}
      </div>

      {/* Search */}
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
          placeholder={`Cari ${productLabelLower}...`}
          value={search}
          onChange={e => {
            setSearch(e.target.value);
            setPage(1);
          }}
        />
      </div>

      {/* Table */}
      <div className="card reveal-animate" style={{ overflow: 'hidden', animationDelay: '0.2s' }}>
        {loading ? (
          <div style={{ display: 'flex', justifyContent: 'center', padding: 40 }}>
            <Loader2 size={24} className="loading-spin" style={{ color: 'var(--accent-em)' }} />
          </div>
        ) : products.length === 0 ? (
          <div className="empty-state">
            <Package size={32} />
            <p>Belum ada {productLabelLower}</p>
          </div>
        ) : (
          <table className="tbl">
            <thead>
              <tr>
                <th>{productLabel}</th>
                <th>SKU</th>
                <th>Stok</th>
                <th>Harga Jual</th>
                <th>Harga Beli</th>
                <th>PPN</th>
                <th>Status</th>
                <th />
              </tr>
            </thead>
            <tbody>
              {products.map((p, i) => (
                <tr
                  key={p.id}
                  className="reveal-animate"
                  style={{ animationDelay: `${0.25 + i * 0.02}s` }}
                >
                  <td>
                    <div style={{ fontWeight: 600 }}>{p.name}</div>
                    {p.category_name && (
                      <div style={{ fontSize: '0.75rem', color: 'var(--text-3)' }}>
                        {p.category_name}
                      </div>
                    )}
                  </td>
                  <td
                    style={{ color: 'var(--text-2)', fontFamily: 'monospace', fontSize: '0.82rem' }}
                  >
                    {p.sku}
                  </td>
                  <td>
                    {p.stock_qty !== undefined ? (
                      <span
                        className={`badge ${p.stock_qty <= 0 ? 'badge-red' : p.stock_qty <= 5 ? 'badge-amber' : 'badge-green'}`}
                      >
                        {p.stock_qty} {p.unit}
                      </span>
                    ) : (
                      '–'
                    )}
                  </td>
                  <td style={{ fontWeight: 600, color: 'var(--accent-em)' }}>
                    {formatRp(p.sell_price)}
                  </td>
                  <td style={{ color: 'var(--text-2)' }}>{formatRp(p.cost_price)}</td>
                  <td style={{ color: 'var(--text-2)' }}>{p.tax_rate}%</td>
                  <td>
                    <span className={`badge ${p.is_active ? 'badge-green' : 'badge-gray'}`}>
                      {p.is_active ? 'Aktif' : 'Nonaktif'}
                    </span>
                  </td>
                  <td>
                    <div style={{ display: 'flex', gap: 4 }}>
                      {can('products.update') && (
                        <button className="btn btn-ghost btn-sm" onClick={() => openEdit(p)}>
                          <Pencil size={13} />
                        </button>
                      )}
                      {can('products.delete') && (
                        <button
                          className="btn btn-ghost btn-sm"
                          style={{ color: 'var(--accent-rd)' }}
                          onClick={() => handleDelete(p.id)}
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
        {/* Pagination */}
        {totalPages > 1 && (
          <div
            style={{
              display: 'flex',
              justifyContent: 'center',
              alignItems: 'center',
              gap: 12,
              padding: 14,
              borderTop: '1px solid var(--border)',
            }}
          >
            <button
              className="btn btn-ghost btn-sm"
              disabled={page <= 1}
              onClick={() => setPage(p => p - 1)}
            >
              <ChevronLeft size={15} />
            </button>
            <span style={{ fontSize: '0.85rem', color: 'var(--text-2)' }}>
              Hal {page} / {totalPages}
            </span>
            <button
              className="btn btn-ghost btn-sm"
              disabled={page >= totalPages}
              onClick={() => setPage(p => p + 1)}
            >
              <ChevronRight size={15} />
            </button>
          </div>
        )}
      </div>

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
              maxWidth: 480,
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
              <h2 style={{ fontWeight: 700, fontSize: '1.1rem', margin: 0 }}>
                {sidebar.mode === 'edit' ? `Edit ${productLabel}` : `Tambah ${productLabel}`}
              </h2>
              <button
                onClick={closeSidebar}
                className="btn btn-ghost btn-sm"
                style={{ padding: 6 }}
              >
                <X size={18} />
              </button>
            </div>

            {/* Scrollable Form Body */}
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

              <form
                id="product-form"
                onSubmit={e => {
                  e.preventDefault();
                  handleSave();
                }}
                style={{ display: 'flex', flexDirection: 'column', gap: 20 }}
              >
                {/* Basic Info */}
                <div>
                  <label className="input-label">Nama {productLabel}</label>
                  <input
                    className="input"
                    autoFocus
                    placeholder={`cth. ${isRestaurant ? 'Ayam Potong' : 'Ayam Goreng'}`}
                    value={formData.name}
                    onChange={f('name')}
                  />
                </div>

                <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <div className="input-group">
                    <label className="input-label">SKU</label>
                    <input
                      className="input"
                      placeholder="Otomatis jika kosong"
                      value={formData.sku}
                      onChange={f('sku')}
                    />
                  </div>
                  <div className="input-group">
                    <label className="input-label">Satuan</label>
                    <input
                      className="input"
                      placeholder="pcs, kg, gr, dll"
                      value={formData.unit}
                      onChange={f('unit')}
                    />
                  </div>
                </div>

                <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <div className="input-group">
                    <label className="input-label">Harga Jual (Rp)</label>
                    <input
                      type="text"
                      className="input"
                      placeholder="0"
                      onFocus={e => e.target.select()}
                      value={
                        formData.sell_price
                          ? Number(formData.sell_price).toLocaleString('id-ID')
                          : ''
                      }
                      onChange={e => {
                        const val = e.target.value.replace(/[^0-9]/g, '');
                        setFormData(d => ({ ...d, sell_price: val }));
                      }}
                    />
                  </div>
                  <div className="input-group">
                    <label className="input-label">Harga Beli (Rp)</label>
                    <input
                      type="text"
                      className="input"
                      placeholder="0"
                      onFocus={e => e.target.select()}
                      value={
                        formData.cost_price
                          ? Number(formData.cost_price).toLocaleString('id-ID')
                          : ''
                      }
                      onChange={e => {
                        const val = e.target.value.replace(/[^0-9]/g, '');
                        setFormData(d => ({ ...d, cost_price: val }));
                      }}
                    />
                  </div>
                </div>

                <div className="input-group">
                  <label
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: 8,
                      fontSize: '0.8rem',
                      fontWeight: 600,
                      marginBottom: formData.use_global_tax ? 0 : 8,
                      color: 'var(--text-2)',
                      cursor: 'pointer',
                    }}
                  >
                    <input
                      type="checkbox"
                      checked={formData.use_global_tax}
                      onChange={e => setFormData(d => ({ ...d, use_global_tax: e.target.checked }))}
                    />
                    Gunakan PPN default toko
                  </label>
                  {!formData.use_global_tax && (
                    <div style={{ marginTop: 8 }}>
                      <label className="input-label">Custom PPN (%)</label>
                      <input
                        className="input"
                        type="number"
                        min={0}
                        max={100}
                        placeholder="10"
                        value={formData.tax_percentage}
                        onChange={f('tax_percentage')}
                      />
                    </div>
                  )}
                </div>

                <div className="input-group">
                  <label className="input-label">Kategori</label>
                  <select
                    className="input"
                    value={formData.category_id}
                    onChange={f('category_id')}
                  >
                    <option value="">Tanpa Kategori</option>
                    {categories.map(c => (
                      <option key={c.id} value={c.id}>
                        {c.name}
                      </option>
                    ))}
                  </select>
                </div>

                {!sidebar.item && (
                  <div className="input-group">
                    <label className="input-label">Stok Awal</label>
                    <input
                      type="number"
                      className="input"
                      value={formData.initial_qty}
                      onChange={f('initial_qty')}
                    />
                  </div>
                )}

                <div
                  className="input-group"
                  style={{
                    padding: '16px',
                    background: 'var(--bg-elevated)',
                    borderRadius: 12,
                    border: '1px solid var(--border)',
                  }}
                >
                  <label
                    className="input-label"
                    style={{ display: 'flex', alignItems: 'center', gap: 6, marginBottom: 8 }}
                  >
                    Ambeng Stok (Min Stok)
                    <span
                      style={{
                        fontSize: '0.68rem',
                        color: 'var(--text-3)',
                        fontWeight: 400,
                        background: 'var(--border)',
                        borderRadius: 4,
                        padding: '1px 6px',
                      }}
                    >
                      Alert
                    </span>
                  </label>
                  <input
                    type="number"
                    className="input"
                    min={0}
                    step={1}
                    value={minStock}
                    onChange={e => setMinStock(e.target.value)}
                    placeholder="0"
                  />
                  <p
                    style={{
                      fontSize: '0.73rem',
                      color: 'var(--text-3)',
                      marginTop: 8,
                      lineHeight: 1.4,
                    }}
                  >
                    Sistem akan menandai stok sebagai &ldquo;Menipis&rdquo; ketika stok saat ini ≤
                    nilai ini.
                  </p>
                </div>
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
                form="product-form"
                className="btn btn-primary"
                disabled={saving}
                style={{ minWidth: 130 }}
              >
                {saving ? (
                  <>
                    <Loader2 size={16} className="loading-spin mr-2" /> Menyimpan...
                  </>
                ) : (
                  <>
                    <Check size={16} className="mr-2" /> Simpan {productLabel}
                  </>
                )}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
