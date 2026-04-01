'use client';

import { useEffect, useState, useCallback } from 'react';
import {
  ClipboardList, Plus, Loader2, X, ChevronRight,
  Package, User, Calendar, FileText, Hash, TrendingUp,
  AlertTriangle, CheckCircle2, Printer, Wallet, CreditCard,
  AlertCircle, BadgeCheck, Clock,
} from 'lucide-react';
import { useAuth } from '@/lib/auth/AuthContext';
import { purchaseOrdersApi, suppliersApi, storesApi } from '@/lib/api/store-apis';
import { productsApi } from '@/lib/api/products';
import { formatRp, formatDate } from '@/lib/utils';
import type { PurchaseOrder, Product, Supplier, Store } from '@/types';
import { ApiError } from '@/lib/api/client';

// ── Constants ─────────────────────────────────────────────────────────────────
const STATUS_BADGE: Record<string, string> = {
  draft: 'badge-gray', ordered: 'badge-blue',
  received: 'badge-green', cancelled: 'badge-red',
};
const STATUS_LABEL: Record<string, string> = {
  draft: 'Draft', ordered: 'Dipesan',
  received: 'Diterima', cancelled: 'Dibatalkan',
};
const PAY_STATUS_CFG = {
  unpaid:  { label: 'Belum Bayar',   color: '#ef4444', bg: 'rgba(239,68,68,0.12)',   icon: AlertCircle },
  partial: { label: 'Sebagian',      color: '#f59e0b', bg: 'rgba(245,158,11,0.12)',  icon: Clock },
  paid:    { label: 'Lunas',         color: '#10b981', bg: 'rgba(16,185,129,0.12)',  icon: BadgeCheck },
};
const EMPTY_FORM = {
  supplier_id: '', notes: '',
  items: [{ product_id: '', quantity: 1, unit_cost: 0 }],
};
type ItemRow    = { product_id: string; quantity: number; unit_cost: number };
type ActionType = 'submit' | 'receive' | 'cancel';

interface PayableSummary {
  total_debt: number; total_paid: number; total_outstanding: number;
  unpaid_count: number; partial_count: number;
}
interface POPayment { id: string; po_id: string; amount: number; note?: string; paid_by_name: string; paid_at: string; }

// ── PayStatus badge ───────────────────────────────────────────────────────────
function PayStatusBadge({ status }: { status: string }) {
  const cfg = PAY_STATUS_CFG[status as keyof typeof PAY_STATUS_CFG] ?? PAY_STATUS_CFG.unpaid;
  const Icon = cfg.icon;
  return (
    <span style={{
      display: 'inline-flex', alignItems: 'center', gap: 4,
      padding: '2px 8px', borderRadius: 6, fontSize: '0.72rem', fontWeight: 600,
      background: cfg.bg, color: cfg.color, whiteSpace: 'nowrap',
    }}>
      <Icon size={11} /> {cfg.label}
    </span>
  );
}

// ── Confirm Modal ─────────────────────────────────────────────────────────────
interface ConfirmModalProps {
  action: ActionType; po: PurchaseOrder;
  onConfirm: () => void; onCancel: () => void; loading: boolean;
}
function ConfirmModal({ action, po, onConfirm, onCancel, loading }: ConfirmModalProps) {
  const config = {
    submit: {
      icon: <CheckCircle2 size={28} style={{ color: '#6366f1' }} />,
      title: 'Submit Purchase Order?',
      body: `PO ${po.po_number} akan dikirim ke supplier. Status berubah menjadi Dipesan.`,
      confirmLabel: 'Ya, Submit PO', confirmClass: 'btn btn-primary',
    },
    receive: {
      icon: <Package size={28} style={{ color: '#10b981' }} />,
      title: 'Terima Purchase Order?',
      body: `Konfirmasi penerimaan PO ${po.po_number}. Stok akan bertambah, HPP produk akan diperbarui, dan hutang ke supplier akan tercatat.`,
      confirmLabel: 'Ya, Terima PO', confirmClass: 'btn btn-primary',
    },
    cancel: {
      icon: <AlertTriangle size={28} style={{ color: '#ef4444' }} />,
      title: 'Batalkan Purchase Order?',
      body: `PO ${po.po_number} akan dibatalkan. Tindakan ini tidak dapat dibatalkan.`,
      confirmLabel: 'Ya, Batalkan', confirmClass: 'btn btn-danger',
    },
  }[action];
  return (
    <div className="modal-overlay" onClick={onCancel}>
      <div className="modal-box" style={{ maxWidth: 420 }} onClick={e => e.stopPropagation()}>
        <div style={{ textAlign: 'center', padding: '8px 0 16px' }}>
          <div style={{ marginBottom: 12 }}>{config.icon}</div>
          <h2 style={{ fontWeight: 800, fontSize: '1.05rem', marginBottom: 10 }}>{config.title}</h2>
          <p style={{ color: 'var(--text-2)', fontSize: '0.875rem', lineHeight: 1.6 }}>{config.body}</p>
        </div>
        <div style={{ display: 'flex', gap: 8 }}>
          <button className="btn btn-secondary" style={{ flex: 1 }} onClick={onCancel} disabled={loading}>Batal</button>
          <button className={config.confirmClass} style={{ flex: 1 }} onClick={onConfirm} disabled={loading}>
            {loading ? <Loader2 size={14} className="loading-spin" /> : null}
            {loading ? 'Memproses...' : config.confirmLabel}
          </button>
        </div>
      </div>
    </div>
  );
}

// ── Pay Modal ─────────────────────────────────────────────────────────────────
interface PayModalProps {
  po: PurchaseOrder; storeId: string;
  onSuccess: () => void; onCancel: () => void;
}
function PayModal({ po, storeId, onSuccess, onCancel }: PayModalProps) {
  const amountDue = (po as any).amount_due ?? ((po.total_amount ?? 0) - ((po as any).amount_paid ?? 0));
  const [amount, setAmount] = useState(String(amountDue > 0 ? amountDue.toFixed(0) : ''));
  const [note, setNote]     = useState('');
  const [saving, setSaving] = useState(false);
  const [error, setError]   = useState('');

  const handleSubmit = async () => {
    const n = parseFloat(amount);
    if (!n || n <= 0) { setError('Jumlah harus lebih dari 0'); return; }
    setSaving(true); setError('');
    try {
      await purchaseOrdersApi.createPayment(storeId, po.id, { amount: n, note: note || undefined });
      onSuccess();
    } catch (e) {
      setError(e instanceof ApiError ? e.message : 'Gagal menyimpan pembayaran');
    } finally { setSaving(false); }
  };

  return (
    <div className="modal-overlay" onClick={onCancel}>
      <div className="modal-box" style={{ maxWidth: 400 }} onClick={e => e.stopPropagation()}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
          <h2 style={{ fontWeight: 800 }}>Bayar Hutang</h2>
          <button className="btn btn-ghost btn-sm" onClick={onCancel}><X size={15} /></button>
        </div>
        {/* PO summary */}
        <div style={{ background: 'var(--bg-elevated)', borderRadius: 10, padding: '12px 14px', marginBottom: 16 }}>
          <div style={{ fontSize: '0.8rem', fontWeight: 700 }}>{po.po_number}</div>
          <div style={{ display: 'flex', justifyContent: 'space-between', marginTop: 6, fontSize: '0.82rem' }}>
            <span style={{ color: 'var(--text-3)' }}>Total PO</span>
            <span style={{ fontWeight: 600 }}>{formatRp(po.total_amount)}</span>
          </div>
          <div style={{ display: 'flex', justifyContent: 'space-between', marginTop: 4, fontSize: '0.82rem' }}>
            <span style={{ color: 'var(--text-3)' }}>Sudah dibayar</span>
            <span style={{ color: '#10b981', fontWeight: 600 }}>{formatRp((po as any).amount_paid ?? 0)}</span>
          </div>
          <div style={{ display: 'flex', justifyContent: 'space-between', marginTop: 4, fontSize: '0.9rem', borderTop: '1px solid var(--border)', paddingTop: 8 }}>
            <span style={{ fontWeight: 700 }}>Sisa hutang</span>
            <span style={{ fontWeight: 800, color: '#ef4444' }}>{formatRp(amountDue)}</span>
          </div>
        </div>
        {error && (
          <div style={{ background: 'rgba(239,68,68,0.1)', borderRadius: 8, padding: '8px 12px', color: '#f87171', fontSize: '0.83rem', marginBottom: 12, border: '1px solid rgba(239,68,68,0.3)' }}>
            {error}
          </div>
        )}
        <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
          <div className="input-group">
            <label className="input-label">Jumlah Bayar (Rp)</label>
            <input type="number" className="input" value={amount} onChange={e => setAmount(e.target.value)} min={1} />
          </div>
          <div className="input-group">
            <label className="input-label">Catatan (opsional)</label>
            <input className="input" value={note} onChange={e => setNote(e.target.value)} placeholder="Transfer, tunai, dll..." />
          </div>
        </div>
        <div style={{ display: 'flex', gap: 8, marginTop: 16 }}>
          <button className="btn btn-secondary" style={{ flex: 1 }} onClick={onCancel}>Batal</button>
          <button className="btn btn-primary" style={{ flex: 1 }} onClick={handleSubmit} disabled={saving}>
            {saving ? <Loader2 size={14} className="loading-spin" /> : <Wallet size={14} />}
            {saving ? 'Menyimpan...' : 'Bayar'}
          </button>
        </div>
      </div>
    </div>
  );
}

// ── Invoice Modal ─────────────────────────────────────────────────────────────
interface InvoiceModalProps { po: PurchaseOrder; store: Store | null; onClose: () => void; }
function InvoiceModal({ po, store, onClose }: InvoiceModalProps) {
  return (
    <>
      <div className="modal-overlay no-print" onClick={onClose} />
      <div style={{ position: 'fixed', inset: 0, zIndex: 202, display: 'flex', alignItems: 'center', justifyContent: 'center', pointerEvents: 'none' }}>
        <div id="po-invoice" style={{ background: '#fff', color: '#111', borderRadius: 12, width: '100%', maxWidth: 680, maxHeight: '92vh', overflowY: 'auto', pointerEvents: 'auto', boxShadow: '0 24px 80px rgba(0,0,0,0.4)', fontFamily: '"Inter","Helvetica Neue",Arial,sans-serif' }}>
          <div className="no-print" style={{ display: 'flex', justifyContent: 'flex-end', gap: 8, padding: '12px 16px', borderBottom: '1px solid #e5e7eb' }}>
            <button onClick={() => window.print()} style={{ display: 'flex', alignItems: 'center', gap: 6, padding: '8px 16px', borderRadius: 8, border: 'none', background: '#111', color: '#fff', fontWeight: 600, fontSize: '0.85rem', cursor: 'pointer' }}>
              <Printer size={14} /> Cetak Invoice
            </button>
            <button onClick={onClose} style={{ padding: '8px 12px', borderRadius: 8, border: '1px solid #e5e7eb', background: 'transparent', cursor: 'pointer' }}><X size={14} /></button>
          </div>
          <div style={{ padding: '32px 40px' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 24 }}>
              <div>
                <div style={{ fontWeight: 800, fontSize: '1.4rem', marginBottom: 4 }}>{store?.name ?? 'Toko'}</div>
                {store?.address && <div style={{ fontSize: '0.82rem', color: '#6b7280' }}>{store.address}</div>}
                {store?.phone   && <div style={{ fontSize: '0.82rem', color: '#6b7280' }}>Telp: {store.phone}</div>}
              </div>
              <div style={{ textAlign: 'right' }}>
                <div style={{ display: 'inline-block', padding: '4px 12px', borderRadius: 6, background: '#f3f4f6', fontSize: '0.75rem', fontWeight: 700, color: '#374151', marginBottom: 8 }}>PURCHASE ORDER</div>
                <div style={{ fontWeight: 800, fontSize: '1.15rem' }}>{po.po_number}</div>
                <div style={{ fontSize: '0.8rem', color: '#6b7280' }}>Tanggal: {formatDate(po.created_at)}</div>
              </div>
            </div>
            <div style={{ borderTop: '2px solid #111', marginBottom: 20 }} />
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 24, marginBottom: 24 }}>
              <div>
                <div style={{ fontSize: '0.7rem', fontWeight: 700, color: '#9ca3af', letterSpacing: '0.08em', marginBottom: 6 }}>DARI (PEMBELI)</div>
                <div style={{ fontWeight: 700 }}>{store?.name ?? '—'}</div>
                {store?.address && <div style={{ fontSize: '0.8rem', color: '#4b5563' }}>{store.address}</div>}
              </div>
              <div>
                <div style={{ fontSize: '0.7rem', fontWeight: 700, color: '#9ca3af', letterSpacing: '0.08em', marginBottom: 6 }}>KEPADA (SUPPLIER)</div>
                <div style={{ fontWeight: 700 }}>{po.supplier_name ?? 'Tanpa Supplier'}</div>
              </div>
            </div>
            <table style={{ width: '100%', borderCollapse: 'collapse', marginBottom: 20 }}>
              <thead>
                <tr style={{ borderBottom: '2px solid #e5e7eb' }}>
                  {['#','Nama Produk','SKU','Qty','Satuan','Harga Beli','Subtotal'].map(h => (
                    <th key={h} style={{ padding: '8px 10px', textAlign: h==='Harga Beli'||h==='Subtotal'?'right':h==='#'||h==='Qty'?'center':'left', fontSize: '0.72rem', fontWeight: 700, color: '#6b7280', textTransform: 'uppercase' }}>{h}</th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {(po.items ?? []).map((item, i) => (
                  <tr key={item.id ?? i} style={{ borderBottom: '1px solid #f3f4f6' }}>
                    <td style={{ padding: '10px', textAlign: 'center', color: '#9ca3af', fontSize: '0.8rem' }}>{i+1}</td>
                    <td style={{ padding: '10px', fontWeight: 600, fontSize: '0.85rem' }}>{item.product_name}</td>
                    <td style={{ padding: '10px', color: '#6b7280', fontSize: '0.78rem', fontFamily: 'monospace' }}>{item.product_sku}</td>
                    <td style={{ padding: '10px', textAlign: 'center', fontWeight: 600 }}>{item.quantity}</td>
                    <td style={{ padding: '10px', color: '#6b7280' }}>{item.unit}</td>
                    <td style={{ padding: '10px', textAlign: 'right' }}>{formatRp(item.unit_cost)}</td>
                    <td style={{ padding: '10px', textAlign: 'right', fontWeight: 700 }}>{formatRp(item.subtotal)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
            <div style={{ display: 'flex', justifyContent: 'flex-end', marginBottom: 20 }}>
              <div style={{ minWidth: 260, borderTop: '2px solid #111', paddingTop: 12 }}>
                <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                  <span style={{ fontWeight: 800 }}>TOTAL PEMBELIAN</span>
                  <span style={{ fontWeight: 800 }}>{formatRp(po.total_amount)}</span>
                </div>
              </div>
            </div>
            {po.notes && <div style={{ background: '#f9fafb', borderRadius: 8, padding: '12px 16px', marginBottom: 20 }}><div style={{ fontSize: '0.72rem', color: '#9ca3af', fontWeight: 700, marginBottom: 4 }}>CATATAN</div><div style={{ fontSize: '0.85rem' }}>{po.notes}</div></div>}
            <div style={{ borderTop: '1px solid #e5e7eb', paddingTop: 16, display: 'flex', justifyContent: 'space-between', alignItems: 'flex-end' }}>
              <div style={{ fontSize: '0.72rem', color: '#9ca3af' }}>Dokumen dibuat otomatis oleh MoedahPOS</div>
              <div style={{ textAlign: 'center' }}>
                <div style={{ borderTop: '1px solid #9ca3af', width: 140, marginBottom: 4 }} />
                <div style={{ fontSize: '0.72rem', color: '#9ca3af' }}>Tanda Tangan Supplier</div>
              </div>
            </div>
          </div>
        </div>
      </div>
      <style>{`@media print{body>*:not(#portal-root){display:none!important}.no-print{display:none!important}#po-invoice{position:fixed!important;inset:0!important;max-height:none!important;border-radius:0!important;box-shadow:none!important;overflow:visible!important}}`}</style>
    </>
  );
}

// ── Detail Drawer ─────────────────────────────────────────────────────────────
interface PODetailDrawerProps {
  po: PurchaseOrder; storeId: string; payments: POPayment[];
  onClose: () => void; onInvoice: () => void;
  onAction: (action: ActionType) => void; onPay: () => void;
}
function PODetailDrawer({ po, storeId, payments, onClose, onInvoice, onAction, onPay }: PODetailDrawerProps) {
  const payStatus  = (po as any).payment_status ?? 'unpaid';
  const amountPaid = (po as any).amount_paid ?? 0;
  const amountDue  = (po as any).amount_due  ?? (po.total_amount - amountPaid);

  return (
    <>
      <div onClick={onClose} style={{ position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.5)', backdropFilter: 'blur(2px)', zIndex: 200 }} />
      <div style={{ position: 'fixed', top: 0, right: 0, bottom: 0, width: 520, background: 'var(--bg-card)', borderLeft: '1px solid var(--border)', zIndex: 201, overflowY: 'auto', display: 'flex', flexDirection: 'column' }}>
        {/* Header */}
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '18px 20px', borderBottom: '1px solid var(--border)', position: 'sticky', top: 0, background: 'var(--bg-card)', zIndex: 1 }}>
          <div>
            <div style={{ fontWeight: 800, fontSize: '1rem' }}>{po.po_number}</div>
            <div style={{ display: 'flex', gap: 6, marginTop: 4 }}>
              <span className={`badge ${STATUS_BADGE[po.status]}`}>{STATUS_LABEL[po.status]}</span>
              {po.status === 'received' && <PayStatusBadge status={payStatus} />}
            </div>
          </div>
          <div style={{ display: 'flex', gap: 6 }}>
            <button className="btn btn-secondary btn-sm" onClick={onInvoice}><Printer size={13} /> Invoice</button>
            <button className="btn btn-ghost btn-sm" onClick={onClose}><X size={16} /></button>
          </div>
        </div>

        <div style={{ padding: 20, display: 'flex', flexDirection: 'column', gap: 18, flex: 1 }}>
          {/* Meta */}
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 10 }}>
            {[
              { icon: Hash,     label: 'Nomor PO',     val: po.po_number },
              { icon: User,     label: 'Supplier',     val: (po as any).supplier_name ?? '—' },
              { icon: User,     label: 'Dibuat Oleh',  val: (po as any).ordered_by_name ?? '—' },
              { icon: Calendar, label: 'Tanggal Buat', val: formatDate(po.created_at) },
            ].map(({ icon: Icon, label, val }) => (
              <div key={label} style={{ background: 'var(--bg-elevated)', borderRadius: 10, padding: '10px 14px' }}>
                <div style={{ fontSize: '0.68rem', color: 'var(--text-3)', display: 'flex', alignItems: 'center', gap: 4, marginBottom: 4 }}><Icon size={11} /> {label}</div>
                <div style={{ fontWeight: 600, fontSize: '0.85rem' }}>{val}</div>
              </div>
            ))}
          </div>

          {/* Items */}
          <div>
            <div style={{ fontWeight: 700, marginBottom: 8, fontSize: '0.9rem' }}>Item ({po.items?.length ?? 0})</div>
            {(po.items ?? []).map((item, i) => (
              <div key={item.id ?? i} style={{ background: 'var(--bg-elevated)', borderRadius: 10, padding: '11px 14px', display: 'grid', gridTemplateColumns: '1fr auto', gap: 8, marginBottom: 6 }}>
                <div>
                  <div style={{ fontWeight: 600, fontSize: '0.87rem' }}>{item.product_name}</div>
                  <div style={{ fontSize: '0.74rem', color: 'var(--text-3)', marginTop: 2 }}>SKU: {item.product_sku} · {item.unit}</div>
                </div>
                <div style={{ textAlign: 'right' }}>
                  <div style={{ fontWeight: 700, color: 'var(--accent-em)' }}>{formatRp(item.subtotal)}</div>
                  <div style={{ fontSize: '0.73rem', color: 'var(--text-3)' }}>{item.quantity} × {formatRp(item.unit_cost)}</div>
                </div>
              </div>
            ))}
          </div>

          {/* Financials */}
          <div style={{ background: 'var(--bg-elevated)', borderRadius: 12, padding: '14px 16px', display: 'flex', flexDirection: 'column', gap: 8 }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.85rem' }}>
              <span style={{ color: 'var(--text-2)' }}>Total PO</span>
              <span style={{ fontWeight: 700 }}>{formatRp(po.total_amount)}</span>
            </div>
            <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.85rem' }}>
              <span style={{ color: '#10b981' }}>Sudah Dibayar</span>
              <span style={{ fontWeight: 700, color: '#10b981' }}>{formatRp(amountPaid)}</span>
            </div>
            <div style={{ borderTop: '1px solid var(--border)', paddingTop: 8, display: 'flex', justifyContent: 'space-between', fontSize: '0.9rem' }}>
              <span style={{ fontWeight: 700 }}>Sisa Hutang</span>
              <span style={{ fontWeight: 800, color: amountDue > 0 ? '#ef4444' : '#10b981' }}>{formatRp(amountDue)}</span>
            </div>
          </div>

          {/* Payment history */}
          {po.status === 'received' && (
            <div>
              <div style={{ fontWeight: 700, marginBottom: 8, fontSize: '0.9rem' }}>Riwayat Pembayaran</div>
              {payments.length === 0
                ? <div style={{ fontSize: '0.82rem', color: 'var(--text-3)', padding: '12px 0', textAlign: 'center' }}>Belum ada pembayaran</div>
                : payments.map(p => (
                  <div key={p.id} style={{ background: 'var(--bg-elevated)', borderRadius: 10, padding: '10px 14px', display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 6 }}>
                    <div>
                      <div style={{ fontWeight: 700, color: '#10b981', fontSize: '0.9rem' }}>{formatRp(p.amount)}</div>
                      <div style={{ fontSize: '0.73rem', color: 'var(--text-3)', marginTop: 2 }}>
                        {p.paid_by_name} · {formatDate(p.paid_at)}
                        {p.note && ` · ${p.note}`}
                      </div>
                    </div>
                    <CheckCircle2 size={16} style={{ color: '#10b981', flexShrink: 0 }} />
                  </div>
                ))
              }
            </div>
          )}

          {/* Action buttons */}
          <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
            {po.status === 'received' && payStatus !== 'paid' && (
              <button className="btn btn-primary" style={{ width: '100%' }} onClick={onPay}>
                <Wallet size={14} /> Bayar Hutang {amountDue > 0 ? `(${formatRp(amountDue)})` : ''}
              </button>
            )}
            {(po.status === 'draft' || po.status === 'ordered') && (
              <div style={{ display: 'flex', gap: 8 }}>
                {po.status === 'draft'   && <button className="btn btn-primary" style={{ flex: 1 }} onClick={() => onAction('submit')}><CheckCircle2 size={14} /> Submit PO</button>}
                {po.status === 'ordered' && <button className="btn btn-primary" style={{ flex: 1 }} onClick={() => onAction('receive')}><Package size={14} /> Terima PO</button>}
                <button className="btn btn-danger" onClick={() => onAction('cancel')}><X size={14} /> Batalkan</button>
              </div>
            )}
          </div>
        </div>
      </div>
    </>
  );
}

// ── Main Page ─────────────────────────────────────────────────────────────────
export default function PurchaseOrdersPage() {
  const { selectedStore } = useAuth();
  const [orders, setOrders]           = useState<PurchaseOrder[]>([]);
  const [loading, setLoading]         = useState(true);
  const [payable, setPayable]         = useState<PayableSummary | null>(null);
  const [showModal, setShowModal]     = useState(false);
  const [detailPO, setDetailPO]       = useState<PurchaseOrder | null>(null);
  const [invoicePO, setInvoicePO]     = useState<PurchaseOrder | null>(null);
  const [payingPO, setPayingPO]       = useState<PurchaseOrder | null>(null);
  const [storeDetail, setStoreDetail] = useState<Store | null>(null);
  const [confirm, setConfirm]         = useState<{ po: PurchaseOrder; action: ActionType } | null>(null);
  const [confirmLoading, setConfirmLoading] = useState(false);
  const [payments, setPayments]       = useState<POPayment[]>([]);
  const [suppliers, setSuppliers]     = useState<Supplier[]>([]);
  const [products, setProducts]       = useState<Product[]>([]);
  const [form, setForm]               = useState(EMPTY_FORM);
  const [saving, setSaving]           = useState(false);
  const [error, setError]             = useState('');

  const storeId = selectedStore?.store_id;

  const load = useCallback(() => {
    if (!storeId) return;
    setLoading(true);
    Promise.all([
      purchaseOrdersApi.list(storeId, { per_page: 50 }),
      purchaseOrdersApi.payableSummary(storeId),
    ]).then(([posRes, payRes]) => {
      setOrders((posRes.data as any).data ?? []);
      setPayable(payRes.data as PayableSummary);
    }).catch(console.error).finally(() => setLoading(false));
  }, [storeId]);

  useEffect(() => { load(); }, [load]);

  useEffect(() => {
    if (!storeId) return;
    Promise.all([
      suppliersApi.list({ per_page: 100 }),
      productsApi.list(storeId, { per_page: 200 }),
      storesApi.get(storeId),
    ]).then(([s, p, st]) => {
      setSuppliers((s.data as any).data ?? []);
      setProducts((p.data as any).data ?? []);
      setStoreDetail(st.data as Store);
    });
  }, [storeId]);

  const openDetail = useCallback(async (po: PurchaseOrder) => {
    setDetailPO(po); setPayments([]);
    try {
      const [r, pmts] = await Promise.all([
        purchaseOrdersApi.get(storeId!, po.id),
        po.status === 'received' ? purchaseOrdersApi.listPayments(storeId!, po.id) : Promise.resolve({ data: [] }),
      ]);
      setDetailPO(r.data as PurchaseOrder);
      setPayments((pmts.data as any) ?? []);
    } catch (e) { console.error(e); }
  }, [storeId]);

  const openInvoice = useCallback(async (po: PurchaseOrder) => {
    let fullPO = po;
    if (!po.items?.length) {
      try { const r = await purchaseOrdersApi.get(storeId!, po.id); fullPO = r.data as PurchaseOrder; }
      catch (e) { console.error(e); }
    }
    setInvoicePO(fullPO);
  }, [storeId]);

  const executeAction = async () => {
    if (!confirm || !storeId) return;
    setConfirmLoading(true);
    const fns: Record<ActionType, () => Promise<any>> = {
      submit:  () => purchaseOrdersApi.submit(storeId, confirm.po.id),
      receive: () => purchaseOrdersApi.receive(storeId, confirm.po.id),
      cancel:  () => purchaseOrdersApi.cancel(storeId, confirm.po.id),
    };
    try {
      await fns[confirm.action]();
      setConfirm(null);
      load();
      if (detailPO?.id === confirm.po.id) openDetail({ ...detailPO });
    } catch (e) {
      alert(e instanceof ApiError ? e.message : 'Gagal');
    } finally { setConfirmLoading(false); }
  };

  const handlePaySuccess = () => {
    setPayingPO(null);
    load();
    if (detailPO) openDetail(detailPO);
  };

  const addItem    = () => setForm(f => ({ ...f, items: [...f.items, { product_id: '', quantity: 1, unit_cost: 0 }] }));
  const removeItem = (i: number) => setForm(f => ({ ...f, items: f.items.filter((_, idx) => idx !== i) }));
  const updateItem = (i: number, k: keyof ItemRow, v: string) => {
    setForm(f => ({
      ...f,
      items: f.items.map((item, idx) => {
        if (idx !== i) return item;
        if (k === 'product_id') {
          const p = products.find(p => p.id === v);
          return { ...item, product_id: v, unit_cost: p?.cost_price ?? item.unit_cost };
        }
        return { ...item, [k]: +v };
      }),
    }));
  };
  const runningTotal = form.items.reduce((s, it) => s + it.quantity * it.unit_cost, 0);

  const handleCreate = async () => {
    if (!storeId) return;
    setSaving(true); setError('');
    try {
      await purchaseOrdersApi.create(storeId, { supplier_id: form.supplier_id || undefined, notes: form.notes, items: form.items.filter(i => i.product_id) });
      setShowModal(false); setForm(EMPTY_FORM); load();
    } catch (e) { setError(e instanceof ApiError ? e.message : 'Gagal membuat PO'); }
    finally { setSaving(false); }
  };

  if (!selectedStore) return (
    <div style={{ padding: 32 }}><div className="empty-state card" style={{ padding: 40 }}><ClipboardList size={40} /><p>Pilih toko terlebih dahulu</p></div></div>
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

      {/* Payable Summary Cards */}
      {payable && (
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit,minmax(180px,1fr))', gap: 12, marginBottom: 20 }}>
          {[
            { label: 'Total Hutang Dagang', val: payable.total_debt,        icon: CreditCard,    color: 'var(--accent-em)' },
            { label: 'Sudah Dibayar',       val: payable.total_paid,        icon: CheckCircle2,  color: '#10b981' },
            { label: 'Sisa Hutang',         val: payable.total_outstanding, icon: AlertCircle,   color: '#ef4444' },
          ].map(({ label, val, icon: Icon, color }) => (
            <div key={label} className="card" style={{ padding: '14px 16px' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 8 }}>
                <Icon size={16} style={{ color }} />
                <span style={{ fontSize: '0.72rem', color: 'var(--text-3)', fontWeight: 600 }}>{label}</span>
              </div>
              <div style={{ fontWeight: 800, fontSize: '1.1rem', color }}>{formatRp(val)}</div>
            </div>
          ))}
          <div className="card" style={{ padding: '14px 16px' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 8 }}>
              <Clock size={16} style={{ color: '#f59e0b' }} />
              <span style={{ fontSize: '0.72rem', color: 'var(--text-3)', fontWeight: 600 }}>Status Hutang</span>
            </div>
            <div style={{ display: 'flex', gap: 8, fontSize: '0.82rem' }}>
              <span style={{ color: '#ef4444', fontWeight: 700 }}>{payable.unpaid_count} belum bayar</span>
              <span style={{ color: 'var(--text-3)' }}>·</span>
              <span style={{ color: '#f59e0b', fontWeight: 700 }}>{payable.partial_count} sebagian</span>
            </div>
          </div>
        </div>
      )}

      {/* List */}
      <div className="card" style={{ overflow: 'hidden' }}>
        {loading ? (
          <div style={{ display: 'flex', justifyContent: 'center', padding: 40 }}><Loader2 size={24} className="loading-spin" style={{ color: 'var(--accent-em)' }} /></div>
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
                <th>Hutang</th>
                <th>Dibuat</th>
                <th>Aksi</th>
              </tr>
            </thead>
            <tbody>
              {orders.map(po => {
                const ps = (po as any).payment_status ?? 'unpaid';
                return (
                  <tr key={po.id} style={{ cursor: 'pointer' }}>
                    <td onClick={() => openDetail(po)} style={{ fontFamily: 'monospace', fontWeight: 700, color: 'var(--accent-em)' }}>{po.po_number}</td>
                    <td onClick={() => openDetail(po)} style={{ color: 'var(--text-2)' }}>{(po as any).supplier_name ?? '—'}</td>
                    <td onClick={() => openDetail(po)}>
                      <span style={{ background: 'var(--bg-elevated)', borderRadius: 6, padding: '2px 8px', fontSize: '0.78rem', color: 'var(--text-2)' }}>{po.items?.length ?? 0} item</span>
                    </td>
                    <td onClick={() => openDetail(po)} style={{ fontWeight: 700, color: 'var(--accent-em)' }}>{formatRp(po.total_amount)}</td>
                    <td onClick={() => openDetail(po)}><span className={`badge ${STATUS_BADGE[po.status]}`}>{STATUS_LABEL[po.status]}</span></td>
                    <td onClick={() => openDetail(po)}>
                      {po.status === 'received' ? <PayStatusBadge status={ps} /> : <span style={{ color: 'var(--text-3)', fontSize: '0.78rem' }}>—</span>}
                    </td>
                    <td onClick={() => openDetail(po)} style={{ color: 'var(--text-3)', fontSize: '0.8rem' }}>{formatDate(po.created_at)}</td>
                    <td>
                      <div style={{ display: 'flex', gap: 4 }}>
                        {po.status === 'draft'    && <button className="btn btn-secondary btn-sm" onClick={e => { e.stopPropagation(); setConfirm({ po, action: 'submit' }); }}>Submit</button>}
                        {po.status === 'ordered'  && <button className="btn btn-primary btn-sm"   onClick={e => { e.stopPropagation(); setConfirm({ po, action: 'receive' }); }}>Terima</button>}
                        {po.status === 'received' && ps !== 'paid' && (
                          <button className="btn btn-secondary btn-sm" style={{ color: '#10b981' }} onClick={e => { e.stopPropagation(); setPayingPO(po); }}>
                            <Wallet size={12} /> Bayar
                          </button>
                        )}
                        {(po.status === 'draft' || po.status === 'ordered') && <button className="btn btn-danger btn-sm" onClick={e => { e.stopPropagation(); setConfirm({ po, action: 'cancel' }); }}>Batal</button>}
                        <button className="btn btn-ghost btn-sm" onClick={e => { e.stopPropagation(); openInvoice(po); }}><Printer size={13} /></button>
                        <button className="btn btn-ghost btn-sm" onClick={() => openDetail(po)}><ChevronRight size={14} /></button>
                      </div>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        )}
      </div>

      {/* Overlays */}
      {detailPO && <PODetailDrawer po={detailPO} storeId={storeId!} payments={payments} onClose={() => setDetailPO(null)} onInvoice={() => openInvoice(detailPO)} onAction={a => setConfirm({ po: detailPO, action: a })} onPay={() => setPayingPO(detailPO)} />}
      {invoicePO && <InvoiceModal po={invoicePO} store={storeDetail} onClose={() => setInvoicePO(null)} />}
      {payingPO  && <PayModal po={payingPO} storeId={storeId!} onSuccess={handlePaySuccess} onCancel={() => setPayingPO(null)} />}
      {confirm   && <ConfirmModal action={confirm.action} po={confirm.po} onConfirm={executeAction} onCancel={() => setConfirm(null)} loading={confirmLoading} />}

      {/* Create Modal */}
      {showModal && (
        <div className="modal-overlay" onClick={() => setShowModal(false)}>
          <div className="modal-box" style={{ maxWidth: 580, maxHeight: '90vh', overflowY: 'auto' }} onClick={e => e.stopPropagation()}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 18 }}>
              <h2 style={{ fontWeight: 800 }}>Buat Purchase Order</h2>
              <button className="btn btn-ghost btn-sm" onClick={() => setShowModal(false)}><X size={15} /></button>
            </div>
            {error && <div style={{ background: 'rgba(239,68,68,0.12)', borderRadius: 8, padding: '8px 12px', color: '#f87171', fontSize: '0.83rem', marginBottom: 14, border: '1px solid rgba(239,68,68,0.3)' }}>{error}</div>}
            <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
              <div className="input-group">
                <label className="input-label">Supplier (opsional)</label>
                <select className="input" value={form.supplier_id} onChange={e => setForm(f => ({ ...f, supplier_id: e.target.value }))}>
                  <option value="">Tanpa Supplier</option>
                  {suppliers.map(s => <option key={s.id} value={s.id}>{s.name}</option>)}
                </select>
              </div>
              <div className="input-group">
                <label className="input-label">Catatan (opsional)</label>
                <input className="input" value={form.notes} onChange={e => setForm(f => ({ ...f, notes: e.target.value }))} placeholder="Catatan pemesanan..." />
              </div>
              <div>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 10 }}>
                  <label className="input-label" style={{ margin: 0 }}>Item Pembelian</label>
                  <button className="btn btn-ghost btn-sm" onClick={addItem}><Plus size={13} /> Tambah Item</button>
                </div>
                <div style={{ display: 'grid', gridTemplateColumns: '1fr 90px 130px auto', gap: 6, padding: '0 4px 6px' }}>
                  {['PRODUK','QTY','HARGA BELI/UNIT',''].map(h => <span key={h} style={{ fontSize: '0.7rem', color: 'var(--text-3)', fontWeight: 700 }}>{h}</span>)}
                </div>
                {form.items.map((item, i) => {
                  const sp = products.find(p => p.id === item.product_id);
                  const lineTotal = item.quantity * item.unit_cost;
                  return (
                    <div key={i} style={{ marginBottom: 8 }}>
                      <div style={{ display: 'grid', gridTemplateColumns: '1fr 90px 130px auto', gap: 6 }}>
                        <select className="input" value={item.product_id} onChange={e => updateItem(i, 'product_id', e.target.value)}>
                          <option value="">Pilih produk...</option>
                          {products.map(p => <option key={p.id} value={p.id}>{p.name}</option>)}
                        </select>
                        <input type="number" className="input" placeholder="Qty" min={1} value={item.quantity} onChange={e => updateItem(i, 'quantity', e.target.value)} />
                        <input type="number" className="input" placeholder="Harga beli" min={0} value={item.unit_cost} onChange={e => updateItem(i, 'unit_cost', e.target.value)} />
                        <button className="btn btn-ghost btn-sm" style={{ color: 'var(--accent-rd)', alignSelf: 'center' }} onClick={() => removeItem(i)}><X size={13} /></button>
                      </div>
                      {sp && <div style={{ display: 'flex', justifyContent: 'space-between', padding: '4px 4px 0', fontSize: '0.72rem', color: 'var(--text-3)' }}>
                        <span>SKU: {sp.sku} · Stok: {sp.stock_qty ?? '?'} {sp.unit} · HPP: {formatRp(sp.cost_price)}</span>
                        {lineTotal > 0 && <span style={{ color: 'var(--accent-em)', fontWeight: 600 }}>= {formatRp(lineTotal)}</span>}
                      </div>}
                    </div>
                  );
                })}
                {runningTotal > 0 && <div style={{ marginTop: 10, padding: '10px 14px', background: 'rgba(16,185,129,0.08)', border: '1px solid rgba(16,185,129,0.25)', borderRadius: 10, display: 'flex', justifyContent: 'space-between' }}>
                  <span style={{ fontSize: '0.85rem', color: 'var(--text-2)' }}>Total Estimasi</span>
                  <span style={{ fontWeight: 800, color: '#10b981' }}>{formatRp(runningTotal)}</span>
                </div>}
              </div>
            </div>
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
