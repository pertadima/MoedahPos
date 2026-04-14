'use client';

import { useEffect, useState, useCallback, Fragment, useRef } from 'react';
import {
  ClipboardList,
  Plus,
  Loader2,
  X,
  ChevronRight,
  ChevronDown,
  Package,
  User,
  Calendar,
  Hash,
  AlertTriangle,
  CheckCircle2,
  Printer,
  Wallet,
  CreditCard,
  AlertCircle,
  BadgeCheck,
  Clock,
  Eye,
  Check,
  Trash2,
} from 'lucide-react';
import { useAuth } from '@/lib/auth/AuthContext';
import { purchaseOrdersApi, suppliersApi, storesApi } from '@/lib/api/store-apis';
import { productsApi } from '@/lib/api/products';
import { formatRp, formatDate, formatNumberInput, parseNumberInput } from '@/lib/utils';
import type { PurchaseOrder, Product, Supplier, Store } from '@/types';
import { ApiError } from '@/lib/api/client';
import {
  listTermins,
  createTerminSchedule,
  recordPayment,
  type Termin,
  type RecordPaymentRequest,
} from '@/lib/api/termins';

// ── Constants ─────────────────────────────────────────────────────────────────
const STATUS_BADGE: Record<string, string> = {
  draft: 'badge-gray',
  ordered: 'badge-blue',
  received: 'badge-green',
  cancelled: 'badge-red',
};
const STATUS_LABEL: Record<string, string> = {
  draft: 'Draft',
  ordered: 'Dipesan',
  received: 'Diterima',
  cancelled: 'Dibatalkan',
};
const PAY_STATUS_CFG = {
  unpaid: { label: 'Belum Bayar', color: '#ef4444', bg: 'rgba(239,68,68,0.12)', icon: AlertCircle },
  partial: { label: 'Sebagian', color: '#f59e0b', bg: 'rgba(245,158,11,0.12)', icon: Clock },
  paid: { label: 'Lunas', color: '#10b981', bg: 'rgba(16,185,129,0.12)', icon: BadgeCheck },
};
const EMPTY_FORM: {
  supplier_id: string;
  notes: string;
  items: ItemRow[];
} = {
  supplier_id: '',
  notes: '',
  items: [{ product_id: '', quantity: 1, unit_cost: 0 }],
};
type ItemRow = {
  product_id: string;
  quantity: number;
  unit_cost: number;
  product_name?: string;
  product_sku?: string;
  unit?: string;
};

function SearchableSelect({
  options,
  value,
  onChange,
  placeholder = 'Pilih...',
  className = 'input',
}: {
  options: { value: string; label: string }[];
  value: string;
  onChange: (v: string) => void;
  placeholder?: string;
  className?: string;
}) {
  const [open, setOpen] = useState(false);
  const [search, setSearch] = useState('');
  const wrapperRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (wrapperRef.current && !wrapperRef.current.contains(e.target as Node)) {
        setOpen(false);
      }
    }
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const selectedOption = options.find(o => o.value === value);
  const filteredOptions = options.filter(o => o.label.toLowerCase().includes(search.toLowerCase()));

  return (
    <div ref={wrapperRef} style={{ position: 'relative', width: '100%' }}>
      <div
        className={className}
        style={{
          cursor: 'pointer',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          minHeight: 38,
        }}
        onClick={() => {
          setOpen(!open);
          setSearch('');
        }}
      >
        <span
          style={{
            overflow: 'hidden',
            textOverflow: 'ellipsis',
            whiteSpace: 'nowrap',
            color: selectedOption ? 'inherit' : 'var(--text-3)',
          }}
        >
          {selectedOption ? selectedOption.label : placeholder}
        </span>
        <ChevronRight size={14} style={{ transform: open ? 'rotate(90deg)' : 'rotate(0)' }} />
      </div>

      {open && (
        <div
          style={{
            position: 'absolute',
            top: '100%',
            left: 0,
            right: 0,
            zIndex: 3000,
            background: 'var(--bg-card)',
            border: '1px solid var(--border)',
            borderRadius: 8,
            marginTop: 4,
            boxShadow: '0 4px 12px rgba(0,0,0,0.15)',
            maxHeight: 280,
            display: 'flex',
            flexDirection: 'column',
          }}
        >
          <div style={{ padding: 8, borderBottom: '1px solid var(--border)' }}>
            <input
              autoFocus
              className="input"
              placeholder="Cari..."
              style={{ width: '100%', height: 32, fontSize: '0.82rem' }}
              value={search}
              onChange={e => setSearch(e.target.value)}
            />
          </div>
          <div style={{ overflowY: 'auto', flex: 1, padding: 4 }}>
            {!value ? null : (
              <div
                style={{
                  padding: '8px 12px',
                  cursor: 'pointer',
                  fontSize: '0.82rem',
                  borderRadius: 6,
                  color: 'var(--text-3)',
                }}
                onClick={() => {
                  onChange('');
                  setOpen(false);
                }}
                onMouseEnter={e => (e.currentTarget.style.background = 'var(--bg-hover)')}
                onMouseLeave={e => (e.currentTarget.style.background = 'transparent')}
              >
                Tanpa Pilihan / Reset
              </div>
            )}
            {filteredOptions.length === 0 ? (
              <div
                style={{
                  padding: '8px 12px',
                  fontSize: '0.82rem',
                  color: 'var(--text-3)',
                  textAlign: 'center',
                }}
              >
                Tidak ditemukan
              </div>
            ) : (
              filteredOptions.map(o => (
                <div
                  key={o.value}
                  style={{
                    padding: '8px 12px',
                    cursor: 'pointer',
                    fontSize: '0.82rem',
                    borderRadius: 6,
                    background: o.value === value ? 'var(--bg-active)' : 'transparent',
                  }}
                  onClick={() => {
                    onChange(o.value);
                    setOpen(false);
                  }}
                  onMouseEnter={e => {
                    if (o.value !== value) e.currentTarget.style.background = 'var(--bg-hover)';
                  }}
                  onMouseLeave={e => {
                    if (o.value !== value) e.currentTarget.style.background = 'transparent';
                  }}
                >
                  {o.label}
                </div>
              ))
            )}
          </div>
        </div>
      )}
    </div>
  );
}

// ── Server-side product search combobox ───────────────────────────────────────
function ProductSearchSelect({
  storeId,
  value,
  selectedName,
  onSelect,
}: {
  storeId: string;
  value: string;
  selectedName?: string;
  onSelect: (product: Product) => void;
}) {
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState('');
  const [debouncedQuery, setDebouncedQuery] = useState('');
  const [results, setResults] = useState<Product[]>([]);
  const [searching, setSearching] = useState(false);
  const [dropdownStyle, setDropdownStyle] = useState<React.CSSProperties>({});
  const triggerRef = useRef<HTMLDivElement>(null);
  const dropdownRef = useRef<HTMLDivElement>(null);

  // Debounce
  useEffect(() => {
    const t = setTimeout(() => setDebouncedQuery(query), 300);
    return () => clearTimeout(t);
  }, [query]);

  // Fetch from API
  useEffect(() => {
    if (!open) return;
    let cancelled = false;
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setSearching(true);
    productsApi
      .list(storeId, { search: debouncedQuery || undefined, per_page: 20 })
      .then(res => {
        if (!cancelled) setResults(res.data.data ?? []);
      })
      .catch(console.error)
      .finally(() => {
        if (!cancelled) setSearching(false);
      });
    return () => {
      cancelled = true;
    };
  }, [storeId, debouncedQuery, open]);

  // Reposition dropdown on open
  useEffect(() => {
    if (!open || !triggerRef.current) return;
    const rect = triggerRef.current.getBoundingClientRect();
    const spaceBelow = window.innerHeight - rect.bottom;
    const dropHeight = 300;
    if (spaceBelow >= dropHeight) {
      setDropdownStyle({
        position: 'fixed',
        top: rect.bottom + 4,
        left: rect.left,
        width: rect.width,
        zIndex: 9999,
      });
    } else {
      // Open upwards if not enough space below
      setDropdownStyle({
        position: 'fixed',
        bottom: window.innerHeight - rect.top + 4,
        left: rect.left,
        width: rect.width,
        zIndex: 9999,
      });
    }
  }, [open]);

  // Click outside (both trigger and dropdown)
  useEffect(() => {
    function handler(e: MouseEvent) {
      if (
        triggerRef.current &&
        !triggerRef.current.contains(e.target as Node) &&
        dropdownRef.current &&
        !dropdownRef.current.contains(e.target as Node)
      ) {
        setOpen(false);
      }
    }
    document.addEventListener('mousedown', handler);
    return () => document.removeEventListener('mousedown', handler);
  }, []);

  const displayName = selectedName || (value ? value : '');

  return (
    <div style={{ position: 'relative', width: '100%' }}>
      <div
        ref={triggerRef}
        className="input"
        style={{
          cursor: 'pointer',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          minHeight: 38,
        }}
        onClick={() => {
          setOpen(o => !o);
          setQuery('');
        }}
      >
        <span
          style={{
            overflow: 'hidden',
            textOverflow: 'ellipsis',
            whiteSpace: 'nowrap',
            color: displayName ? 'inherit' : 'var(--text-3)',
          }}
        >
          {displayName || 'Pilih produk...'}
        </span>
        <ChevronRight size={14} style={{ transform: open ? 'rotate(90deg)' : 'rotate(0)' }} />
      </div>

      {open && (
        <div
          ref={dropdownRef}
          style={{
            ...dropdownStyle,
            background: 'var(--bg-card)',
            border: '1px solid var(--border)',
            borderRadius: 8,
            boxShadow: '0 8px 24px rgba(0,0,0,0.2)',
            maxHeight: 300,
            display: 'flex',
            flexDirection: 'column',
          }}
        >
          <div style={{ padding: 8, borderBottom: '1px solid var(--border)' }}>
            <input
              autoFocus
              className="input"
              placeholder="Cari nama / SKU produk..."
              style={{ width: '100%', height: 32, fontSize: '0.82rem' }}
              value={query}
              onChange={e => setQuery(e.target.value)}
            />
          </div>
          <div style={{ overflowY: 'auto', flex: 1, padding: 4 }}>
            {searching ? (
              <div
                style={{
                  padding: '12px',
                  textAlign: 'center',
                  fontSize: '0.82rem',
                  color: 'var(--text-3)',
                }}
              >
                Mencari...
              </div>
            ) : results.length === 0 ? (
              <div
                style={{
                  padding: '12px',
                  textAlign: 'center',
                  fontSize: '0.82rem',
                  color: 'var(--text-3)',
                }}
              >
                {debouncedQuery ? 'Tidak ditemukan' : 'Ketik untuk mencari produk'}
              </div>
            ) : (
              results.map(p => (
                <div
                  key={p.id}
                  style={{
                    padding: '8px 12px',
                    cursor: 'pointer',
                    fontSize: '0.82rem',
                    borderRadius: 6,
                    background: p.id === value ? 'var(--bg-active)' : 'transparent',
                  }}
                  onClick={() => {
                    onSelect(p);
                    setOpen(false);
                  }}
                  onMouseEnter={e => {
                    if (p.id !== value) e.currentTarget.style.background = 'var(--bg-hover)';
                  }}
                  onMouseLeave={e => {
                    if (p.id !== value) e.currentTarget.style.background = 'transparent';
                  }}
                >
                  <div style={{ fontWeight: 500 }}>{p.name}</div>
                  <div style={{ fontSize: '0.75rem', color: 'var(--text-3)' }}>
                    SKU: {p.sku} · {p.unit}
                  </div>
                </div>
              ))
            )}
          </div>
        </div>
      )}
    </div>
  );
}

type ActionType = 'submit' | 'receive' | 'cancel';

interface PayableSummary {
  total_debt: number;
  total_paid: number;
  total_outstanding: number;
  unpaid_count: number;
  partial_count: number;
}
interface POPayment {
  id: string;
  po_id: string;
  amount: number;
  note?: string;
  paid_by_name: string;
  paid_at: string;
}

interface SavedPOItem {
  product_id: string;
  product_name?: string;
  product_sku?: string;
  unit?: string;
  quantity: number;
  unit_cost: number;
}

type PurchaseOrderListResponse = { data?: PurchaseOrder[] };
type SupplierListResponse = { data?: Supplier[] };

// ── PayStatus badge ───────────────────────────────────────────────────────────
function PayStatusBadge({ status }: { status: string }) {
  const cfg = PAY_STATUS_CFG[status as keyof typeof PAY_STATUS_CFG] ?? PAY_STATUS_CFG.unpaid;
  const Icon = cfg.icon;
  return (
    <span
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        gap: 4,
        padding: '2px 8px',
        borderRadius: 6,
        fontSize: '0.72rem',
        fontWeight: 600,
        background: cfg.bg,
        color: cfg.color,
        whiteSpace: 'nowrap',
      }}
    >
      <Icon size={11} /> {cfg.label}
    </span>
  );
}

// ── Termin helpers ───────────────────────────────────────────────────────────

function terminStatusCfg(t: Termin) {
  if (t.status === 'paid') return { label: 'Lunas', bg: '#dcfce7', color: '#16a34a' };
  if (t.status === 'overdue' || t.is_overdue)
    return { label: 'Jatuh Tempo', bg: '#fee2e2', color: '#dc2626' };
  if (t.status === 'partial') return { label: 'Sebagian', bg: '#fef3c7', color: '#d97706' };
  return { label: 'Belum Bayar', bg: '#f3f4f6', color: '#6b7280' };
}

function formatIDR(n: number) {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(n);
}

// ── TerminPanel ───────────────────────────────────────────────────────────────

interface TerminPanelProps {
  po: PurchaseOrder;
  storeId: string;
  onOpenDoc: (type: 'invoice' | 'receipt' | 'termin_agreement') => void;
  onUpdate?: () => void;
}

function TerminPanel({ po, storeId, onOpenDoc, onUpdate }: TerminPanelProps) {
  const [termins, setTermins] = useState<Termin[]>([]);
  const [loading, setLoading] = useState(true);
  const [showAddModal, setShowAddModal] = useState(false);
  const [payTarget, setPayTarget] = useState<Termin | null>(null);
  const [expandedPayments, setExpandedPayments] = useState<Set<string>>(new Set());

  const load = useCallback(() => {
    setLoading(true);
    listTermins(storeId, po.id)
      .then(setTermins)
      .catch(console.error)
      .finally(() => setLoading(false));
  }, [storeId, po.id]);

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    load();
  }, [load]);

  const togglePay = (id: string) =>
    setExpandedPayments(prev => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });

  return (
    <div style={{ padding: '12px 16px 16px', background: 'rgba(255,255,255,0.03)' }}>
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          marginBottom: 12,
        }}
      >
        <span style={{ fontWeight: 600, fontSize: '0.82rem', color: 'var(--text-2)' }}>
          Jadwal Termin
        </span>
        <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
          {(['invoice', 'receipt', 'termin_agreement'] as const).map(type => (
            <button
              key={type}
              id={`doc-${type}-${po.id}`}
              onClick={() => onOpenDoc(type)}
              style={{
                padding: '4px 10px',
                borderRadius: 6,
                border: '1px solid var(--border)',
                background: 'transparent',
                color: 'var(--text-2)',
                fontSize: '0.75rem',
                cursor: 'pointer',
              }}
            >
              {type === 'invoice'
                ? '📄 Invoice'
                : type === 'receipt'
                  ? '🧾 Kwitansi'
                  : '📋 Perjanjian'}
            </button>
          ))}
          {po.status === 'received' && (
            <button
              id={`add-termin-${po.id}`}
              onClick={() => setShowAddModal(true)}
              style={{
                padding: '4px 12px',
                borderRadius: 6,
                border: 'none',
                background: 'var(--accent-em)',
                color: '#fff',
                fontSize: '0.75rem',
                cursor: 'pointer',
                fontWeight: 600,
              }}
            >
              + Termin
            </button>
          )}
        </div>
      </div>

      {loading ? (
        <div style={{ textAlign: 'center', padding: 20, color: 'var(--text-3)' }}>Memuat…</div>
      ) : termins.length === 0 ? (
        <div
          style={{ textAlign: 'center', padding: 20, color: 'var(--text-3)', fontSize: '0.82rem' }}
        >
          {po.status === 'received'
            ? 'Belum ada termin. Tambahkan jadwal pembayaran.'
            : 'PO harus diterima sebelum termin dapat dibuat.'}
        </div>
      ) : (
        <div style={{ overflowX: 'auto' }}>
          <table className="tbl" style={{ fontSize: '0.82rem' }}>
            <thead>
              <tr>
                <th>No.</th>
                <th>Jatuh Tempo</th>
                <th>Jumlah</th>
                <th>Dibayar</th>
                <th>Sisa</th>
                <th>Status</th>
                <th>Aksi</th>
              </tr>
            </thead>
            <tbody>
              {termins.map(t => {
                const cfg = terminStatusCfg(t);
                return (
                  <Fragment key={t.id}>
                    <tr>
                      <td style={{ fontWeight: 600 }}>Termin {t.termin_number}</td>
                      <td style={{ color: t.is_overdue ? '#dc2626' : 'inherit' }}>
                        {formatDate(t.due_date)}
                      </td>
                      <td>{formatIDR(t.amount)}</td>
                      <td style={{ color: '#16a34a', fontWeight: 600 }}>
                        {formatIDR(t.amount_paid)}
                      </td>
                      <td
                        style={{ color: t.amount_due > 0 ? '#dc2626' : '#16a34a', fontWeight: 600 }}
                      >
                        {formatIDR(t.amount_due)}
                      </td>
                      <td>
                        <span
                          style={{
                            background: cfg.bg,
                            color: cfg.color,
                            borderRadius: 5,
                            padding: '2px 8px',
                            fontSize: '0.75rem',
                            fontWeight: 600,
                          }}
                        >
                          {cfg.label}
                        </span>
                      </td>
                      <td>
                        <div style={{ display: 'flex', gap: 6 }}>
                          {t.status !== 'paid' && (
                            <button
                              id={`pay-termin-${t.id}`}
                              onClick={() => setPayTarget(t)}
                              style={{
                                padding: '3px 10px',
                                borderRadius: 5,
                                border: 'none',
                                background: 'var(--accent-em)',
                                color: '#fff',
                                fontSize: '0.75rem',
                                cursor: 'pointer',
                              }}
                            >
                              Bayar
                            </button>
                          )}
                          {t.payments.length > 0 && (
                            <button
                              id={`hist-${t.id}`}
                              onClick={() => togglePay(t.id)}
                              style={{
                                padding: '3px 10px',
                                borderRadius: 5,
                                border: '1px solid var(--border)',
                                background: 'transparent',
                                color: 'var(--text-2)',
                                fontSize: '0.75rem',
                                cursor: 'pointer',
                              }}
                            >
                              {expandedPayments.has(t.id) ? '▲' : '▼'} {t.payments.length}
                            </button>
                          )}
                        </div>
                      </td>
                    </tr>
                    {expandedPayments.has(t.id) &&
                      t.payments.map(p => (
                        <tr
                          key={p.id}
                          style={{ background: 'rgba(59,130,246,0.04)', fontSize: '0.78rem' }}
                        >
                          <td colSpan={2} style={{ paddingLeft: 28, color: 'var(--text-3)' }}>
                            {formatDate(p.payment_date)} · {p.payment_method}
                          </td>
                          <td colSpan={2} style={{ color: '#16a34a', fontWeight: 600 }}>
                            +{formatIDR(p.amount_paid)}
                          </td>
                          <td colSpan={2} style={{ color: 'var(--text-3)' }}>
                            {p.notes || '—'}
                          </td>
                          <td style={{ color: 'var(--text-3)' }}>{p.recorded_by_name}</td>
                        </tr>
                      ))}
                  </Fragment>
                );
              })}
            </tbody>
          </table>
        </div>
      )}
      {showAddModal && (
        <TerminModal
          po={po}
          storeId={storeId}
          onSuccess={() => {
            setShowAddModal(false);
            load();
            onUpdate?.();
          }}
          onCancel={() => setShowAddModal(false)}
        />
      )}
      {payTarget && (
        <PayTerminModal
          termin={payTarget}
          storeId={storeId}
          poId={po.id}
          onSuccess={() => {
            setPayTarget(null);
            load();
            onUpdate?.();
          }}
          onCancel={() => setPayTarget(null)}
        />
      )}
    </div>
  );
}

// ── TerminModal ───────────────────────────────────────────────────────────────

interface TerminModalProps {
  po: PurchaseOrder;
  storeId: string;
  onSuccess: () => void;
  onCancel: () => void;
}

function TerminModal({ po, storeId, onSuccess, onCancel }: TerminModalProps) {
  const today = new Date().toISOString().slice(0, 10);
  const [rows, setRows] = useState([
    { termin_number: 1, amount: po.total_amount, due_date: today, notes: '' },
  ]);
  const [saving, setSaving] = useState(false);
  const [err, setErr] = useState('');
  const updateRow = (i: number, field: string, value: string | number) =>
    setRows(prev => prev.map((r, idx) => (idx === i ? { ...r, [field]: value } : r)));
  const recalculate = (currentRows: typeof rows) => {
    const count = currentRows.length;
    if (count <= 0) return currentRows;

    const splitAmount = Math.floor(po.total_amount / count);
    const remainder = po.total_amount - splitAmount * count;
    const baseDateStr = currentRows[0]?.due_date || today;

    return currentRows.map((r, i) => {
      const d = new Date(baseDateStr);
      d.setDate(d.getDate() + i);
      let amt = splitAmount;
      if (i === count - 1) amt += remainder;
      return { ...r, termin_number: i + 1, amount: amt, due_date: d.toISOString().slice(0, 10) };
    });
  };

  const addRow = () => {
    setRows(prev =>
      recalculate([
        ...prev,
        { termin_number: prev.length + 1, amount: 0, due_date: today, notes: '' },
      ])
    );
  };

  const removeRow = (i: number) => {
    setRows(prev => recalculate(prev.filter((_, idx) => idx !== i)));
  };
  const totalTermin = rows.reduce((s, r) => s + Number(r.amount), 0);
  const handleSave = async () => {
    setSaving(true);
    setErr('');
    try {
      await createTerminSchedule(storeId, po.id, {
        termins: rows.map(r => ({ ...r, amount: Number(r.amount) })),
      });
      onSuccess();
    } catch (e: unknown) {
      setErr(e instanceof Error ? e.message : 'Gagal');
    } finally {
      setSaving(false);
    }
  };
  return (
    <div
      style={{
        position: 'fixed',
        inset: 0,
        background: 'rgba(0,0,0,0.55)',
        zIndex: 200,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
      }}
    >
      <div
        style={{
          background: 'var(--bg-card)',
          borderRadius: 14,
          padding: 28,
          width: 620,
          maxHeight: '85vh',
          overflowY: 'auto',
        }}
      >
        <div style={{ fontWeight: 700, fontSize: '1rem', marginBottom: 12 }}>
          Buat Jadwal Termin — {po.po_number}
        </div>
        <div style={{ fontSize: '0.82rem', color: 'var(--text-2)', marginBottom: 14 }}>
          Total PO: <strong>{formatIDR(po.total_amount)}</strong> · Total termin:{' '}
          <strong
            style={{
              color: Math.abs(totalTermin - po.total_amount) < 0.01 ? '#16a34a' : '#d97706',
            }}
          >
            {formatIDR(totalTermin)}
          </strong>
        </div>
        {rows.map((row, i) => (
          <div
            key={i}
            style={{
              display: 'grid',
              gridTemplateColumns: '50px 1fr 1.2fr 1fr 34px',
              gap: 8,
              marginBottom: 8,
              alignItems: 'end',
            }}
          >
            <div>
              <label
                style={{
                  fontSize: '0.72rem',
                  color: 'var(--text-3)',
                  display: 'block',
                  marginBottom: 2,
                }}
              >
                No.
              </label>
              <input
                type="number"
                value={row.termin_number}
                disabled
                style={{
                  width: '100%',
                  padding: '6px 7px',
                  borderRadius: 7,
                  border: '1px solid var(--border)',
                  background: 'var(--bg-hover)',
                  color: 'var(--text-3)',
                }}
              />
            </div>
            <div>
              <label
                style={{
                  fontSize: '0.72rem',
                  color: 'var(--text-3)',
                  display: 'block',
                  marginBottom: 2,
                }}
              >
                Jumlah (Rp)
              </label>
              <input
                id={`ta-${i}`}
                type="text"
                value={formatNumberInput(row.amount)}
                onChange={e => updateRow(i, 'amount', parseNumberInput(e.target.value))}
                style={{
                  width: '100%',
                  padding: '6px 7px',
                  borderRadius: 7,
                  border: '1px solid var(--border)',
                  background: 'var(--bg-card)',
                  color: 'var(--text-1)',
                }}
              />
            </div>
            <div>
              <label
                style={{
                  fontSize: '0.72rem',
                  color: 'var(--text-3)',
                  display: 'block',
                  marginBottom: 2,
                }}
              >
                Jatuh Tempo
              </label>
              <input
                id={`td-${i}`}
                type="date"
                value={row.due_date}
                onChange={e => updateRow(i, 'due_date', e.target.value)}
                style={{
                  width: '100%',
                  padding: '6px 7px',
                  borderRadius: 7,
                  border: '1px solid var(--border)',
                  background: 'var(--bg-card)',
                  color: 'var(--text-1)',
                }}
              />
            </div>
            <div>
              <label
                style={{
                  fontSize: '0.72rem',
                  color: 'var(--text-3)',
                  display: 'block',
                  marginBottom: 2,
                }}
              >
                Catatan
              </label>
              <input
                type="text"
                value={row.notes}
                onChange={e => updateRow(i, 'notes', e.target.value)}
                placeholder="opsional"
                style={{
                  width: '100%',
                  padding: '6px 7px',
                  borderRadius: 7,
                  border: '1px solid var(--border)',
                  background: 'var(--bg-card)',
                  color: 'var(--text-1)',
                }}
              />
            </div>
            <button
              onClick={() => removeRow(i)}
              disabled={rows.length === 1}
              style={{
                padding: '6px 7px',
                borderRadius: 7,
                border: 'none',
                background: '#fee2e2',
                color: '#dc2626',
                cursor: 'pointer',
              }}
            >
              ✕
            </button>
          </div>
        ))}
        <button
          id="add-row-btn"
          onClick={addRow}
          style={{
            marginBottom: 14,
            padding: '5px 14px',
            borderRadius: 7,
            border: '1px dashed var(--border)',
            background: 'transparent',
            color: 'var(--text-2)',
            cursor: 'pointer',
            fontSize: '0.82rem',
          }}
        >
          + Tambah Termin
        </button>
        {err && (
          <div style={{ color: '#dc2626', fontSize: '0.82rem', marginBottom: 10 }}>{err}</div>
        )}
        <div style={{ display: 'flex', gap: 10, justifyContent: 'flex-end' }}>
          <button
            onClick={onCancel}
            style={{
              padding: '8px 18px',
              borderRadius: 8,
              border: '1px solid var(--border)',
              background: 'transparent',
              color: 'var(--text-2)',
              cursor: 'pointer',
            }}
          >
            Batal
          </button>
          <button
            id="save-schedule-btn"
            onClick={handleSave}
            disabled={saving}
            style={{
              padding: '8px 18px',
              borderRadius: 8,
              border: 'none',
              background: 'var(--accent-em)',
              color: '#fff',
              cursor: 'pointer',
              fontWeight: 600,
            }}
          >
            {saving ? '…' : 'Simpan Jadwal'}
          </button>
        </div>
      </div>
    </div>
  );
}

// ── PayTerminModal ────────────────────────────────────────────────────────────

interface PayTerminModalProps {
  termin: Termin;
  storeId: string;
  poId: string;
  onSuccess: () => void;
  onCancel: () => void;
}

function PayTerminModal({ termin, storeId, poId, onSuccess, onCancel }: PayTerminModalProps) {
  const today = new Date().toISOString().slice(0, 10);
  const [amount, setAmount] = useState(String(termin.amount_due.toFixed(0)));
  const [date, setDate] = useState(today);
  const [method, setMethod] = useState<RecordPaymentRequest['payment_method']>('cash');
  const [notes, setNotes] = useState('');
  const [saving, setSaving] = useState(false);
  const [err, setErr] = useState('');
  const handleSave = async () => {
    const n = parseFloat(amount);
    if (isNaN(n) || n <= 0) {
      setErr('Jumlah harus > 0');
      return;
    }
    if (n > termin.amount_due) {
      setErr(`Melebihi sisa termin (${formatIDR(termin.amount_due)})`);
      return;
    }
    setSaving(true);
    setErr('');
    try {
      await recordPayment(storeId, poId, termin.id, {
        amount_paid: n,
        payment_date: date,
        payment_method: method,
        notes,
      });
      onSuccess();
    } catch (e: unknown) {
      setErr(e instanceof Error ? e.message : 'Gagal');
    } finally {
      setSaving(false);
    }
  };
  return (
    <div
      style={{
        position: 'fixed',
        inset: 0,
        background: 'rgba(0,0,0,0.55)',
        zIndex: 210,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
      }}
    >
      <div style={{ background: 'var(--bg-card)', borderRadius: 14, padding: 28, width: 420 }}>
        <div style={{ fontWeight: 700, fontSize: '1rem', marginBottom: 4 }}>
          Catat Pembayaran — Termin {termin.termin_number}
        </div>
        <div style={{ fontSize: '0.82rem', color: 'var(--text-2)', marginBottom: 18 }}>
          Sisa: <strong style={{ color: '#dc2626' }}>{formatIDR(termin.amount_due)}</strong>
        </div>
        {[
          {
            label: 'Jumlah Bayar (Rp)',
            el: (
              <input
                id="pay-amt"
                type="text"
                value={formatNumberInput(amount)}
                onChange={e => setAmount(String(parseNumberInput(e.target.value)))}
                style={{
                  width: '100%',
                  padding: '8px 12px',
                  borderRadius: 8,
                  border: '1px solid var(--border)',
                  background: 'var(--bg-card)',
                  color: 'var(--text-1)',
                }}
              />
            ),
          },
          {
            label: 'Tanggal Bayar',
            el: (
              <input
                id="pay-dt"
                type="date"
                value={date}
                onChange={e => setDate(e.target.value)}
                style={{
                  width: '100%',
                  padding: '8px 12px',
                  borderRadius: 8,
                  border: '1px solid var(--border)',
                  background: 'var(--bg-card)',
                  color: 'var(--text-1)',
                }}
              />
            ),
          },
          {
            label: 'Metode',
            el: (
              <select
                id="pay-mth"
                value={method}
                onChange={e => setMethod(e.target.value as typeof method)}
                style={{
                  width: '100%',
                  padding: '8px 12px',
                  borderRadius: 8,
                  border: '1px solid var(--border)',
                  background: 'var(--bg-card)',
                  color: 'var(--text-1)',
                }}
              >
                {['cash', 'transfer', 'check', 'other'].map(m => (
                  <option key={m} value={m}>
                    {m[0].toUpperCase() + m.slice(1)}
                  </option>
                ))}
              </select>
            ),
          },
          {
            label: 'Catatan',
            el: (
              <input
                id="pay-nt"
                type="text"
                value={notes}
                onChange={e => setNotes(e.target.value)}
                placeholder="opsional"
                style={{
                  width: '100%',
                  padding: '8px 12px',
                  borderRadius: 8,
                  border: '1px solid var(--border)',
                  background: 'var(--bg-card)',
                  color: 'var(--text-1)',
                }}
              />
            ),
          },
        ].map(({ label, el }) => (
          <div key={label} style={{ marginBottom: 12 }}>
            <label
              style={{
                fontSize: '0.78rem',
                color: 'var(--text-2)',
                display: 'block',
                marginBottom: 4,
              }}
            >
              {label}
            </label>
            {el}
          </div>
        ))}
        {err && (
          <div style={{ color: '#dc2626', fontSize: '0.82rem', marginBottom: 10 }}>{err}</div>
        )}
        <div style={{ display: 'flex', gap: 10, justifyContent: 'flex-end', marginTop: 8 }}>
          <button
            onClick={onCancel}
            style={{
              padding: '8px 18px',
              borderRadius: 8,
              border: '1px solid var(--border)',
              background: 'transparent',
              color: 'var(--text-2)',
              cursor: 'pointer',
            }}
          >
            Batal
          </button>
          <button
            id="save-pay-btn"
            onClick={handleSave}
            disabled={saving}
            style={{
              padding: '8px 18px',
              borderRadius: 8,
              border: 'none',
              background: 'var(--accent-em)',
              color: '#fff',
              cursor: 'pointer',
              fontWeight: 600,
            }}
          >
            {saving ? '…' : 'Catat Pembayaran'}
          </button>
        </div>
      </div>
    </div>
  );
}

// ── Confirm Modal ─────────────────────────────────────────────────────────────
interface ConfirmModalProps {
  action: ActionType;
  po: PurchaseOrder;
  onConfirm: () => void;
  onCancel: () => void;
  loading: boolean;
}
function ConfirmModal({ action, po, onConfirm, onCancel, loading }: ConfirmModalProps) {
  const config = {
    submit: {
      icon: <CheckCircle2 size={28} style={{ color: '#6366f1' }} />,
      title: 'Submit Pembelian?',
      body: `PO ${po.po_number} akan dikirim ke supplier. Status berubah menjadi Dipesan.`,
      confirmLabel: 'Ya, Submit PO',
      confirmClass: 'btn btn-primary',
    },
    receive: {
      icon: <Package size={28} style={{ color: '#10b981' }} />,
      title: 'Terima Pembelian?',
      body: `Konfirmasi penerimaan PO ${po.po_number}. Stok akan bertambah, HPP produk akan diperbarui, dan hutang ke supplier akan tercatat.`,
      confirmLabel: 'Ya, Terima PO',
      confirmClass: 'btn btn-primary',
    },
    cancel: {
      icon: <AlertTriangle size={28} style={{ color: '#ef4444' }} />,
      title: 'Batalkan Pembelian?',
      body: `PO ${po.po_number} akan dibatalkan. Tindakan ini tidak dapat dibatalkan.`,
      confirmLabel: 'Ya, Batalkan',
      confirmClass: 'btn btn-danger',
    },
  }[action];
  return (
    <div className="modal-overlay" onClick={onCancel}>
      <div className="modal-box" style={{ maxWidth: 420 }} onClick={e => e.stopPropagation()}>
        <div style={{ textAlign: 'center', padding: '8px 0 16px' }}>
          <div style={{ marginBottom: 12 }}>{config.icon}</div>
          <h2 style={{ fontWeight: 800, fontSize: '1.05rem', marginBottom: 10 }}>{config.title}</h2>
          <p style={{ color: 'var(--text-2)', fontSize: '0.875rem', lineHeight: 1.6 }}>
            {config.body}
          </p>
        </div>
        <div style={{ display: 'flex', gap: 8 }}>
          <button
            className="btn btn-secondary"
            style={{ flex: 1 }}
            onClick={onCancel}
            disabled={loading}
          >
            Batal
          </button>
          <button
            className={config.confirmClass}
            style={{ flex: 1 }}
            onClick={onConfirm}
            disabled={loading}
          >
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
  po: PurchaseOrder;
  storeId: string;
  onSuccess: () => void;
  onCancel: () => void;
}
function PayModal({ po, storeId, onSuccess, onCancel }: PayModalProps) {
  const amountPaid = po.amount_paid ?? 0;
  const amountDue = po.amount_due ?? (po.total_amount ?? 0) - amountPaid;
  const [amount, setAmount] = useState(String(amountDue > 0 ? amountDue.toFixed(0) : ''));
  const [note, setNote] = useState('');
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async () => {
    const n = parseFloat(amount);
    if (!n || n <= 0) {
      setError('Jumlah harus lebih dari 0');
      return;
    }
    setSaving(true);
    setError('');
    try {
      await purchaseOrdersApi.createPayment(storeId, po.id, { amount: n, note: note || undefined });
      onSuccess();
    } catch (e) {
      setError(e instanceof ApiError ? e.message : 'Gagal menyimpan pembayaran');
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="modal-overlay" onClick={onCancel}>
      <div className="modal-box" style={{ maxWidth: 400 }} onClick={e => e.stopPropagation()}>
        <div
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            marginBottom: 16,
          }}
        >
          <h2 style={{ fontWeight: 800 }}>Bayar Hutang</h2>
          <button className="btn btn-ghost btn-sm" onClick={onCancel}>
            <X size={15} />
          </button>
        </div>
        {/* PO summary */}
        <div
          style={{
            background: 'var(--bg-elevated)',
            borderRadius: 10,
            padding: '12px 14px',
            marginBottom: 16,
          }}
        >
          <div style={{ fontSize: '0.8rem', fontWeight: 700 }}>{po.po_number}</div>
          <div
            style={{
              display: 'flex',
              justifyContent: 'space-between',
              marginTop: 6,
              fontSize: '0.82rem',
            }}
          >
            <span style={{ color: 'var(--text-3)' }}>Total PO</span>
            <span style={{ fontWeight: 600 }}>{formatRp(po.total_amount)}</span>
          </div>
          <div
            style={{
              display: 'flex',
              justifyContent: 'space-between',
              marginTop: 4,
              fontSize: '0.82rem',
            }}
          >
            <span style={{ color: 'var(--text-3)' }}>Sudah dibayar</span>
            <span style={{ color: '#10b981', fontWeight: 600 }}>{formatRp(amountPaid)}</span>
          </div>
          <div
            style={{
              display: 'flex',
              justifyContent: 'space-between',
              marginTop: 4,
              fontSize: '0.9rem',
              borderTop: '1px solid var(--border)',
              paddingTop: 8,
            }}
          >
            <span style={{ fontWeight: 700 }}>Sisa hutang</span>
            <span style={{ fontWeight: 800, color: '#ef4444' }}>{formatRp(amountDue)}</span>
          </div>
        </div>
        {error && (
          <div
            style={{
              background: 'rgba(239,68,68,0.1)',
              borderRadius: 8,
              padding: '8px 12px',
              color: '#f87171',
              fontSize: '0.83rem',
              marginBottom: 12,
              border: '1px solid rgba(239,68,68,0.3)',
            }}
          >
            {error}
          </div>
        )}
        <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
          <div className="input-group">
            <label className="input-label">Jumlah Bayar (Rp)</label>
            <input
              type="number"
              className="input"
              value={amount}
              onChange={e => setAmount(e.target.value)}
              min={1}
            />
          </div>
          <div className="input-group">
            <label className="input-label">Catatan (opsional)</label>
            <input
              className="input"
              value={note}
              onChange={e => setNote(e.target.value)}
              placeholder="Transfer, tunai, dll..."
            />
          </div>
        </div>
        <div style={{ display: 'flex', gap: 8, marginTop: 16 }}>
          <button className="btn btn-secondary" style={{ flex: 1 }} onClick={onCancel}>
            Batal
          </button>
          <button
            className="btn btn-primary"
            style={{ flex: 1 }}
            onClick={handleSubmit}
            disabled={saving}
          >
            {saving ? <Loader2 size={14} className="loading-spin" /> : <Wallet size={14} />}
            {saving ? 'Menyimpan...' : 'Bayar'}
          </button>
        </div>
      </div>
    </div>
  );
}

// ── Invoice Modal ─────────────────────────────────────────────────────────────
interface InvoiceModalProps {
  po: PurchaseOrder;
  store: Store | null;
  onClose: () => void;
}
function InvoiceModal({ po, store, onClose }: InvoiceModalProps) {
  return (
    <>
      <div className="modal-overlay no-print" onClick={onClose} />
      <div
        style={{
          position: 'fixed',
          inset: 0,
          zIndex: 202,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          pointerEvents: 'none',
        }}
      >
        <div
          id="po-invoice"
          style={{
            background: '#fff',
            color: '#111',
            borderRadius: 12,
            width: '100%',
            maxWidth: 680,
            maxHeight: '92vh',
            overflowY: 'auto',
            pointerEvents: 'auto',
            boxShadow: '0 24px 80px rgba(0,0,0,0.4)',
            fontFamily: '"Inter","Helvetica Neue",Arial,sans-serif',
          }}
        >
          <div
            className="no-print"
            style={{
              display: 'flex',
              justifyContent: 'flex-end',
              gap: 8,
              padding: '12px 16px',
              borderBottom: '1px solid #e5e7eb',
            }}
          >
            <button
              onClick={() => window.print()}
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: 6,
                padding: '8px 16px',
                borderRadius: 8,
                border: 'none',
                background: '#111',
                color: '#fff',
                fontWeight: 600,
                fontSize: '0.85rem',
                cursor: 'pointer',
              }}
            >
              <Printer size={14} /> Cetak Invoice
            </button>
            <button
              onClick={onClose}
              style={{
                padding: '8px 12px',
                borderRadius: 8,
                border: '1px solid #e5e7eb',
                background: 'transparent',
                cursor: 'pointer',
              }}
            >
              <X size={14} />
            </button>
          </div>
          <div style={{ padding: '32px 40px' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 24 }}>
              <div>
                <div style={{ fontWeight: 800, fontSize: '1.4rem', marginBottom: 4 }}>
                  {store?.name ?? 'Toko'}
                </div>
                {store?.address && (
                  <div style={{ fontSize: '0.82rem', color: '#6b7280' }}>{store.address}</div>
                )}
                {store?.phone && (
                  <div style={{ fontSize: '0.82rem', color: '#6b7280' }}>Telp: {store.phone}</div>
                )}
              </div>
              <div style={{ textAlign: 'right' }}>
                <div
                  style={{
                    display: 'inline-block',
                    padding: '4px 12px',
                    borderRadius: 6,
                    background: '#f3f4f6',
                    fontSize: '0.75rem',
                    fontWeight: 700,
                    color: '#374151',
                    marginBottom: 8,
                  }}
                >
                  PURCHASE ORDER
                </div>
                <div style={{ fontWeight: 800, fontSize: '1.15rem' }}>{po.po_number}</div>
                <div style={{ fontSize: '0.8rem', color: '#6b7280' }}>
                  Tanggal: {formatDate(po.created_at)}
                </div>
              </div>
            </div>
            <div style={{ borderTop: '2px solid #111', marginBottom: 20 }} />
            <div
              style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 24, marginBottom: 24 }}
            >
              <div>
                <div
                  style={{
                    fontSize: '0.7rem',
                    fontWeight: 700,
                    color: '#9ca3af',
                    letterSpacing: '0.08em',
                    marginBottom: 6,
                  }}
                >
                  DARI (PEMBELI)
                </div>
                <div style={{ fontWeight: 700 }}>{store?.name ?? '—'}</div>
                {store?.address && (
                  <div style={{ fontSize: '0.8rem', color: '#4b5563' }}>{store.address}</div>
                )}
              </div>
              <div>
                <div
                  style={{
                    fontSize: '0.7rem',
                    fontWeight: 700,
                    color: '#9ca3af',
                    letterSpacing: '0.08em',
                    marginBottom: 6,
                  }}
                >
                  KEPADA (SUPPLIER)
                </div>
                <div style={{ fontWeight: 700 }}>{po.supplier_name ?? 'Tanpa Supplier'}</div>
              </div>
            </div>
            <table style={{ width: '100%', borderCollapse: 'collapse', marginBottom: 20 }}>
              <thead>
                <tr style={{ borderBottom: '2px solid #e5e7eb' }}>
                  {['#', 'Nama Produk', 'SKU', 'Qty', 'Satuan', 'Harga Beli', 'Subtotal'].map(h => (
                    <th
                      key={h}
                      style={{
                        padding: '8px 10px',
                        textAlign:
                          h === 'Harga Beli' || h === 'Subtotal'
                            ? 'right'
                            : h === '#' || h === 'Qty'
                              ? 'center'
                              : 'left',
                        fontSize: '0.72rem',
                        fontWeight: 700,
                        color: '#6b7280',
                        textTransform: 'uppercase',
                      }}
                    >
                      {h}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {(po.items ?? []).map((item, i) => (
                  <tr key={item.id ?? i} style={{ borderBottom: '1px solid #f3f4f6' }}>
                    <td
                      style={{
                        padding: '10px',
                        textAlign: 'center',
                        color: '#9ca3af',
                        fontSize: '0.8rem',
                      }}
                    >
                      {i + 1}
                    </td>
                    <td style={{ padding: '10px', fontWeight: 600, fontSize: '0.85rem' }}>
                      {item.product_name}
                    </td>
                    <td
                      style={{
                        padding: '10px',
                        color: '#6b7280',
                        fontSize: '0.78rem',
                        fontFamily: 'monospace',
                      }}
                    >
                      {item.product_sku}
                    </td>
                    <td style={{ padding: '10px', textAlign: 'center', fontWeight: 600 }}>
                      {item.quantity}
                    </td>
                    <td style={{ padding: '10px', color: '#6b7280' }}>{item.unit}</td>
                    <td style={{ padding: '10px', textAlign: 'right' }}>
                      {formatRp(item.unit_cost)}
                    </td>
                    <td style={{ padding: '10px', textAlign: 'right', fontWeight: 700 }}>
                      {formatRp(item.subtotal)}
                    </td>
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
            {po.notes && (
              <div
                style={{
                  background: '#f9fafb',
                  borderRadius: 8,
                  padding: '12px 16px',
                  marginBottom: 20,
                }}
              >
                <div
                  style={{
                    fontSize: '0.72rem',
                    color: '#9ca3af',
                    fontWeight: 700,
                    marginBottom: 4,
                  }}
                >
                  CATATAN
                </div>
                <div style={{ fontSize: '0.85rem' }}>{po.notes}</div>
              </div>
            )}
            <div
              style={{
                borderTop: '1px solid #e5e7eb',
                paddingTop: 16,
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'flex-end',
              }}
            >
              <div style={{ fontSize: '0.72rem', color: '#9ca3af' }}>
                Dokumen dibuat otomatis oleh MoedahPOS
              </div>
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
  po: PurchaseOrder;
  storeId: string;
  payments: POPayment[];
  onClose: () => void;
  onInvoice: () => void;
  onAction: (action: ActionType) => void;
  onPay: () => void;
  onOpenDoc: (type: 'invoice' | 'receipt' | 'termin_agreement') => void;
  onUpdate?: () => void;
}
function PODetailDrawer({
  po,
  storeId,
  payments,
  onClose,
  onInvoice,
  onAction,
  onPay,
  onOpenDoc,
  onUpdate,
}: PODetailDrawerProps) {
  const payStatus = po.payment_status ?? 'unpaid';
  const amountPaid = po.amount_paid ?? 0;
  const amountDue = po.amount_due ?? po.total_amount - amountPaid;

  return (
    <>
      <div
        onClick={onClose}
        style={{
          position: 'fixed',
          inset: 0,
          background: 'rgba(0,0,0,0.5)',
          backdropFilter: 'blur(2px)',
          zIndex: 200,
        }}
      />
      <div
        style={{
          position: 'fixed',
          top: 0,
          right: 0,
          bottom: 0,
          width: 'min(720px, 100vw)',
          background: 'var(--bg-card)',
          borderLeft: '1px solid var(--border)',
          zIndex: 201,
          overflowY: 'auto',
          display: 'flex',
          flexDirection: 'column',
        }}
      >
        {/* Header */}
        <div
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            padding: '18px 20px',
            borderBottom: '1px solid var(--border)',
            position: 'sticky',
            top: 0,
            background: 'var(--bg-card)',
            zIndex: 1,
          }}
        >
          <div>
            <div style={{ fontWeight: 800, fontSize: '1rem' }}>{po.po_number}</div>
            <div style={{ display: 'flex', gap: 6, marginTop: 4 }}>
              <span className={`badge ${STATUS_BADGE[po.status]}`}>{STATUS_LABEL[po.status]}</span>
              {po.status === 'received' && <PayStatusBadge status={payStatus} />}
            </div>
          </div>
          <div style={{ display: 'flex', gap: 6 }}>
            <button className="btn btn-secondary btn-sm" onClick={onInvoice}>
              <Printer size={13} /> Invoice
            </button>
            <button className="btn btn-ghost btn-sm" onClick={onClose}>
              <X size={16} />
            </button>
          </div>
        </div>

        <div style={{ padding: 20, display: 'flex', flexDirection: 'column', gap: 18, flex: 1 }}>
          {/* Meta */}
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 10 }}>
            {[
              { icon: Hash, label: 'Nomor PO', val: po.po_number },
              { icon: User, label: 'Supplier', val: po.supplier_name ?? '—' },
              { icon: User, label: 'Dibuat Oleh', val: po.ordered_by_name ?? '—' },
              { icon: Calendar, label: 'Tanggal Buat', val: formatDate(po.created_at) },
            ].map(({ icon: Icon, label, val }) => (
              <div
                key={label}
                style={{ background: 'var(--bg-elevated)', borderRadius: 10, padding: '10px 14px' }}
              >
                <div
                  style={{
                    fontSize: '0.68rem',
                    color: 'var(--text-3)',
                    display: 'flex',
                    alignItems: 'center',
                    gap: 4,
                    marginBottom: 4,
                  }}
                >
                  <Icon size={11} /> {label}
                </div>
                <div style={{ fontWeight: 600, fontSize: '0.85rem' }}>{val}</div>
              </div>
            ))}
          </div>

          {/* Items */}
          <div>
            <div style={{ fontWeight: 700, marginBottom: 8, fontSize: '0.9rem' }}>
              Item ({po.items?.length ?? 0})
            </div>
            {(po.items ?? []).map((item, i) => (
              <div
                key={item.id ?? i}
                style={{
                  background: 'var(--bg-elevated)',
                  borderRadius: 10,
                  padding: '11px 14px',
                  display: 'grid',
                  gridTemplateColumns: '1fr auto',
                  gap: 8,
                  marginBottom: 6,
                }}
              >
                <div>
                  <div style={{ fontWeight: 600, fontSize: '0.87rem' }}>{item.product_name}</div>
                  <div style={{ fontSize: '0.74rem', color: 'var(--text-3)', marginTop: 2 }}>
                    SKU: {item.product_sku} · {item.unit}
                  </div>
                </div>
                <div style={{ textAlign: 'right' }}>
                  <div style={{ fontWeight: 700, color: 'var(--accent-em)' }}>
                    {formatRp(item.subtotal)}
                  </div>
                  <div style={{ fontSize: '0.73rem', color: 'var(--text-3)' }}>
                    {item.quantity} × {formatRp(item.unit_cost)}
                  </div>
                </div>
              </div>
            ))}
          </div>

          {/* Financials */}
          <div
            style={{
              background: 'var(--bg-elevated)',
              borderRadius: 12,
              padding: '14px 16px',
              display: 'flex',
              flexDirection: 'column',
              gap: 8,
            }}
          >
            <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.85rem' }}>
              <span style={{ color: 'var(--text-2)' }}>Total PO</span>
              <span style={{ fontWeight: 700 }}>{formatRp(po.total_amount)}</span>
            </div>
            <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.85rem' }}>
              <span style={{ color: '#10b981' }}>Sudah Dibayar</span>
              <span style={{ fontWeight: 700, color: '#10b981' }}>{formatRp(amountPaid)}</span>
            </div>
            <div
              style={{
                borderTop: '1px solid var(--border)',
                paddingTop: 8,
                display: 'flex',
                justifyContent: 'space-between',
                fontSize: '0.9rem',
              }}
            >
              <span style={{ fontWeight: 700 }}>Sisa Hutang</span>
              <span style={{ fontWeight: 800, color: amountDue > 0 ? '#ef4444' : '#10b981' }}>
                {formatRp(amountDue)}
              </span>
            </div>
          </div>

          {/* Payment history */}
          {po.status === 'received' && (
            <div>
              <div style={{ fontWeight: 700, marginBottom: 8, fontSize: '0.9rem' }}>
                Riwayat Pembayaran
              </div>
              {payments.length === 0 ? (
                <div
                  style={{
                    fontSize: '0.82rem',
                    color: 'var(--text-3)',
                    padding: '12px 0',
                    textAlign: 'center',
                  }}
                >
                  Belum ada pembayaran
                </div>
              ) : (
                payments.map(p => (
                  <div
                    key={p.id}
                    style={{
                      background: 'var(--bg-elevated)',
                      borderRadius: 10,
                      padding: '10px 14px',
                      display: 'flex',
                      justifyContent: 'space-between',
                      alignItems: 'center',
                      marginBottom: 6,
                    }}
                  >
                    <div>
                      <div style={{ fontWeight: 700, color: '#10b981', fontSize: '0.9rem' }}>
                        {formatRp(p.amount)}
                      </div>
                      <div style={{ fontSize: '0.73rem', color: 'var(--text-3)', marginTop: 2 }}>
                        {p.paid_by_name} · {formatDate(p.paid_at)}
                        {p.note && ` · ${p.note}`}
                      </div>
                    </div>
                    <CheckCircle2 size={16} style={{ color: '#10b981', flexShrink: 0 }} />
                  </div>
                ))
              )}
            </div>
          )}

          {/* Termin Schedule */}
          <div style={{ borderTop: '1px solid var(--border)', paddingTop: 14 }}>
            <TerminPanel po={po} storeId={storeId} onOpenDoc={onOpenDoc} onUpdate={onUpdate} />
          </div>

          {/* Action buttons */}
          <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
            {po.status === 'received' && payStatus !== 'paid' && (
              <button className="btn btn-primary" style={{ width: '100%' }} onClick={onPay}>
                <Wallet size={14} /> Bayar Hutang {amountDue > 0 ? `(${formatRp(amountDue)})` : ''}
              </button>
            )}
            {(po.status === 'draft' || po.status === 'ordered') && (
              <div style={{ display: 'flex', gap: 8 }}>
                {po.status === 'draft' && (
                  <button
                    className="btn btn-primary"
                    style={{ flex: 1 }}
                    onClick={() => onAction('submit')}
                  >
                    <CheckCircle2 size={14} /> Submit PO
                  </button>
                )}
                {po.status === 'ordered' && (
                  <button
                    className="btn btn-primary"
                    style={{ flex: 1 }}
                    onClick={() => onAction('receive')}
                  >
                    <Package size={14} /> Terima PO
                  </button>
                )}
                <button className="btn btn-danger" onClick={() => onAction('cancel')}>
                  <X size={14} /> Batalkan
                </button>
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
  const [orders, setOrders] = useState<PurchaseOrder[]>([]);
  const [loading, setLoading] = useState(true);
  const [payable, setPayable] = useState<PayableSummary | null>(null);
  const [showModal, setShowModal] = useState(false);
  const [detailPO, setDetailPO] = useState<PurchaseOrder | null>(null);
  const [invoicePO, setInvoicePO] = useState<PurchaseOrder | null>(null);
  const [payingPO, setPayingPO] = useState<PurchaseOrder | null>(null);
  const [storeDetail, setStoreDetail] = useState<Store | null>(null);
  const [expandedRow, setExpandedRow] = useState<string | null>(null);
  const [confirm, setConfirm] = useState<{ po: PurchaseOrder; action: ActionType } | null>(null);
  const [confirmLoading, setConfirmLoading] = useState(false);
  const [payments, setPayments] = useState<POPayment[]>([]);
  const [suppliers, setSuppliers] = useState<Supplier[]>([]);
  const [products, setProducts] = useState<Map<string, Product>>(new Map());
  const [form, setForm] = useState(EMPTY_FORM);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');

  const storeId = selectedStore?.store_id;

  const load = useCallback(() => {
    if (!storeId) return;
    setLoading(true);
    Promise.all([
      purchaseOrdersApi.list(storeId, { per_page: 50 }),
      purchaseOrdersApi.payableSummary(storeId),
    ])
      .then(([posRes, payRes]) => {
        setOrders((posRes.data as PurchaseOrderListResponse).data ?? []);
        setPayable(payRes.data as PayableSummary);
      })
      .catch(console.error)
      .finally(() => setLoading(false));
  }, [storeId]);

  useEffect(() => {
    load();
  }, [load]);

  useEffect(() => {
    const saved = sessionStorage.getItem('openCreatePOWithItems');
    if (saved) {
      try {
        const parsed = JSON.parse(saved) as SavedPOItem[];
        if (Array.isArray(parsed) && parsed.length > 0) {
          // Seed the product cache so the form items display their names correctly
          setProducts(prevProducts => {
            const next = new Map(prevProducts);
            parsed
              .filter(it => it.product_id && it.product_name)
              .forEach(it => {
                if (!next.has(it.product_id)) {
                  next.set(it.product_id, {
                    id: it.product_id,
                    name: it.product_name,
                    sku: it.product_sku || '',
                    unit: it.unit || '',
                    cost_price: it.unit_cost || 0,
                    sell_price: 0,
                    store_id: storeId,
                    category_id: null,
                    tax_rate: 0,
                    is_active: true,
                    created_at: new Date().toISOString(),
                    updated_at: new Date().toISOString(),
                  } as unknown as Product);
                }
              });
            return next;
          });

          setForm(f => ({
            ...f,
            items: parsed.map(it => ({
              product_id: it.product_id,
              quantity: it.quantity,
              unit_cost: it.unit_cost,
              product_name: it.product_name,
              product_sku: it.product_sku,
              unit: it.unit,
            })),
            notes: 'Otomatis dari Stok Menipis',
          }));
          setShowModal(true);
        }
      } catch (e) {
        console.error(e);
      }
      sessionStorage.removeItem('openCreatePOWithItems');
    }
  }, [storeId]);

  useEffect(() => {
    if (!storeId) return;
    Promise.all([suppliersApi.list({ per_page: 100 }), storesApi.get(storeId)]).then(([s, st]) => {
      setSuppliers((s.data as SupplierListResponse).data ?? []);
      setStoreDetail(st.data as Store);
    });
  }, [storeId]);

  useEffect(() => {
    const saved = sessionStorage.getItem('openCreatePOWithItems');
    if (saved) {
      try {
        const parsed = JSON.parse(saved) as SavedPOItem[];
        if (Array.isArray(parsed) && parsed.length > 0) {
          setForm(f => ({
            ...f,
            items: parsed.map(it => ({
              product_id: it.product_id,
              quantity: it.quantity,
              unit_cost: it.unit_cost,
              product_name: it.product_name,
              product_sku: it.product_sku,
              unit: it.unit,
            })),
            notes: 'Otomatis dari Stok Menipis',
          }));
          setShowModal(true);
        }
      } catch (e) {
        console.error(e);
      }
      sessionStorage.removeItem('openCreatePOWithItems');
    }
  }, []);

  const openDetail = useCallback(
    async (po: PurchaseOrder) => {
      if (!storeId) return;
      setDetailPO(po);
      setPayments([]);
      try {
        const [r, pmts] = await Promise.all([
          purchaseOrdersApi.get(storeId, po.id),
          po.status === 'received'
            ? purchaseOrdersApi.listPayments(storeId, po.id)
            : Promise.resolve({ data: [] as POPayment[] }),
        ]);
        setDetailPO(r.data as PurchaseOrder);
        setPayments((pmts.data as POPayment[]) ?? []);
      } catch (e) {
        console.error(e);
      }
    },
    [storeId]
  );

  const openInvoice = useCallback(
    async (po: PurchaseOrder) => {
      if (!storeId) return;
      let fullPO = po;
      if (!po.items?.length) {
        try {
          const r = await purchaseOrdersApi.get(storeId, po.id);
          fullPO = r.data as PurchaseOrder;
        } catch (e) {
          console.error(e);
        }
      }
      setInvoicePO(fullPO);
    },
    [storeId]
  );

  const executeAction = async () => {
    if (!confirm || !storeId) return;
    setConfirmLoading(true);
    const fns: Record<ActionType, () => Promise<unknown>> = {
      submit: () => purchaseOrdersApi.submit(storeId, confirm.po.id),
      receive: () => purchaseOrdersApi.receive(storeId, confirm.po.id),
      cancel: () => purchaseOrdersApi.cancel(storeId, confirm.po.id),
    };
    try {
      await fns[confirm.action]();
      setConfirm(null);
      load();
      if (detailPO?.id === confirm.po.id) openDetail({ ...detailPO });
    } catch (e) {
      alert(e instanceof ApiError ? e.message : 'Gagal');
    } finally {
      setConfirmLoading(false);
    }
  };

  const handlePaySuccess = () => {
    setPayingPO(null);
    load();
    if (detailPO) openDetail(detailPO);
  };

  const addItem = () =>
    setForm(f => ({ ...f, items: [...f.items, { product_id: '', quantity: 1, unit_cost: 0 }] }));
  const removeItem = (i: number) =>
    setForm(f => ({ ...f, items: f.items.filter((_, idx) => idx !== i) }));
  const updateItem = (i: number, update: Partial<ItemRow>) => {
    setForm(f => ({
      ...f,
      items: f.items.map((item, idx) => {
        if (idx !== i) return item;
        const next = { ...item, ...update };
        if (update.product_id) {
          const p = products.get(String(update.product_id));
          if (p) {
            next.unit_cost = p.cost_price > 0 ? p.cost_price : next.unit_cost;
            next.product_name = p.name;
            next.product_sku = p.sku;
            next.unit = p.unit;
          }
        }
        return next;
      }),
    }));
  };
  const calculateTotal = () => form.items.reduce((s, it) => s + it.quantity * it.unit_cost, 0);

  const handleCreate = async () => {
    if (!storeId) return;
    setSaving(true);
    setError('');
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
    } finally {
      setSaving(false);
    }
  };

  if (!selectedStore)
    return (
      <div style={{ padding: 32 }}>
        <div className="empty-state card" style={{ padding: 40 }}>
          <ClipboardList size={40} />
          <p>Pilih toko terlebih dahulu</p>
        </div>
      </div>
    );

  return (
    <div className="w-full p-6">
      {/* Header */}
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'flex-start',
          marginBottom: 20,
        }}
      >
        <div>
          <h1 className="page-title" style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
            <ClipboardList size={22} style={{ color: 'var(--accent-em)' }} />
            Pembelian
          </h1>
          <p className="page-subtitle">{selectedStore.store_name}</p>
        </div>
        <button
          className="btn btn-primary"
          onClick={() => {
            setForm(EMPTY_FORM);
            setError('');
            setShowModal(true);
          }}
        >
          <Plus size={15} /> Buat PO
        </button>
      </div>

      {/* Payable Summary Cards */}
      {payable && (
        <div
          style={{
            display: 'grid',
            gridTemplateColumns: 'repeat(auto-fit,minmax(180px,1fr))',
            gap: 12,
            marginBottom: 20,
          }}
        >
          {[
            {
              label: 'Total Hutang Dagang',
              val: payable.total_debt,
              icon: CreditCard,
              color: 'var(--accent-em)',
            },
            {
              label: 'Sudah Dibayar',
              val: payable.total_paid,
              icon: CheckCircle2,
              color: '#10b981',
            },
            {
              label: 'Sisa Hutang',
              val: payable.total_outstanding,
              icon: AlertCircle,
              color: '#ef4444',
            },
          ].map(({ label, val, icon: Icon, color }) => (
            <div key={label} className="card" style={{ padding: '14px 16px' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 8 }}>
                <Icon size={16} style={{ color }} />
                <span style={{ fontSize: '0.72rem', color: 'var(--text-3)', fontWeight: 600 }}>
                  {label}
                </span>
              </div>
              <div style={{ fontWeight: 800, fontSize: '1.1rem', color }}>{formatRp(val)}</div>
            </div>
          ))}
          <div className="card" style={{ padding: '14px 16px' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 8 }}>
              <Clock size={16} style={{ color: '#f59e0b' }} />
              <span style={{ fontSize: '0.72rem', color: 'var(--text-3)', fontWeight: 600 }}>
                Status Hutang
              </span>
            </div>
            <div style={{ display: 'flex', gap: 8, fontSize: '0.82rem' }}>
              <span style={{ color: '#ef4444', fontWeight: 700 }}>
                {payable.unpaid_count} belum bayar
              </span>
              <span style={{ color: 'var(--text-3)' }}>·</span>
              <span style={{ color: '#f59e0b', fontWeight: 700 }}>
                {payable.partial_count} sebagian
              </span>
            </div>
          </div>
        </div>
      )}

      {/* List */}
      <div className="card" style={{ overflow: 'hidden' }}>
        {loading ? (
          <div style={{ display: 'flex', justifyContent: 'center', padding: 40 }}>
            <Loader2 size={24} className="loading-spin" style={{ color: 'var(--accent-em)' }} />
          </div>
        ) : orders.length === 0 ? (
          <div className="empty-state">
            <ClipboardList size={32} />
            <p>Belum ada purchase order</p>
          </div>
        ) : (
          <table className="tbl">
            <thead>
              <tr>
                <th style={{ width: 40 }} />
                <th>Nomor PO</th>
                <th>Supplier</th>
                <th>Items</th>
                <th>Total</th>
                <th>Status</th>
                <th>Hutang</th>
                <th>Deadline</th>
                <th>Dibuat</th>
                <th>Dibuat Oleh</th>
                <th>Aksi</th>
              </tr>
            </thead>
            <tbody>
              {orders.map(po => {
                const ps = po.payment_status ?? 'unpaid';
                const isExpanded = expandedRow === po.id;

                return (
                  <Fragment key={po.id}>
                    <tr
                      style={{
                        background: isExpanded ? 'var(--bg-elevated)' : 'transparent',
                      }}
                    >
                      <td
                        onClick={() => setExpandedRow(isExpanded ? null : po.id)}
                        style={{ textAlign: 'center', cursor: 'pointer' }}
                      >
                        {isExpanded ? <ChevronDown size={16} /> : <ChevronRight size={16} />}
                      </td>
                      <td
                        style={{
                          fontFamily: 'monospace',
                          fontWeight: 700,
                          color: 'var(--accent-em)',
                        }}
                      >
                        {po.po_number}
                      </td>
                      <td style={{ color: 'var(--text-2)' }}>{po.supplier_name ?? '—'}</td>
                      <td>
                        <span
                          style={{
                            background: 'var(--bg-elevated)',
                            borderRadius: 6,
                            padding: '2px 8px',
                            fontSize: '0.78rem',
                            color: 'var(--text-2)',
                          }}
                        >
                          {po.total_items ?? po.items?.length ?? 0} item
                        </span>
                      </td>
                      <td style={{ fontWeight: 700, color: 'var(--accent-em)' }}>
                        {formatRp(po.total_amount)}
                      </td>
                      <td>
                        <span className={`badge ${STATUS_BADGE[po.status]}`}>
                          {STATUS_LABEL[po.status]}
                        </span>
                      </td>
                      <td>
                        {po.status === 'received' ? (
                          <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
                            <PayStatusBadge status={ps} />
                            {(po.amount_due ?? 0) > 0 && (
                              <span
                                style={{ fontSize: '0.75rem', fontWeight: 600, color: '#ef4444' }}
                              >
                                Sisa: {formatRp(po.amount_due ?? 0)}
                              </span>
                            )}
                          </div>
                        ) : (
                          <span style={{ color: 'var(--text-3)', fontSize: '0.78rem' }}>—</span>
                        )}
                      </td>
                      <td>
                        {po.next_deadline ? (
                          <span style={{ fontSize: '0.8rem', fontWeight: 600, color: '#dc2626' }}>
                            {formatDate(po.next_deadline)}
                          </span>
                        ) : (
                          <span style={{ color: 'var(--text-3)', fontSize: '0.78rem' }}>—</span>
                        )}
                      </td>
                      <td style={{ color: 'var(--text-3)', fontSize: '0.8rem' }}>
                        {formatDate(po.created_at)}
                      </td>
                      <td style={{ color: 'var(--text-3)', fontSize: '0.8rem' }}>
                        {po.ordered_by_name || '—'}
                      </td>
                      <td>
                        <div style={{ display: 'flex', gap: 4 }}>
                          <button
                            className="btn btn-ghost btn-sm"
                            onClick={e => {
                              e.stopPropagation();
                              openDetail(po);
                            }}
                            title="Lihat detail"
                            aria-label={`Lihat detail pembelian ${po.po_number}`}
                          >
                            <Eye size={13} />
                          </button>
                          {po.status === 'draft' && (
                            <button
                              className="btn btn-secondary btn-sm"
                              onClick={e => {
                                e.stopPropagation();
                                setConfirm({ po, action: 'submit' });
                              }}
                            >
                              Submit
                            </button>
                          )}
                          {po.status === 'ordered' && (
                            <button
                              className="btn btn-primary btn-sm"
                              onClick={e => {
                                e.stopPropagation();
                                setConfirm({ po, action: 'receive' });
                              }}
                            >
                              Terima
                            </button>
                          )}
                          {po.status === 'received' && ps !== 'paid' && (
                            <button
                              className="btn btn-secondary btn-sm"
                              style={{ color: '#10b981' }}
                              onClick={e => {
                                e.stopPropagation();
                                setPayingPO(po);
                              }}
                            >
                              <Wallet size={12} /> Bayar
                            </button>
                          )}
                          {(po.status === 'draft' || po.status === 'ordered') && (
                            <button
                              className="btn btn-danger btn-sm"
                              onClick={e => {
                                e.stopPropagation();
                                setConfirm({ po, action: 'cancel' });
                              }}
                            >
                              Batal
                            </button>
                          )}
                          <button
                            className="btn btn-ghost btn-sm"
                            onClick={e => {
                              e.stopPropagation();
                              openInvoice(po);
                            }}
                          >
                            <Printer size={13} />
                          </button>
                        </div>
                      </td>
                    </tr>
                    {isExpanded && (
                      <tr>
                        <td
                          colSpan={10}
                          style={{ padding: 0, borderBottom: '2px solid var(--accent-em)' }}
                        >
                          <div style={{ background: 'var(--bg-card)' }}>
                            <TerminPanel
                              po={po}
                              storeId={storeId ?? ''}
                              onOpenDoc={type =>
                                window.open(
                                  `/purchase-orders/${po.id}/document?type=${type}`,
                                  '_blank'
                                )
                              }
                              onUpdate={load}
                            />
                          </div>
                        </td>
                      </tr>
                    )}
                  </Fragment>
                );
              })}
            </tbody>
          </table>
        )}
      </div>

      {/* Overlays */}
      {detailPO && (
        <PODetailDrawer
          po={detailPO}
          storeId={storeId ?? ''}
          payments={payments}
          onClose={() => setDetailPO(null)}
          onInvoice={() => openInvoice(detailPO)}
          onAction={a => setConfirm({ po: detailPO, action: a })}
          onPay={() => setPayingPO(detailPO)}
          onOpenDoc={type =>
            window.open(`/purchase-orders/${detailPO.id}/document?type=${type}`, '_blank')
          }
          onUpdate={() => {
            load();
            openDetail(detailPO);
          }}
        />
      )}
      {invoicePO && (
        <InvoiceModal po={invoicePO} store={storeDetail} onClose={() => setInvoicePO(null)} />
      )}
      {payingPO && (
        <PayModal
          po={payingPO}
          storeId={storeId ?? ''}
          onSuccess={handlePaySuccess}
          onCancel={() => setPayingPO(null)}
        />
      )}
      {confirm && (
        <ConfirmModal
          action={confirm.action}
          po={confirm.po}
          onConfirm={executeAction}
          onCancel={() => setConfirm(null)}
          loading={confirmLoading}
        />
      )}

      {/* ── Create / Edit PO Sidebar ─────────────────────────────────────────── */}
      {showModal && (
        <div
          style={{
            position: 'fixed',
            inset: 0,
            zIndex: 1000,
            display: 'flex',
            justifyContent: 'flex-end',
          }}
          onClick={() => setShowModal(false)}
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
              maxWidth: 620,
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
              <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                <ClipboardList size={22} style={{ color: 'var(--accent-em)' }} />
                <h2 style={{ fontWeight: 800, fontSize: '1.2rem', margin: 0 }}>Buat Pembelian</h2>
              </div>
              <button
                onClick={() => setShowModal(false)}
                className="btn btn-ghost btn-sm"
                style={{ padding: 6 }}
              >
                <X size={20} />
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

              <div style={{ display: 'flex', flexDirection: 'column', gap: 24 }}>
                {/* Supplier Selection */}
                <div className="input-group">
                  <label
                    className="input-label"
                    style={{ display: 'flex', alignItems: 'center', gap: 6 }}
                  >
                    <User size={14} style={{ color: 'var(--text-3)' }} /> Supplier
                  </label>
                  <SearchableSelect
                    placeholder="Pilih Supplier..."
                    options={suppliers.map(s => ({ value: s.id, label: s.name }))}
                    value={form.supplier_id}
                    onChange={v => setForm(f => ({ ...f, supplier_id: v }))}
                  />
                </div>

                {/* Items Section */}
                <div>
                  <div
                    style={{
                      display: 'flex',
                      justifyContent: 'space-between',
                      alignItems: 'center',
                      marginBottom: 12,
                    }}
                  >
                    <label
                      className="input-label"
                      style={{ margin: 0, display: 'flex', alignItems: 'center', gap: 6 }}
                    >
                      <Package size={14} style={{ color: 'var(--text-3)' }} /> Item Pembelian
                    </label>
                    <span
                      style={{
                        fontSize: '0.75rem',
                        fontWeight: 600,
                        color: 'var(--accent-em)',
                        background: 'var(--bg-elevated)',
                        padding: '2px 8px',
                        borderRadius: 6,
                      }}
                    >
                      {form.items.length} Item
                    </span>
                  </div>

                  <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
                    {form.items.map((item, i) => {
                      const lineTotal = item.quantity * item.unit_cost;
                      return (
                        <div
                          key={i}
                          style={{
                            background: 'var(--bg-elevated)',
                            borderRadius: 12,
                            padding: '12px 14px',
                            border: '1px solid var(--border)',
                            display: 'flex',
                            flexDirection: 'column',
                            gap: 12,
                            position: 'relative',
                            transition: 'all 0.2s ease',
                          }}
                        >
                          {/* Line 1: Product, Qty, Price & Delete */}
                          <div style={{ display: 'flex', gap: 10, alignItems: 'flex-end' }}>
                            <div
                              style={{ flex: 1, display: 'flex', flexDirection: 'column', gap: 3 }}
                            >
                              <label
                                style={{
                                  fontSize: '0.6rem',
                                  color: 'var(--text-3)',
                                  fontWeight: 700,
                                  textTransform: 'uppercase',
                                  letterSpacing: '0.025em',
                                }}
                              >
                                Produk
                              </label>
                              <ProductSearchSelect
                                storeId={storeId ?? ''}
                                value={item.product_id}
                                selectedName={
                                  item.product_name ||
                                  (item.product_id
                                    ? products.get(item.product_id)?.name
                                    : undefined)
                                }
                                onSelect={p => {
                                  setProducts(prev => new Map(prev).set(p.id, p));
                                  updateItem(i, {
                                    product_id: p.id,
                                    product_name: p.name,
                                    product_sku: p.sku,
                                    unit: p.unit,
                                    unit_cost: p.cost_price || 0,
                                  });
                                }}
                              />
                            </div>
                            <div
                              style={{
                                width: 64,
                                display: 'flex',
                                flexDirection: 'column',
                                gap: 3,
                              }}
                            >
                              <label
                                style={{
                                  fontSize: '0.6rem',
                                  color: 'var(--text-3)',
                                  fontWeight: 700,
                                  textTransform: 'uppercase',
                                  letterSpacing: '0.025em',
                                }}
                              >
                                Jumlah
                              </label>
                              <input
                                type="number"
                                className="input"
                                style={{ height: 32, fontSize: '0.8rem', padding: '0 8px' }}
                                placeholder="Qty"
                                min={1}
                                value={item.quantity}
                                onFocus={e => e.target.select()}
                                onChange={e => updateItem(i, { quantity: +e.target.value })}
                              />
                            </div>
                            <div
                              style={{
                                paddingBottom: 10,
                                fontSize: '0.9rem',
                                color: 'var(--text-3)',
                                opacity: 0.6,
                              }}
                            >
                              ×
                            </div>
                            <div
                              style={{
                                width: 120,
                                display: 'flex',
                                flexDirection: 'column',
                                gap: 3,
                              }}
                            >
                              <label
                                style={{
                                  fontSize: '0.6rem',
                                  color: 'var(--text-3)',
                                  fontWeight: 700,
                                  textTransform: 'uppercase',
                                  letterSpacing: '0.025em',
                                }}
                              >
                                Harga / Item
                              </label>
                              <input
                                type="text"
                                className="input"
                                style={{ height: 32, fontSize: '0.8rem', padding: '0 8px' }}
                                placeholder="Harga"
                                onFocus={e => e.target.select()}
                                value={formatNumberInput(String(item.unit_cost))}
                                onChange={e =>
                                  updateItem(i, { unit_cost: parseNumberInput(e.target.value) })
                                }
                              />
                            </div>
                            <div style={{ paddingBottom: 2 }}>
                              <button
                                className="btn btn-ghost btn-sm"
                                style={{
                                  color: '#ef4444',
                                  padding: 4,
                                  height: 32,
                                  width: 32,
                                  borderRadius: 8,
                                }}
                                onClick={() => removeItem(i)}
                                disabled={form.items.length === 1}
                              >
                                <Trash2 size={15} />
                              </button>
                            </div>
                          </div>

                          {/* Line 2: SKU, Unit & Total Price */}
                          <div
                            style={{
                              display: 'flex',
                              justifyContent: 'space-between',
                              alignItems: 'center',
                              paddingTop: 8,
                              borderTop: '1px dashed var(--border)',
                              marginTop: 0,
                            }}
                          >
                            <div
                              style={{
                                fontSize: '0.7rem',
                                color: 'var(--text-3)',
                                fontWeight: 500,
                                display: 'flex',
                                alignItems: 'center',
                                gap: 6,
                              }}
                            >
                              <span style={{ opacity: 0.7 }}>{item.product_sku || '—'}</span>
                              <span style={{ opacity: 0.3 }}>·</span>
                              <span
                                style={{
                                  background: 'var(--bg-card)',
                                  padding: '1px 5px',
                                  borderRadius: 4,
                                  fontSize: '0.65rem',
                                }}
                              >
                                {item.unit || 'pcs'}
                              </span>
                            </div>
                            <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                              <span
                                style={{
                                  fontSize: '0.6rem',
                                  color: 'var(--text-3)',
                                  textTransform: 'uppercase',
                                  fontWeight: 700,
                                  letterSpacing: '0.05em',
                                }}
                              >
                                Total
                              </span>
                              <span
                                style={{
                                  fontWeight: 800,
                                  fontSize: '1.05rem',
                                  color: 'var(--accent-em)',
                                }}
                              >
                                {formatRp(lineTotal)}
                              </span>
                            </div>
                          </div>
                        </div>
                      );
                    })}
                  </div>

                  <button
                    className="btn btn-ghost"
                    style={{
                      width: '100%',
                      marginTop: 12,
                      border: '1px dashed var(--border)',
                      padding: '12px',
                      fontSize: '0.85rem',
                      borderRadius: 12,
                    }}
                    onClick={addItem}
                  >
                    <Plus size={14} /> Tambah Item Lain
                  </button>
                </div>

                {/* Notes Section */}
                <div className="input-group">
                  <label className="input-label">Catatan</label>
                  <textarea
                    className="input"
                    rows={3}
                    placeholder="Contoh: Pembayaran termin 30 hari..."
                    style={{ resize: 'none', padding: '10px 12px' }}
                    value={form.notes}
                    onChange={e => setForm(f => ({ ...f, notes: e.target.value }))}
                  />
                </div>

                {/* Total Summary Button */}
                <button
                  disabled={saving || calculateTotal() === 0}
                  onClick={handleCreate}
                  style={{
                    width: '100%',
                    background: 'var(--accent-em)',
                    backgroundImage: 'linear-gradient(135deg, var(--accent-em) 0%, #3b82f6 100%)',
                    borderRadius: 16,
                    padding: '20px',
                    color: '#fff',
                    marginTop: 8,
                    textAlign: 'left',
                    boxShadow: '0 8px 24px rgba(59,130,246,0.25)',
                    border: 'none',
                    cursor: saving || calculateTotal() === 0 ? 'not-allowed' : 'pointer',
                    transition: 'all 0.2s ease',
                    opacity: saving || calculateTotal() === 0 ? 0.7 : 1,
                    display: 'flex',
                    flexDirection: 'column',
                    gap: 4,
                  }}
                  onMouseEnter={e => {
                    if (!saving && calculateTotal() > 0) {
                      e.currentTarget.style.transform = 'translateY(-2px)';
                      e.currentTarget.style.boxShadow = '0 12px 32px rgba(59,130,246,0.35)';
                    }
                  }}
                  onMouseLeave={e => {
                    e.currentTarget.style.transform = 'translateY(0)';
                    e.currentTarget.style.boxShadow = '0 8px 24px rgba(59,130,246,0.25)';
                  }}
                >
                  <div
                    style={{
                      display: 'flex',
                      justifyContent: 'space-between',
                      alignItems: 'center',
                      width: '100%',
                    }}
                  >
                    <div style={{ opacity: 0.9, fontSize: '0.85rem', fontWeight: 600 }}>
                      Total Pembelian
                    </div>
                    {saving ? <Loader2 size={18} className="loading-spin" /> : <Check size={20} />}
                  </div>
                  <div style={{ fontSize: '1.45rem', fontWeight: 800 }}>
                    {formatRp(calculateTotal())}
                  </div>
                  <div style={{ fontSize: '0.75rem', opacity: 0.8, fontWeight: 500 }}>
                    Klik untuk Simpan Pembelian
                  </div>
                </button>
              </div>
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
              <button
                className="btn btn-ghost"
                onClick={() => setShowModal(false)}
                style={{ minWidth: 100 }}
              >
                Batal
              </button>
              <button
                className="btn btn-primary"
                disabled={saving}
                style={{ minWidth: 160, gap: 8 }}
                onClick={handleCreate}
              >
                {saving ? <Loader2 size={16} className="loading-spin" /> : <Check size={16} />}
                {saving ? 'Menyimpan...' : 'Simpan Pembelian'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
