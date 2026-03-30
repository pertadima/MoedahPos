'use client';

import { useEffect, useState } from 'react';
import { ClipboardList, Plus, Loader2, X, ChevronRight } from 'lucide-react';
import { useAuth } from '@/lib/auth/AuthContext';
import { purchaseOrdersApi } from '@/lib/api/store-apis';
import { productsApi } from '@/lib/api/products';
import { suppliersApi } from '@/lib/api/store-apis';
import { formatRp, formatDate } from '@/lib/utils';
import type { PurchaseOrder, Product, Supplier } from '@/types';
import { ApiError } from '@/lib/api/client';

const STATUS_BADGE: Record<string, string> = { draft: 'badge-gray', ordered: 'badge-blue', received: 'badge-green', cancelled: 'badge-red' };
const STATUS_LABEL: Record<string, string> = { draft: 'Draft', ordered: 'Dipesan', received: 'Diterima', cancelled: 'Dibatalkan' };

export default function PurchaseOrdersPage() {
  const { selectedStore } = useAuth();
  const [orders, setOrders] = useState<PurchaseOrder[]>([]);
  const [loading, setLoading] = useState(true);
  const [showModal, setShowModal] = useState(false);
  const [suppliers, setSuppliers] = useState<Supplier[]>([]);
  const [products, setProducts] = useState<Product[]>([]);
  const [form, setForm] = useState({ supplier_id: '', notes: '', items: [{ product_id: '', quantity: 1, unit_cost: 0 }] });
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');

  const storeId = selectedStore?.store_id;

  const load = () => {
    if (!storeId) return;
    setLoading(true);
    purchaseOrdersApi.list(storeId, { per_page: 30 }).then(r => setOrders((r.data as any).data ?? [])).catch(console.error).finally(() => setLoading(false));
  };

  useEffect(() => { load(); }, [storeId]);
  useEffect(() => {
    if (!storeId) return;
    Promise.all([suppliersApi.list({ per_page: 100 }), productsApi.list(storeId, { per_page: 200 })]).then(([s, p]) => {
      setSuppliers((s.data as any).data ?? []);
      setProducts((p.data as any).data ?? []);
    });
  }, [storeId]);

  const addItem = () => setForm(f => ({ ...f, items: [...f.items, { product_id: '', quantity: 1, unit_cost: 0 }] }));
  const removeItem = (i: number) => setForm(f => ({ ...f, items: f.items.filter((_, idx) => idx !== i) }));
  const updateItem = (i: number, k: string, v: string) => setForm(f => ({ ...f, items: f.items.map((item, idx) => idx === i ? { ...item, [k]: k === 'product_id' ? v : +v } : item) }));

  const handleCreate = async () => {
    if (!storeId) return;
    setSaving(true); setError('');
    try {
      await purchaseOrdersApi.create(storeId, { supplier_id: form.supplier_id || undefined, notes: form.notes, items: form.items.filter(i => i.product_id) });
      setShowModal(false); load();
    } catch (e) { setError(e instanceof ApiError ? e.message : 'Gagal membuat PO'); }
    finally { setSaving(false); }
  };

  const handleAction = async (po: PurchaseOrder, action: 'submit' | 'receive' | 'cancel') => {
    if (!storeId) return;
    const fns = { submit: () => purchaseOrdersApi.submit(storeId, po.id), receive: () => purchaseOrdersApi.receive(storeId, po.id), cancel: () => purchaseOrdersApi.cancel(storeId, po.id) };
    try { await fns[action](); load(); } catch (e) { alert(e instanceof ApiError ? e.message : 'Gagal'); }
  };

  if (!selectedStore) return <div style={{ padding: 32 }}><div className="empty-state card" style={{ padding: 40 }}><ClipboardList size={40} /><p>Pilih toko terlebih dahulu</p></div></div>;

  return (
    <div style={{ padding: 24 }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: 20 }}>
        <div><h1 className="page-title">Purchase Order</h1><p className="page-subtitle">{selectedStore.store_name}</p></div>
        <button className="btn btn-primary" onClick={() => { setForm({ supplier_id: '', notes: '', items: [{ product_id: '', quantity: 1, unit_cost: 0 }] }); setError(''); setShowModal(true); }}>
          <Plus size={15} /> Buat PO
        </button>
      </div>

      <div className="card" style={{ overflow: 'hidden' }}>
        {loading ? <div style={{ display: 'flex', justifyContent: 'center', padding: 40 }}><Loader2 size={24} className="loading-spin" style={{ color: 'var(--accent-em)' }} /></div>
        : orders.length === 0 ? <div className="empty-state"><ClipboardList size={32} /><p>Belum ada purchase order</p></div>
        : (
          <table className="tbl">
            <thead><tr><th>Nomor PO</th><th>Supplier</th><th>Total</th><th>Status</th><th>Dibuat</th><th>Aksi</th></tr></thead>
            <tbody>
              {orders.map(po => (
                <tr key={po.id}>
                  <td style={{ fontFamily: 'monospace', fontWeight: 600 }}>{po.po_number}</td>
                  <td style={{ color: 'var(--text-2)' }}>{po.supplier_name ?? '–'}</td>
                  <td style={{ fontWeight: 700, color: 'var(--accent-em)' }}>{formatRp(po.total_amount)}</td>
                  <td><span className={`badge ${STATUS_BADGE[po.status]}`}>{STATUS_LABEL[po.status]}</span></td>
                  <td style={{ color: 'var(--text-3)', fontSize: '0.8rem' }}>{formatDate(po.created_at)}</td>
                  <td>
                    <div style={{ display: 'flex', gap: 4 }}>
                      {po.status === 'draft' && <button className="btn btn-secondary btn-sm" onClick={() => handleAction(po, 'submit')}>Submit</button>}
                      {po.status === 'ordered' && <button className="btn btn-primary btn-sm" onClick={() => handleAction(po, 'receive')}>Terima</button>}
                      {(po.status === 'draft' || po.status === 'ordered') && <button className="btn btn-danger btn-sm" onClick={() => handleAction(po, 'cancel')}>Batal</button>}
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
          <div className="modal-box" style={{ maxWidth: 520 }} onClick={e => e.stopPropagation()}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 18 }}>
              <h2 style={{ fontWeight: 700 }}>Buat Purchase Order</h2>
              <button className="btn btn-ghost btn-sm" onClick={() => setShowModal(false)}><X size={15} /></button>
            </div>
            {error && <div style={{ background: 'rgba(239,68,68,0.12)', borderRadius: 8, padding: '8px 12px', color: '#f87171', fontSize: '0.83rem', marginBottom: 14, border: '1px solid rgba(239,68,68,0.3)' }}>{error}</div>}
            <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
              <div className="input-group"><label className="input-label">Supplier</label>
                <select className="input" value={form.supplier_id} onChange={e => setForm(f => ({ ...f, supplier_id: e.target.value }))}>
                  <option value="">Tanpa Supplier</option>
                  {suppliers.map(s => <option key={s.id} value={s.id}>{s.name}</option>)}
                </select>
              </div>
              <div className="input-group"><label className="input-label">Catatan</label>
                <input className="input" value={form.notes} onChange={e => setForm(f => ({ ...f, notes: e.target.value }))} placeholder="Opsional" />
              </div>
              <div>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 8 }}>
                  <label className="input-label">Item</label>
                  <button className="btn btn-ghost btn-sm" onClick={addItem}><Plus size={13} /> Tambah</button>
                </div>
                {form.items.map((item, i) => (
                  <div key={i} style={{ display: 'grid', gridTemplateColumns: '1fr 90px 110px auto', gap: 6, marginBottom: 6 }}>
                    <select className="input" value={item.product_id} onChange={e => updateItem(i, 'product_id', e.target.value)}>
                      <option value="">Pilih produk</option>
                      {products.map(p => <option key={p.id} value={p.id}>{p.name}</option>)}
                    </select>
                    <input type="number" className="input" placeholder="Qty" value={item.quantity} onChange={e => updateItem(i, 'quantity', e.target.value)} />
                    <input type="number" className="input" placeholder="Harga" value={item.unit_cost} onChange={e => updateItem(i, 'unit_cost', e.target.value)} />
                    <button className="btn btn-ghost btn-sm" style={{ color: 'var(--accent-rd)' }} onClick={() => removeItem(i)}><X size={13} /></button>
                  </div>
                ))}
              </div>
            </div>
            <div style={{ display: 'flex', gap: 8, marginTop: 20 }}>
              <button className="btn btn-secondary" style={{ flex: 1 }} onClick={() => setShowModal(false)}>Batal</button>
              <button className="btn btn-primary" style={{ flex: 1 }} disabled={saving} onClick={handleCreate}>
                {saving ? <Loader2 size={15} className="loading-spin" /> : null}
                {saving ? 'Menyimpan...' : 'Buat PO'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
