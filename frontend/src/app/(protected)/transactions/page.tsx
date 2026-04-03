'use client';

import { useEffect, useState, useCallback } from 'react';
import {
  Search, X, Printer, ChevronLeft, ChevronRight,
  Receipt, CheckCircle2, XCircle, Clock, CalendarDays,
  User, CreditCard, Loader2, Eye,
} from 'lucide-react';
import { useAuth } from '@/lib/auth/AuthContext';
import { transactionsApi } from '@/lib/api/transactions';
import { formatRp } from '@/lib/utils';
import type { Transaction } from '@/types';
import { ApiError } from '@/lib/api/client';

// ── Date helpers ──────────────────────────────────────────────────────────────
function today() { return new Date().toISOString().slice(0, 10); }
function daysAgo(n: number) {
  const d = new Date(); d.setDate(d.getDate() - n);
  return d.toISOString().slice(0, 10);
}
function startOfWeek() {
  const d = new Date();
  const day = d.getDay(); // 0=Sun
  d.setDate(d.getDate() - (day === 0 ? 6 : day - 1)); // Mon
  return d.toISOString().slice(0, 10);
}
function startOfMonth() {
  const d = new Date(); d.setDate(1);
  return d.toISOString().slice(0, 10);
}

type Preset = 'today' | 'week' | 'month' | 'custom';

const PRESETS: { key: Preset; label: string }[] = [
  { key: 'today',  label: 'Hari Ini' },
  { key: 'week',   label: 'Minggu Ini' },
  { key: 'month',  label: 'Bulan Ini' },
  { key: 'custom', label: 'Kustom' },
];

function presetRange(p: Preset): [string, string] {
  switch (p) {
    case 'today':  return [today(), today()];
    case 'week':   return [startOfWeek(), today()];
    case 'month':  return [startOfMonth(), today()];
    default:       return [daysAgo(30), today()];
  }
}

// ── Status badge ─────────────────────────────────────────────────────────────
function StatusBadge({ status }: { status: string }) {
  const cfg: Record<string, { color: string; icon: React.ReactNode }> = {
    completed: { color: 'var(--accent-em)',  icon: <CheckCircle2 size={11} /> },
    voided:    { color: 'var(--accent-rd)',  icon: <XCircle size={11} /> },
    draft:     { color: 'var(--accent-am)',  icon: <Clock size={11} /> },
  };
  const { color, icon } = cfg[status] ?? { color: 'var(--text-3)', icon: null };
  return (
    <span style={{
      display: 'inline-flex', alignItems: 'center', gap: 3,
      padding: '2px 8px', borderRadius: 6, fontSize: '0.72rem', fontWeight: 600,
      background: `${color}20`, color,
    }}>
      {icon} {status === 'completed' ? 'Selesai' : status === 'voided' ? 'Dibatalkan' : 'Draft'}
    </span>
  );
}

// ── Payment method badge ──────────────────────────────────────────────────────
function PayBadge({ method }: { method: string }) {
  const colors: Record<string, string> = {
    cash: '#10b981', qris: '#8b5cf6', card: '#3b82f6', transfer: '#f59e0b',
  };
  return (
    <span style={{
      display: 'inline-flex', alignItems: 'center', gap: 3,
      padding: '2px 8px', borderRadius: 6, fontSize: '0.72rem', fontWeight: 700,
      background: `${colors[method] ?? '#6b7280'}20`, color: colors[method] ?? '#6b7280',
    }}>
      {method.toUpperCase()}
    </span>
  );
}

// ── Receipt Printout ─────────────────────────────────────────────────────────
function ReceiptModal({ txn, onClose }: { txn: Transaction; onClose: () => void }) {
  const handlePrint = () => {
    // Build self-contained receipt HTML with print-safe (hardcoded) styles.
    // We cannot rely on CSS custom properties (--var) in a print context because
    // those resolve to transparent/nothing on most browsers.
    const lines = (txn.items ?? []).map(item => `
      <div class="item">
        <div class="item-name">${item.product_name}</div>
        <div class="item-row">
          <span>${item.quantity} × ${formatRp(item.unit_price)}</span>
          <span>${formatRp(item.subtotal)}</span>
        </div>
      </div>`).join('');

    const receiptHtml = `<!DOCTYPE html>
<html lang="id">
<head>
  <meta charset="UTF-8" />
  <title>Struk - ${txn.id.slice(0,8).toUpperCase()}</title>
  <style>
    * { margin: 0; padding: 0; box-sizing: border-box; }
    body {
      font-family: 'Courier New', Courier, monospace;
      font-size: 12px;
      color: #000;
      background: #fff;
      width: 80mm;
      padding: 8px;
    }
    .center { text-align: center; }
    .bold { font-weight: 700; }
    .divider { border-top: 1px dashed #999; margin: 8px 0; }
    .row { display: flex; justify-content: space-between; margin: 2px 0; }
    .total-row { display: flex; justify-content: space-between; font-weight: 800; font-size: 14px; margin-top: 4px; }
    .item { margin-bottom: 6px; }
    .item-name { font-weight: 600; }
    .item-row { display: flex; justify-content: space-between; font-size: 11px; color: #444; }
    .muted { color: #555; font-size: 11px; }
    .footer { text-align: center; margin-top: 10px; font-size: 11px; color: #555; }
    @page { size: 80mm auto; margin: 0; }
  </style>
</head>
<body>
  <div class="center bold" style="font-size:14px;margin-bottom:4px;">MoedahPOS</div>
  <div class="center muted">${new Date(txn.created_at).toLocaleString('id-ID', { dateStyle: 'long', timeStyle: 'short' })}</div>
  <div class="center muted">Kasir: ${txn.cashier_name ?? '-'}</div>
  ${txn.customer_name ? `<div class="center muted">Pelanggan: ${txn.customer_name}</div>` : ''}
  <div class="divider"></div>
  ${lines}
  <div class="divider"></div>
  <div class="row muted"><span>Subtotal</span><span>${formatRp(txn.subtotal)}</span></div>
  ${txn.discount_amt > 0 ? `<div class="row muted"><span>Diskon</span><span>-${formatRp(txn.discount_amt)}</span></div>` : ''}
  <div class="row muted"><span>PPN</span><span>${formatRp(txn.tax_amt)}</span></div>
  <div class="total-row"><span>TOTAL</span><span>${formatRp(txn.total)}</span></div>
  <div class="row muted" style="margin-top:4px;"><span>Bayar (${(txn.payment_method ?? '').toUpperCase()})</span><span>${formatRp(txn.payment_amount)}</span></div>
  ${txn.change_amount > 0 ? `<div class="row muted"><span>Kembalian</span><span>${formatRp(txn.change_amount)}</span></div>` : ''}
  <div class="footer">
    <div class="divider"></div>
    Terima kasih telah berbelanja!<br/>
    No. Transaksi: ${txn.id.slice(0,8).toUpperCase()}
  </div>
</body>
</html>`;

    const pw = window.open('', '_blank', 'width=360,height=600');
    if (!pw) return;
    pw.document.write(receiptHtml);
    pw.document.close();
    // Wait for fonts/images to load then trigger print
    pw.onload = () => {
      pw.focus();
      pw.print();
      // Close the helper window after print dialog closes
      pw.onafterprint = () => pw.close();
    };
    // Fallback: some browsers fire load synchronously
    setTimeout(() => {
      try { pw.focus(); pw.print(); } catch { /* already printed */ }
    }, 500);
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-box" style={{ maxWidth: 350 }} onClick={e => e.stopPropagation()}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
          <h2 style={{ fontWeight: 700, fontSize: '1rem' }}>Cetak Ulang Struk</h2>
          <button onClick={onClose} className="btn btn-ghost btn-sm"><X size={16} /></button>
        </div>

        {/* Preview — uses CSS vars for screen display */}
        <div id="receipt-content" style={{ fontFamily: "'Courier New', monospace", fontSize: '0.82rem', lineHeight: 1.7 }}>
          <div style={{ textAlign: 'center', marginBottom: 12, borderBottom: '1px dashed var(--border-md)', paddingBottom: 10 }}>
            <div style={{ fontWeight: 800, fontSize: '1rem' }}>MoedahPOS</div>
            <div style={{ color: 'var(--text-2)', fontSize: '0.75rem' }}>
              {new Date(txn.created_at).toLocaleString('id-ID', { dateStyle: 'full', timeStyle: 'short' })}
            </div>
            <div style={{ color: 'var(--text-2)', fontSize: '0.75rem' }}>Kasir: {txn.cashier_name}</div>
            {txn.customer_name && <div style={{ color: 'var(--text-2)', fontSize: '0.75rem' }}>Pelanggan: {txn.customer_name}</div>}
          </div>

          {txn.items?.map((item, i) => (
            <div key={i} style={{ marginBottom: 6 }}>
              <div style={{ fontWeight: 600 }}>{item.product_name}</div>
              <div style={{ display: 'flex', justifyContent: 'space-between', color: 'var(--text-2)', fontSize: '0.78rem' }}>
                <span>{item.quantity} × {formatRp(item.unit_price)}</span>
                <span>{formatRp(item.subtotal)}</span>
              </div>
            </div>
          ))}

          <div style={{ borderTop: '1px dashed var(--border-md)', marginTop: 10, paddingTop: 10 }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', color: 'var(--text-2)', fontSize: '0.8rem' }}>
              <span>Subtotal</span><span>{formatRp(txn.subtotal)}</span>
            </div>
            {txn.discount_amt > 0 && (
              <div style={{ display: 'flex', justifyContent: 'space-between', color: 'var(--accent-rd)', fontSize: '0.8rem' }}>
                <span>Diskon</span><span>-{formatRp(txn.discount_amt)}</span>
              </div>
            )}
            <div style={{ display: 'flex', justifyContent: 'space-between', color: 'var(--text-2)', fontSize: '0.8rem' }}>
              <span>PPN</span><span>{formatRp(txn.tax_amt)}</span>
            </div>
            <div style={{ display: 'flex', justifyContent: 'space-between', fontWeight: 800, fontSize: '1rem', marginTop: 4, color: 'var(--accent-em)' }}>
              <span>TOTAL</span><span>{formatRp(txn.total)}</span>
            </div>
            <div style={{ display: 'flex', justifyContent: 'space-between', color: 'var(--text-2)', fontSize: '0.8rem', marginTop: 4 }}>
              <span>Bayar ({txn.payment_method.toUpperCase()})</span><span>{formatRp(txn.payment_amount)}</span>
            </div>
            {txn.change_amount > 0 && (
              <div style={{ display: 'flex', justifyContent: 'space-between', color: 'var(--accent-em)', fontSize: '0.8rem' }}>
                <span>Kembalian</span><span>{formatRp(txn.change_amount)}</span>
              </div>
            )}
          </div>

          <div style={{ textAlign: 'center', marginTop: 12, color: 'var(--text-3)', fontSize: '0.75rem', borderTop: '1px dashed var(--border-md)', paddingTop: 10 }}>
            Terima kasih telah berbelanja!<br />
            No. Transaksi: {txn.id.slice(0, 8).toUpperCase()}
          </div>
        </div>

        <div style={{ display: 'flex', gap: 8, marginTop: 16 }}>
          <button className="btn btn-secondary" style={{ flex: 1 }} onClick={onClose}>Tutup</button>
          <button className="btn btn-primary" style={{ flex: 1 }} onClick={handlePrint}>
            <Printer size={15} /> Cetak
          </button>
        </div>
      </div>
    </div>
  );
}


// ── Detail Drawer ─────────────────────────────────────────────────────────────
function DetailDrawer({ txn, onClose, onReprint }: {
  txn: Transaction; onClose: () => void; onReprint: (t: Transaction) => void;
}) {
  return (
    <div
      style={{
        position: 'fixed', inset: 0, zIndex: 200,
        display: 'flex', justifyContent: 'flex-end',
      }}
      onClick={onClose}
    >
      {/* backdrop */}
      <div style={{ position: 'absolute', inset: 0, background: 'rgba(0,0,0,0.5)' }} />
      {/* drawer */}
      <div
        style={{
          position: 'relative', width: 420, height: '100%',
          background: 'var(--bg-card)', borderLeft: '1px solid var(--border)',
          display: 'flex', flexDirection: 'column', overflow: 'hidden',
          boxShadow: '-4px 0 24px rgba(0,0,0,0.35)',
        }}
        onClick={e => e.stopPropagation()}
      >
        {/* header */}
        <div style={{ padding: '18px 20px', borderBottom: '1px solid var(--border)', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <div>
            <div style={{ fontWeight: 700, fontSize: '1rem' }}>Detail Transaksi</div>
            <div style={{ fontSize: '0.75rem', color: 'var(--text-3)', fontFamily: 'monospace', marginTop: 2 }}>
              #{txn.id.slice(0, 8).toUpperCase()}
            </div>
          </div>
          <button className="btn btn-ghost btn-sm" onClick={onClose}><X size={16} /></button>
        </div>

        {/* body */}
        <div style={{ flex: 1, overflowY: 'auto', padding: '20px' }}>

          {/* cashier shift card */}
          <div style={{
            display: 'flex', alignItems: 'center', gap: 12,
            background: 'var(--bg-elevated)', borderRadius: 10, padding: '12px 14px', marginBottom: 16,
            border: '1px solid var(--border-md)',
          }}>
            <div style={{ width: 38, height: 38, borderRadius: 10, background: 'rgba(99,102,241,0.15)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
              <User size={18} style={{ color: '#818cf8' }} />
            </div>
            <div>
              <div style={{ fontWeight: 700, fontSize: '0.92rem' }}>{txn.cashier_name}</div>
              <div style={{ fontSize: '0.75rem', color: 'var(--text-3)' }}>
                {new Date(txn.created_at).toLocaleString('id-ID', { dateStyle: 'long', timeStyle: 'short' })}
              </div>
            </div>
            <div style={{ marginLeft: 'auto' }}>
              <StatusBadge status={txn.status} />
            </div>
          </div>

          {/* payment summary */}
          <div style={{
            background: 'rgba(16,185,129,0.07)', borderRadius: 10, padding: '14px',
            border: '1px solid rgba(16,185,129,0.15)', marginBottom: 16,
          }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 6, marginBottom: 10 }}>
              <CreditCard size={14} style={{ color: 'var(--accent-em)' }} />
              <span style={{ fontWeight: 700, fontSize: '0.85rem' }}>Pembayaran</span>
              <span style={{ marginLeft: 'auto' }}><PayBadge method={txn.payment_method} /></span>
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 5 }}>
              {[
                { label: 'Subtotal', val: formatRp(txn.subtotal) },
                ...(txn.discount_amt > 0 ? [{ label: 'Diskon', val: `-${formatRp(txn.discount_amt)}`, red: true }] : []),
                { label: 'PPN', val: formatRp(txn.tax_amt) },
              ].map(({ label, val, red }) => (
                <div key={label} style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.82rem', color: red ? 'var(--accent-rd)' : 'var(--text-2)' }}>
                  <span>{label}</span><span>{val}</span>
                </div>
              ))}
              <div style={{ display: 'flex', justifyContent: 'space-between', fontWeight: 800, fontSize: '1rem', color: 'var(--accent-em)', paddingTop: 6, borderTop: '1px dashed rgba(16,185,129,0.3)', marginTop: 4 }}>
                <span>TOTAL</span><span>{formatRp(txn.total)}</span>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.82rem', color: 'var(--text-2)' }}>
                <span>Dibayar</span><span>{formatRp(txn.payment_amount)}</span>
              </div>
              {txn.change_amount > 0 && (
                <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.82rem', color: 'var(--accent-em)', fontWeight: 600 }}>
                  <span>Kembalian</span><span>{formatRp(txn.change_amount)}</span>
                </div>
              )}
            </div>
          </div>

          {/* customer info */}
          <div style={{
            marginBottom: 16, padding: '12px 14px',
            background: txn.customer_name ? 'rgba(16,185,129,0.07)' : 'var(--bg-elevated)',
            borderRadius: 10,
            border: txn.customer_name ? '1px solid rgba(16,185,129,0.25)' : '1px solid var(--border-md)',
          }}>
            <div style={{ fontSize: '0.68rem', color: txn.customer_name ? '#10b981' : 'var(--text-3)', marginBottom: 6, textTransform: 'uppercase', letterSpacing: '0.08em', fontWeight: 700, display: 'flex', alignItems: 'center', gap: 4 }}>
              <User size={11} /> Pelanggan
            </div>
            {txn.customer_name ? (
              <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                <div style={{ width: 34, height: 34, borderRadius: '50%', background: 'linear-gradient(135deg, #10b981, #059669)', display: 'flex', alignItems: 'center', justifyContent: 'center', fontWeight: 800, fontSize: '0.85rem', color: '#fff', flexShrink: 0 }}>
                  {txn.customer_name.charAt(0).toUpperCase()}
                </div>
                <div>
                  <div style={{ fontWeight: 700, fontSize: '0.92rem', color: '#10b981' }}>{txn.customer_name}</div>
                  {txn.customer_phone && <div style={{ fontSize: '0.78rem', color: 'var(--text-2)' }}>📞 {txn.customer_phone}</div>}
                </div>
              </div>
            ) : (
              <div style={{ fontSize: '0.82rem', color: 'var(--text-3)', fontStyle: 'italic' }}>Tidak ada data customer</div>
            )}
          </div>

          {/* items */}
          <div style={{ fontSize: '0.75rem', color: 'var(--text-3)', marginBottom: 8, textTransform: 'uppercase', letterSpacing: '0.05em' }}>Item ({txn.items?.length ?? 0})</div>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
            {(txn.items ?? []).map((item, i) => (
              <div key={i} style={{
                display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start',
                padding: '10px 12px', background: 'var(--bg-elevated)', borderRadius: 8,
                border: '1px solid var(--border)',
              }}>
                <div style={{ flex: 1 }}>
                  <div style={{ fontWeight: 600, fontSize: '0.88rem' }}>{item.product_name}</div>
                  <div style={{ fontSize: '0.75rem', color: 'var(--text-3)', marginTop: 1 }}>
                    {formatRp(item.unit_price)} × {item.quantity}
                    {item.tax_rate > 0 && ` · PPN ${item.tax_rate}%`}
                  </div>
                </div>
                <div style={{ fontWeight: 700, fontSize: '0.88rem', color: 'var(--accent-em)', marginLeft: 12, whiteSpace: 'nowrap' }}>
                  {formatRp(item.subtotal)}
                </div>
              </div>
            ))}
          </div>

          {txn.notes && (
            <div style={{ marginTop: 12, padding: '10px 12px', background: 'var(--bg-elevated)', borderRadius: 8, border: '1px solid var(--border)', fontSize: '0.82rem', color: 'var(--text-2)' }}>
              📝 {txn.notes}
            </div>
          )}
        </div>

        {/* footer */}
        <div style={{ padding: '14px 20px', borderTop: '1px solid var(--border)', display: 'flex', gap: 8 }}>
          <button className="btn btn-secondary" style={{ flex: 1 }} onClick={onClose}>
            Tutup
          </button>
          <button
            className="btn btn-primary" style={{ flex: 1 }}
            onClick={() => onReprint(txn)}
            disabled={txn.status !== 'completed'}
          >
            <Printer size={14} /> Cetak Ulang
          </button>
        </div>
      </div>
    </div>
  );
}

// ── Main Page ─────────────────────────────────────────────────────────────────
export default function TransactionsPage() {
  const { selectedStore } = useAuth();
  const storeId = selectedStore?.store_id;

  // Filters
  const [preset, setPreset] = useState<Preset>('today');
  const [dateFrom, setDateFrom] = useState(() => today());
  const [dateTo, setDateTo] = useState(() => today());
  const [search, setSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState('');

  // Data
  const [txns, setTxns] = useState<Transaction[]>([]);
  const [meta, setMeta] = useState({ page: 1, per_page: 20, total: 0, total_pages: 1 });
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  // UI
  const [detail, setDetail] = useState<Transaction | null>(null);
  const [detailLoading, setDetailLoading] = useState(false);
  const [reprint, setReprint] = useState<Transaction | null>(null);

  // Summaries for the header
  const totalRevenue = txns.filter(t => t.status === 'completed').reduce((s, t) => s + t.total, 0);
  const completedCount = txns.filter(t => t.status === 'completed').length;

  const load = useCallback(async (p = 1) => {
    if (!storeId) return;
    setLoading(true);
    setError('');
    try {
      const res = await transactionsApi.list(storeId, {
        page: p,
        per_page: 20,
        date_from: dateFrom,
        date_to: dateTo,
        status: statusFilter || undefined,
      });
      const body = res.data as any;
      setTxns(body.data ?? []);
      setMeta(body.meta ?? { page: p, per_page: 20, total: 0, total_pages: 1 });
      setPage(p);
    } catch (err) {
      if (err instanceof ApiError) setError(err.message);
      else setError('Gagal memuat data transaksi');
    } finally {
      setLoading(false);
    }
  }, [storeId, dateFrom, dateTo, statusFilter]);

  useEffect(() => { load(1); }, [storeId]);

  // Apply preset → update dates
  const applyPreset = (p: Preset) => {
    setPreset(p);
    if (p !== 'custom') {
      const [from, to] = presetRange(p);
      setDateFrom(from);
      setDateTo(to);
    }
  };

  const openDetail = async (txn: Transaction) => {
    // Always fetch the full transaction to get items
    setDetailLoading(true);
    try {
      const res = await transactionsApi.get(storeId!, txn.id);
      setDetail(res.data as Transaction);
    } catch {
      setDetail(txn); // fallback to list data
    } finally {
      setDetailLoading(false);
    }
  };

  // Filtered by search (client-side)
  const filtered = txns.filter(t => {
    if (!search) return true;
    const q = search.toLowerCase();
    return (
      t.id.toLowerCase().includes(q) ||
      t.cashier_name?.toLowerCase().includes(q) ||
      t.customer_name?.toLowerCase().includes(q)
    );
  });

  if (!selectedStore) {
    return (
      <div style={{ padding: 32 }}>
        <div className="empty-state card" style={{ padding: 48 }}>
          <Receipt size={40} />
          <p>Pilih toko untuk melihat riwayat transaksi</p>
        </div>
      </div>
    );
  }

  return (
    <div style={{ padding: 24 }}>
      {/* ── Header ── */}
      <div style={{ marginBottom: 20 }}>
        <h1 className="page-title">Riwayat Transaksi</h1>
        <p className="page-subtitle">{selectedStore.store_name}</p>
      </div>

      {/* ── Preset Tabs ── */}
      <div style={{ display: 'flex', gap: 4, marginBottom: 14, background: 'var(--bg-card)', borderRadius: 10, padding: 4, width: 'fit-content', border: '1px solid var(--border)' }}>
        {PRESETS.map(p => (
          <button key={p.key}
            onClick={() => applyPreset(p.key)}
            style={{
              padding: '6px 14px', borderRadius: 8, border: 'none', cursor: 'pointer',
              fontSize: '0.82rem', fontWeight: 500, transition: 'all 0.15s',
              background: preset === p.key ? 'var(--accent-in)' : 'transparent',
              color: preset === p.key ? '#fff' : 'var(--text-2)',
            }}
          >
            {p.label}
          </button>
        ))}
      </div>

      {/* ── Filter Row ── */}
      <div style={{ display: 'flex', gap: 8, marginBottom: 16, flexWrap: 'wrap', alignItems: 'center' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
          <CalendarDays size={14} style={{ color: 'var(--text-3)' }} />
          <input
            type="date" className="input" style={{ width: 145 }}
            value={dateFrom}
            onChange={e => { setDateFrom(e.target.value); setPreset('custom'); }}
          />
          <span style={{ color: 'var(--text-3)', fontSize: '0.85rem' }}>s/d</span>
          <input
            type="date" className="input" style={{ width: 145 }}
            value={dateTo}
            onChange={e => { setDateTo(e.target.value); setPreset('custom'); }}
          />
        </div>

        <select
          className="input" style={{ width: 140 }}
          value={statusFilter}
          onChange={e => setStatusFilter(e.target.value)}
        >
          <option value="">Semua Status</option>
          <option value="completed">Selesai</option>
          <option value="voided">Dibatalkan</option>
        </select>

        <div style={{ position: 'relative', flex: 1, minWidth: 200 }}>
          <Search size={14} style={{ position: 'absolute', left: 10, top: '50%', transform: 'translateY(-50%)', color: 'var(--text-3)' }} />
          <input
            className="input" style={{ paddingLeft: 32 }}
            placeholder="Cari ID, kasir, pelanggan..."
            value={search} onChange={e => setSearch(e.target.value)}
          />
        </div>

        <button className="btn btn-primary" onClick={() => load(1)} disabled={loading}>
          {loading ? <Loader2 size={14} className="loading-spin" /> : null}
          Tampilkan
        </button>
      </div>

      {/* ── Summary Cards ── */}
      {txns.length > 0 && (
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3,1fr)', gap: 12, marginBottom: 16 }}>
          {[
            { label: 'Total Transaksi', val: meta.total, color: '#6366f1' },
            { label: 'Selesai', val: completedCount, color: '#10b981' },
            { label: 'Total Penjualan', val: formatRp(totalRevenue), color: '#f59e0b', big: true },
          ].map(({ label, val, color, big }) => (
            <div key={label} className="stat-card" style={{ padding: '14px 16px' }}>
              <div style={{ fontSize: '0.75rem', color: 'var(--text-3)', marginBottom: 4 }}>{label}</div>
              <div style={{ fontWeight: 800, fontSize: big ? '1.1rem' : '1.4rem', color }}>{val}</div>
            </div>
          ))}
        </div>
      )}

      {/* ── Error ── */}
      {error && (
        <div style={{ background: 'rgba(239,68,68,0.1)', border: '1px solid rgba(239,68,68,0.3)', borderRadius: 8, padding: '10px 14px', color: '#f87171', fontSize: '0.85rem', marginBottom: 12, display: 'flex', justifyContent: 'space-between' }}>
          {error}
          <button onClick={() => setError('')} style={{ background: 'none', border: 'none', color: '#f87171', cursor: 'pointer' }}><X size={14} /></button>
        </div>
      )}

      {/* ── Table ── */}
      <div className="card" style={{ overflow: 'hidden' }}>
        {loading ? (
          <div style={{ display: 'flex', justifyContent: 'center', padding: 48 }}>
            <Loader2 size={26} className="loading-spin" style={{ color: 'var(--accent-em)' }} />
          </div>
        ) : filtered.length === 0 ? (
          <div className="empty-state" style={{ padding: 48 }}>
            <Receipt size={36} />
            <p>Tidak ada transaksi pada periode ini</p>
          </div>
        ) : (
          <table className="tbl">
            <thead>
              <tr>
                <th>ID</th>
                <th>Waktu</th>
                <th>Kasir</th>
                <th>Pelanggan</th>
                <th>Metode</th>
                <th>Total</th>
                <th>Status</th>
                <th style={{ width: 60 }}></th>
              </tr>
            </thead>
            <tbody>
              {filtered.map(txn => (
                <tr key={txn.id} style={{ cursor: 'pointer' }} onClick={() => openDetail(txn)}>
                  <td>
                    <span style={{ fontFamily: 'monospace', fontSize: '0.8rem', color: 'var(--text-2)' }}>
                      #{txn.id.slice(0, 8).toUpperCase()}
                    </span>
                  </td>
                  <td>
                    <div style={{ fontSize: '0.82rem' }}>{new Date(txn.created_at).toLocaleDateString('id-ID', { day: '2-digit', month: 'short' })}</div>
                    <div style={{ fontSize: '0.72rem', color: 'var(--text-3)' }}>{new Date(txn.created_at).toLocaleTimeString('id-ID', { timeStyle: 'short' })}</div>
                  </td>
                  <td>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
                      <div style={{
                        width: 26, height: 26, borderRadius: '50%',
                        background: 'rgba(99,102,241,0.15)', flexShrink: 0,
                        display: 'flex', alignItems: 'center', justifyContent: 'center',
                        fontSize: '0.65rem', fontWeight: 700, color: '#818cf8',
                      }}>
                        {txn.cashier_name?.split(' ').map(w => w[0]).join('').slice(0, 2).toUpperCase()}
                      </div>
                      <span style={{ fontSize: '0.85rem' }}>{txn.cashier_name}</span>
                    </div>
                  </td>
                  <td>
                    {txn.customer_name ? (
                      <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
                        <div style={{
                          width: 24, height: 24, borderRadius: '50%', flexShrink: 0,
                          background: 'linear-gradient(135deg, rgba(16,185,129,0.3), rgba(16,185,129,0.15))',
                          border: '1.5px solid rgba(16,185,129,0.4)',
                          display: 'flex', alignItems: 'center', justifyContent: 'center',
                          fontSize: '0.62rem', fontWeight: 800, color: '#10b981',
                        }}>
                          {txn.customer_name.charAt(0).toUpperCase()}
                        </div>
                        <span style={{ fontSize: '0.83rem', fontWeight: 500 }}>{txn.customer_name}</span>
                      </div>
                    ) : (
                      <span style={{ color: 'var(--text-3)', fontStyle: 'italic', fontSize: '0.8rem' }}>–</span>
                    )}
                  </td>
                  <td><PayBadge method={txn.payment_method} /></td>
                  <td style={{ fontWeight: 700, color: 'var(--accent-em)' }}>{formatRp(txn.total)}</td>
                  <td><StatusBadge status={txn.status} /></td>
                  <td onClick={e => { e.stopPropagation(); openDetail(txn); }}>
                    <button className="btn btn-ghost btn-sm" style={{ padding: '4px 6px' }}>
                      <Eye size={14} />
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      {/* ── Pagination ── */}
      {meta.total_pages > 1 && (
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginTop: 14 }}>
          <span style={{ fontSize: '0.8rem', color: 'var(--text-3)' }}>
            Hal {meta.page} dari {meta.total_pages} · {meta.total} transaksi
          </span>
          <div style={{ display: 'flex', gap: 6 }}>
            <button className="btn btn-secondary btn-sm" onClick={() => load(page - 1)} disabled={page <= 1}>
              <ChevronLeft size={14} />
            </button>
            {Array.from({ length: Math.min(meta.total_pages, 5) }, (_, i) => {
              const p = Math.max(1, page - 2) + i;
              if (p > meta.total_pages) return null;
              return (
                <button key={p} className={`btn btn-sm ${p === page ? 'btn-primary' : 'btn-secondary'}`}
                  onClick={() => load(p)}>
                  {p}
                </button>
              );
            })}
            <button className="btn btn-secondary btn-sm" onClick={() => load(page + 1)} disabled={page >= meta.total_pages}>
              <ChevronRight size={14} />
            </button>
          </div>
        </div>
      )}

      {/* ── Detail Drawer ── */}
      {detailLoading && (
        <div style={{ position: 'fixed', inset: 0, zIndex: 200, background: 'rgba(0,0,0,0.4)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
          <Loader2 size={32} className="loading-spin" style={{ color: '#fff' }} />
        </div>
      )}
      {detail && !detailLoading && (
        <DetailDrawer
          txn={detail}
          onClose={() => setDetail(null)}
          onReprint={(t) => { setDetail(null); setReprint(t); }}
        />
      )}

      {/* ── Reprint Modal ── */}
      {reprint && (
        <ReceiptModal txn={reprint} onClose={() => setReprint(null)} />
      )}
    </div>
  );
}
