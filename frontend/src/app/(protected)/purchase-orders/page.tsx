'use client';

import { useEffect, useState, useCallback } from 'react';
import {
  ClipboardList, Plus, Loader2, X, ChevronRight,
  Package, User, Calendar, FileText, Hash, TrendingUp,
} from 'lucide-react';
import { useAuth } from '@/lib/auth/AuthContext';
import { purchaseOrdersApi, suppliersApi } from '@/lib/api/store-apis';
import { productsApi } from '@/lib/api/products';
import { formatRp, formatDate } from '@/lib/utils';
import type { PurchaseOrder, Product, Supplier } from '@/types';
import { ApiError } from '@/lib/api/client';

const STATUS_BADGE: Record<string, string> = {
  draft: 'badge-gray', ordered: 'badge-blue',
  received: 'badge-green', cancelled: 'badge-red',
};
const STATUS_LABEL: Record<string, string> = {
  draft: 'Draft', ordered: 'Dipesan',
  received: 'Diterima', cancelled: 'Dibatalkan',
};

// ── Item row in the create form ───────────────────────────────────────────────
interface ItemRow { product_id: string; quantity: number; unit_cost: number; }

const EMPTY_FORM = { supplier_id: '', notes: '', items: [{ product_id: '', quantity: 1, unit_cost: 0 }] as ItemRow[] };

// ── Detail Drawer ─────────────────────────────────────────────────────────────
function PODetailDrawer({ po, onClose }: { po: PurchaseOrder; onClose: () => void }) {
  return (
    <>
      {/* Backdrop */}
      <div onClick={onClose} style={{
        position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.5)',
        backdropFilter: 'blur(2px)', zIndex: 200,
      }} />
      {/* Drawer */}
      <div style={{
        position: 'fixed', top: 0, right: 0, bottom: 0, width: 520,
        background: 'var(--bg-card)', borderLeft: '1px solid var(--border)',
        zIndex: 201, overflowY: 'auto', display: 'flex', flexDirection: 'column',
      }}>
        {/* Drawer header */}
        <div style={{
          display: 'flex', justifyContent: 'space-between', alignItems: 'center',
          padding: '18px 20px', borderBottom: '1px solid var(--border)',
          position: 'sticky', top: 0, background: 'var(--bg-card)', zIndex: 1,
        }}>
          <div>
            <div style={{ fontWeight: 800, fontSize: '1rem' }}>{po.po_number}</div>
            <span className={`badge ${STATUS_BADGE[po.status]}`} style={{ marginTop: 4 }}>
              {STATUS_LABEL[po.status]}
            </span>
          </div>
          <button className="btn btn-ghost btn-sm" onClick={onClose}><X size={16} /></button>
        </div>

        <div style={{ padding: '20px', display: 'flex', flexDirection: 'column', gap: 20 }}>
          {/* Meta info grid */}
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
            {[
              { icon: User,      label: 'Supplier',    val: po.supplier_name ?? '—' },
              { icon: Hash,      label: 'Nomor PO',    val: po.po_number },
              { icon: User,      label: 'Dibuat Oleh', val: po.ordered_by_name ?? '—' },
              { icon: Calendar,  label: 'Tanggal Buat',val: formatDate(po.created_at) },
              po.ordered_at
                ? { icon: Calendar, label: 'Dipesan',  val: formatDate(po.ordered_at) }
                : null,
              po.received_at
                ? { icon: Calendar, label: 'Diterima', val: formatDate(po.received_at) }
                : null,
              po.received_by_name
                ? { icon: User, label: 'Diterima Oleh', val: po.received_by_name }
                : null,
            ].filter(Boolean).map((m: any) => (
              <div key={m.label} style={{
                background: 'var(--bg-elevated)', borderRadius: 10, padding: '10px 14px',
                display: 'flex', flexDirection: 'column', gap: 4,
              }}>
                <div style={{ fontSize: '0.7rem', color: 'var(--text-3)', display: 'flex', alignItems: 'center', gap: 4 }}>
                  <m.icon size={11} /> {m.label}
                </div>
                <div style={{ fontWeight: 600, fontSize: '0.85rem' }}>{m.val}</div>
              </div>
            ))}
          </div>

          {/* Notes */}
          {po.notes && (
            <div style={{ background: 'var(--bg-elevated)', borderRadius: 10, padding: '12px 14px' }}>
              <div style={{ fontSize: '0.72rem', color: 'var(--text-3)', marginBottom: 6, display: 'flex', alignItems: 'center', gap: 4 }}>
                <FileText size={11} /> Catatan
              </div>
              <div style={{ fontSize: '0.85rem' }}>{po.notes}</div>
            </div>
          )}

          {/* Line items */}
          <div>
            <div style={{ fontWeight: 700, marginBottom: 10, fontSize: '0.9rem' }}>
              Item ({po.items?.length ?? 0})
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
              {(po.items ?? []).map((item, i) => (
                <div key={item.id ?? i} style={{
                  background: 'var(--bg-elevated)', borderRadius: 10, padding: '12px 14px',
                  display: 'grid', gridTemplateColumns: '1fr auto', gap: 8, alignItems: 'start',
                }}>
                  <div>
                    <div style={{ fontWeight: 600, fontSize: '0.88rem' }}>{item.product_name}</div>
                    <div style={{ fontSize: '0.75rem', color: 'var(--text-3)', marginTop: 3 }}>
                      SKU: {item.product_sku} · Unit: {item.unit}
                    </div>
                    {po.status === 'received' && item.received_qty > 0 && (
                      <div style={{ fontSize: '0.73rem', color: '#10b981', marginTop: 3 }}>
                        Diterima: {item.received_qty} {item.unit}
                      </div>
                    )}
                  </div>
                  <div style={{ textAlign: 'right' }}>
                    <div style={{ fontWeight: 700, color: 'var(--accent-em)' }}>
                      {formatRp(item.subtotal)}
                    </div>
                    <div style={{ fontSize: '0.75rem', color: 'var(--text-3)', marginTop: 2 }}>
                      {item.quantity} × {formatRp(item.unit_cost)}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>

          {/* Total */}
          <div style={{
            background: 'rgba(16,185,129,0.08)', border: '1px solid rgba(16,185,129,0.25)',
            borderRadius: 12, padding: '14px 18px',
            display: 'flex', justifyContent: 'space-between', alignItems: 'center',
          }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 8, color: '#10b981', fontWeight: 700 }}>
              <TrendingUp size={16} /> Total Pembelian
            </div>
            <div style={{ fontWeight: 800, fontSize: '1.15rem', color: '#10b981' }}>
              {formatRp(po.total_amount)}
            </div>
          </div>
        </div>
      </div>
    </>
  );
}

// ── Main Page ─────────────────────────────────────────────────────────────────
export default function PurchaseOrdersPage() {
  const { selectedStore } = useAuth();
  const [orders, setOrders]       = useState<PurchaseOrder[]>([]);
  const [loading, setLoading]     = useState(true);
  const [showModal, setShowModal] = useState(false);
  const [detailPO, setDetailPO]   = useState<PurchaseOrder | null>(null);
  const [loadingDetail, setLoadingDetail] = useState(false);
  const [suppliers, setSuppliers] = useState<Supplier[]>([]);
  const [products, setProducts]   = useState<Product[]>([]);
  const [form, setForm]           = useState(EMPTY_FORM);
  const [saving, setSaving]       = useState(false);
  const [error, setError]         = useState('');

  const storeId = selectedStore?.store_id;

  const load = useCallback(() => {
    if (!storeId) return;
    setLoading(true);
    purchaseOrdersApi.list(storeId, { per_page: 50 })
      .then(r => setOrders((r.data as any).data ?? []))
      .catch(console.error)
      .finally(() => setLoading(false));
  }, [storeId]);

  useEffect(() => { load(); }, [load]);

  useEffect(() => {
    if (!storeId) return;
    Promise.all([
      suppliersApi.list({ per_page: 100 }),
      productsApi.list(storeId, { per_page: 200 }),
    ]).then(([s, p]) => {
      setSuppliers((s.data as any).data ?? []);
      setProducts((p.data as any).data ?? []);
    });
  }, [storeId]);

  // ── Open detail drawer ────────────────────────────────────────────────────
  const openDetail = async (po: PurchaseOrder) => {
    setDetailPO(po); // show immediately with list data
    setLoadingDetail(true);
    try {
      const r = await purchaseOrdersApi.get(storeId!, po.id);
      setDetailPO(r.data as PurchaseOrder);
    } catch (e) { console.error(e); }
    finally { setLoadingDetail(false); }
  };

  // ── Item helpers ──────────────────────────────────────────────────────────
  const addItem = () =>
    setForm(f => ({ ...f, items: [...f.items, { product_id: '', quantity: 1, unit_cost: 0 }] }));

  const removeItem = (i: number) =>
    setForm(f => ({ ...f, items: f.items.filter((_, idx) => idx !== i) }));

  const updateItem = (i: number, k: keyof ItemRow, v: string) => {
    setForm(f => {
      const items = f.items.map((item, idx) => {
        if (idx !== i) return item;
        if (k === 'product_id') {
          // Auto-fill unit_cost from product's cost_price
          const product = products.find(p => p.id === v);
          return { ...item, product_id: v, unit_cost: product ? product.cost_price : item.unit_cost };
        }
        return { ...item, [k]: +v };
      });
      return { ...f, items };
    });
  };

  // Running total
  const runningTotal = form.items.reduce((s, item) => s + (item.quantity * item.unit_cost), 0);

  // ── Submit ────────────────────────────────────────────────────────────────
  const handleCreate = async () => {
    if (!storeId) return;
    setSaving(true); setError('');
    try {
      await purchaseOrdersApi.create(storeId, {
        supplier_id: form.supplier_id || undefined,
        notes: form.notes,
        items: form.items.filter(i => i.product_id),
      });
      setShowModal(false);
      setForm(EMPTY_FORM);
      load();
    } catch (e) {
      setError(e instanceof ApiError ? e.message : 'Gagal membuat PO');
    } finally { setSaving(false); }
  };

  const handleAction = async (po: PurchaseOrder, action: 'submit' | 'receive' | 'cancel') => {
    if (!storeId) return;
    const fns = {
      submit:  () => purchaseOrdersApi.submit(storeId, po.id),
      receive: () => purchaseOrdersApi.receive(storeId, po.id),
      cancel:  () => purchaseOrdersApi.cancel(storeId, po.id),
    };
    try {
      await fns[action]();
      load();
      // Refresh drawer if open
      if (detailPO?.id === po.id) openDetail({ ...po });
    } catch (e) {
      alert(e instanceof ApiError ? e.message : 'Gagal');
    }
  };

  if (!selectedStore) return (
    <div style={{ padding: 32 }}>
      <div className="empty-state card" style={{ padding: 40 }}>
        <ClipboardList size={40} /><p>Pilih toko terlebih dahulu</p>
      </div>
    </div>
  );

  return (
    <div style={{ padding: 24 }}>
      {/* Header */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: 20 }}>
        <div>
          <h1 className="page-title">Purchase Order</h1>
          <p className="page-subtitle">{selectedStore.store_name}</p>
        </div>
        <button className="btn btn-primary" onClick={() => { setForm(EMPTY_FORM); setError(''); setShowModal(true); }}>
          <Plus size={15} /> Buat PO
        </button>
      </div>

      {/* List */}
      <div className="card" style={{ overflow: 'hidden' }}>
        {loading ? (
          <div style={{ display: 'flex', justifyContent: 'center', padding: 40 }}>
            <Loader2 size={24} className="loading-spin" style={{ color: 'var(--accent-em)' }} />
          </div>
        ) : orders.length === 0 ? (
          <div className="empty-state"><ClipboardList size={32} /><p>Belum ada purchase order</p></div>
        ) : (
          <table className="tbl">
            <thead>
              <tr>
                <th>Nomor PO</th>
                <th>Supplier</th>
                <th>Items</th>
                <th>Total</th>
                <th>Status</th>
                <th>Dibuat</th>
                <th>Aksi</th>
              </tr>
            </thead>
            <tbody>
              {orders.map(po => (
                <tr key={po.id} style={{ cursor: 'pointer' }}>
                  {/* Clickable cells open the drawer */}
                  <td onClick={() => openDetail(po)} style={{ fontFamily: 'monospace', fontWeight: 700, color: 'var(--accent-em)' }}>
                    {po.po_number}
                  </td>
                  <td onClick={() => openDetail(po)} style={{ color: 'var(--text-2)' }}>{po.supplier_name ?? '—'}</td>
                  <td onClick={() => openDetail(po)}>
                    <span style={{
                      background: 'var(--bg-elevated)', borderRadius: 6, padding: '2px 8px',
                      fontSize: '0.78rem', color: 'var(--text-2)',
                    }}>
                      {po.items?.length ?? 0} item
                    </span>
                  </td>
                  <td onClick={() => openDetail(po)} style={{ fontWeight: 700, color: 'var(--accent-em)' }}>
                    {formatRp(po.total_amount)}
                  </td>
                  <td onClick={() => openDetail(po)}>
                    <span className={`badge ${STATUS_BADGE[po.status]}`}>{STATUS_LABEL[po.status]}</span>
                  </td>
                  <td onClick={() => openDetail(po)} style={{ color: 'var(--text-3)', fontSize: '0.8rem' }}>
                    {formatDate(po.created_at)}
                  </td>
                  <td>
                    <div style={{ display: 'flex', gap: 4 }}>
                      {po.status === 'draft'   && <button className="btn btn-secondary btn-sm" onClick={e => { e.stopPropagation(); handleAction(po, 'submit'); }}>Submit</button>}
                      {po.status === 'ordered' && <button className="btn btn-primary btn-sm"   onClick={e => { e.stopPropagation(); handleAction(po, 'receive'); }}>Terima</button>}
                      {(po.status === 'draft' || po.status === 'ordered') &&
                        <button className="btn btn-danger btn-sm" onClick={e => { e.stopPropagation(); handleAction(po, 'cancel'); }}>Batal</button>
                      }
                      <button className="btn btn-ghost btn-sm" onClick={() => openDetail(po)} title="Lihat Detail">
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

      {/* ── Detail Drawer ── */}
      {detailPO && (
        <PODetailDrawer
          po={detailPO}
          onClose={() => setDetailPO(null)}
        />
      )}

      {/* ── Create Modal ── */}
      {showModal && (
        <div className="modal-overlay" onClick={() => setShowModal(false)}>
          <div className="modal-box" style={{ maxWidth: 580, maxHeight: '90vh', overflowY: 'auto' }}
               onClick={e => e.stopPropagation()}>
            {/* Header */}
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 18 }}>
              <h2 style={{ fontWeight: 800 }}>Buat Purchase Order</h2>
              <button className="btn btn-ghost btn-sm" onClick={() => setShowModal(false)}><X size={15} /></button>
            </div>

            {error && (
              <div style={{
                background: 'rgba(239,68,68,0.12)', borderRadius: 8, padding: '8px 12px',
                color: '#f87171', fontSize: '0.83rem', marginBottom: 14,
                border: '1px solid rgba(239,68,68,0.3)',
              }}>{error}</div>
            )}

            <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
              {/* Supplier */}
              <div className="input-group">
                <label className="input-label">Supplier (opsional)</label>
                <select className="input" value={form.supplier_id}
                  onChange={e => setForm(f => ({ ...f, supplier_id: e.target.value }))}>
                  <option value="">Tanpa Supplier</option>
                  {suppliers.map(s => <option key={s.id} value={s.id}>{s.name}</option>)}
                </select>
              </div>

              {/* Notes */}
              <div className="input-group">
                <label className="input-label">Catatan (opsional)</label>
                <input className="input" value={form.notes}
                  onChange={e => setForm(f => ({ ...f, notes: e.target.value }))}
                  placeholder="Catatan pemesanan..." />
              </div>

              {/* Items */}
              <div>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 10 }}>
                  <label className="input-label" style={{ margin: 0 }}>Item Pembelian</label>
                  <button className="btn btn-ghost btn-sm" onClick={addItem}><Plus size={13} /> Tambah Item</button>
                </div>

                {/* Column headers */}
                <div style={{
                  display: 'grid', gridTemplateColumns: '1fr 90px 130px auto',
                  gap: 6, marginBottom: 6, padding: '0 4px',
                }}>
                  <span style={{ fontSize: '0.72rem', color: 'var(--text-3)', fontWeight: 600 }}>PRODUK</span>
                  <span style={{ fontSize: '0.72rem', color: 'var(--text-3)', fontWeight: 600 }}>QTY</span>
                  <span style={{ fontSize: '0.72rem', color: 'var(--text-3)', fontWeight: 600 }}>HARGA BELI / UNIT</span>
                  <span />
                </div>

                {form.items.map((item, i) => {
                  const selectedProduct = products.find(p => p.id === item.product_id);
                  const lineTotal = item.quantity * item.unit_cost;
                  return (
                    <div key={i} style={{ marginBottom: 8 }}>
                      <div style={{ display: 'grid', gridTemplateColumns: '1fr 90px 130px auto', gap: 6 }}>
                        {/* Product selector */}
                        <select className="input" value={item.product_id}
                          onChange={e => updateItem(i, 'product_id', e.target.value)}>
                          <option value="">Pilih produk...</option>
                          {products.map(p => (
                            <option key={p.id} value={p.id}>{p.name}</option>
                          ))}
                        </select>

                        {/* Quantity */}
                        <input type="number" className="input" placeholder="Qty" min={1}
                          value={item.quantity}
                          onChange={e => updateItem(i, 'quantity', e.target.value)} />

                        {/* Unit cost */}
                        <input type="number" className="input" placeholder="Harga beli" min={0}
                          value={item.unit_cost}
                          onChange={e => updateItem(i, 'unit_cost', e.target.value)} />

                        {/* Remove */}
                        <button className="btn btn-ghost btn-sm"
                          style={{ color: 'var(--accent-rd)', alignSelf: 'center' }}
                          onClick={() => removeItem(i)}>
                          <X size={13} />
                        </button>
                      </div>

                      {/* Sub-row: product info + line total */}
                      {selectedProduct && (
                        <div style={{
                          display: 'flex', justifyContent: 'space-between', alignItems: 'center',
                          padding: '4px 4px 0', fontSize: '0.74rem', color: 'var(--text-3)',
                        }}>
                          <span>SKU: {selectedProduct.sku} · Stok: {selectedProduct.stock_qty ?? '?'} {selectedProduct.unit} · HPP saat ini: {formatRp(selectedProduct.cost_price)}</span>
                          {lineTotal > 0 && (
                            <span style={{ color: 'var(--accent-em)', fontWeight: 600 }}>
                              = {formatRp(lineTotal)}
                            </span>
                          )}
                        </div>
                      )}
                    </div>
                  );
                })}

                {/* Running total */}
                {runningTotal > 0 && (
                  <div style={{
                    marginTop: 10, padding: '10px 14px',
                    background: 'rgba(16,185,129,0.08)',
                    border: '1px solid rgba(16,185,129,0.25)',
                    borderRadius: 10, display: 'flex', justifyContent: 'space-between',
                  }}>
                    <span style={{ fontSize: '0.85rem', color: 'var(--text-2)' }}>Total Estimasi</span>
                    <span style={{ fontWeight: 800, color: '#10b981' }}>{formatRp(runningTotal)}</span>
                  </div>
                )}
              </div>
            </div>

            {/* Footer */}
            <div style={{ display: 'flex', gap: 8, marginTop: 20 }}>
              <button className="btn btn-secondary" style={{ flex: 1 }} onClick={() => setShowModal(false)}>Batal</button>
              <button className="btn btn-primary" style={{ flex: 1 }} disabled={saving} onClick={handleCreate}>
                {saving ? <Loader2 size={15} className="loading-spin" /> : <Plus size={15} />}
                {saving ? 'Menyimpan...' : 'Buat PO'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
