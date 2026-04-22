'use client';

import { useEffect, useReducer, useState, useCallback, useRef } from 'react';
import {
  Search,
  ShoppingCart,
  Trash2,
  Printer,
  X,
  Minus,
  Plus,
  Loader2,
  CheckCircle2,
  UtensilsCrossed,
  ShoppingBag,
  UserRound,
  ArrowLeft,
  Clock,
  Users,
  Tag,
  ChevronDown,
  Star,
} from 'lucide-react';
import { useAuth } from '@/lib/auth/AuthContext';
import { useOfflineTransaction } from '@/hooks/useOfflineTransaction';
import { useLoyalty } from '@/hooks/useLoyalty';
import { productsApi } from '@/lib/api/products';
import { menuItemsApi, customersApi, tablesApi } from '@/lib/api/store-apis';
import { transactionsApi } from '@/lib/api/transactions';
import Portal from '@/components/ui/Portal';
import { formatRp, productEmoji } from '@/lib/utils';
import type {
  Product,
  Category,
  CartItem,
  Transaction,
  MenuItem,
  Customer,
  RestaurantTable,
  DiscountType,
  TransactionItem,
  PaginatedData,
} from '@/types';
import { ApiError } from '@/lib/api/client';

// ── Unified cart item type ────────────────────────────────────────────────────
// We re-use CartItem but may hold either a product or a menu item.
// menuItemId is set for restaurant items.
interface PosCartItem extends CartItem {
  menuItemId?: string;
}

interface AggregatedDraftItem extends TransactionItem {
  parsedId: string;
  cost_price?: number;
}

type CartAction =
  | { type: 'ADD_PRODUCT'; product: Product }
  | { type: 'ADD_MENU'; item: MenuItem }
  | { type: 'REMOVE'; id: string }
  | { type: 'SET_QTY'; id: string; qty: number }
  | {
      type: 'SET_DISCOUNT';
      id: string;
      discountType: DiscountType;
      discountValue: number;
    }
  | { type: 'CLEAR' };

/** Compute final unit price + line totals from discount inputs. */
function applyDiscount(
  originalPrice: number,
  qty: number,
  taxRate: number,
  discountType: DiscountType,
  discountValue: number
): { unitPrice: number; subtotal: number; taxAmt: number } {
  let finalPrice: number;
  switch (discountType) {
    case 'FIXED':
      finalPrice = Math.max(0, originalPrice - discountValue);
      break;
    case 'OVERRIDE':
      finalPrice = Math.max(0, discountValue);
      break;
    default: // PERCENTAGE
      finalPrice = originalPrice * (1 - discountValue / 100);
  }
  if (finalPrice < 0) finalPrice = 0;
  const subtotal = finalPrice * qty;
  const taxAmt = subtotal * (taxRate / 100);
  return { unitPrice: finalPrice, subtotal, taxAmt };
}

function makeFromProduct(product: Product, qty = 1): PosCartItem {
  const { unitPrice, subtotal, taxAmt } = applyDiscount(
    product.sell_price,
    qty,
    product.tax_rate,
    'PERCENTAGE',
    0
  );
  return {
    product,
    quantity: qty,
    discount_pct: 0,
    discountType: 'PERCENTAGE',
    discountValue: 0,
    originalPrice: product.sell_price,
    unitPrice,
    subtotal,
    taxAmt,
  };
}

function makeFromMenu(item: MenuItem, qty = 1): PosCartItem {
  const taxRate = item.tax_rate ?? 0;
  const { unitPrice, subtotal, taxAmt } = applyDiscount(
    item.sell_price,
    qty,
    taxRate,
    'PERCENTAGE',
    0
  );
  const fakeProduct: Product = {
    id: item.id,
    name: item.name,
    sku: 'MENU',
    unit: 'porsi',
    description: '',
    sell_price: item.sell_price,
    cost_price: item.cost_price ?? 0,
    tax_rate: taxRate,
    is_active: true,
    store_id: item.store_id,
    category_id: item.category_id ?? '',
    barcode: '',
    created_at: '',
    updated_at: '',
  };
  return {
    product: fakeProduct,
    quantity: qty,
    discount_pct: 0,
    discountType: 'PERCENTAGE',
    discountValue: 0,
    originalPrice: item.sell_price,
    unitPrice,
    subtotal,
    taxAmt,
    menuItemId: item.id,
  };
}

function cartReducer(state: PosCartItem[], action: CartAction): PosCartItem[] {
  switch (action.type) {
    case 'ADD_PRODUCT': {
      const idx = state.findIndex(i => i.product.id === action.product.id && !i.menuItemId);
      if (idx >= 0)
        return state.map((item, i) =>
          i === idx ? makeFromProduct(item.product, item.quantity + 1) : item
        );
      return [...state, makeFromProduct(action.product)];
    }
    case 'ADD_MENU': {
      const idx = state.findIndex(i => i.menuItemId === action.item.id);
      if (idx >= 0)
        return state.map((item, i) =>
          i === idx ? makeFromMenu(action.item, item.quantity + 1) : item
        );
      return [...state, makeFromMenu(action.item)];
    }
    case 'REMOVE':
      return state.filter(i => i.product.id !== action.id);
    case 'SET_QTY': {
      if (action.qty < 1) return state.filter(i => i.product.id !== action.id);
      return state.map(i => {
        if (i.product.id !== action.id) return i;
        const { unitPrice, subtotal, taxAmt } = applyDiscount(
          i.originalPrice,
          action.qty,
          i.product.tax_rate,
          i.discountType,
          i.discountValue
        );
        return { ...i, quantity: action.qty, unitPrice, subtotal, taxAmt };
      });
    }
    case 'SET_DISCOUNT': {
      return state.map(i => {
        if (i.product.id !== action.id) return i;
        const { unitPrice, subtotal, taxAmt } = applyDiscount(
          i.originalPrice,
          i.quantity,
          i.product.tax_rate,
          action.discountType,
          action.discountValue
        );
        return {
          ...i,
          discountType: action.discountType,
          discountValue: action.discountValue,
          discount_pct: action.discountType === 'PERCENTAGE' ? action.discountValue : 0,
          unitPrice,
          subtotal,
          taxAmt,
        };
      });
    }
    case 'CLEAR':
      return [];
    default:
      return state;
  }
}

// ── Payment Modal ─────────────────────────────────────────────────────────────
function PaymentModal({
  total,
  onClose,
  onConfirm,
  loading,
}: {
  total: number;
  onClose: () => void;
  onConfirm: (method: string, amount: number) => void;
  loading: boolean;
}) {
  const methods = ['cash', 'qris', 'card', 'transfer'];
  const [method, setMethod] = useState('cash');
  const [amountStr, setAmountStr] = useState('');

  const amount = parseFloat(amountStr) || 0;
  const change = amount - total;
  const canConfirm = method !== 'cash' || amount >= total;

  const numPress = (val: string) => {
    if (val === '⌫') {
      setAmountStr(s => s.slice(0, -1));
      return;
    }
    if (val === '000') {
      setAmountStr(s => (s + '000').replace(/^0+(\d)/, '$1'));
      return;
    }
    setAmountStr(s => (s + val).replace(/^0+(\d)/, '$1'));
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-box" onClick={e => e.stopPropagation()}>
        <div
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            marginBottom: 20,
          }}
        >
          <h2 style={{ fontWeight: 700, fontSize: '1.1rem' }}>Proses Pembayaran</h2>
          <button onClick={onClose} className="btn btn-ghost btn-sm">
            <X size={16} />
          </button>
        </div>

        <div
          style={{
            background: 'rgba(16,185,129,0.1)',
            borderRadius: 10,
            padding: '14px 18px',
            marginBottom: 18,
            textAlign: 'center',
          }}
        >
          <div style={{ fontSize: '0.8rem', color: 'var(--text-2)', marginBottom: 2 }}>
            Total Pembayaran
          </div>
          <div style={{ fontSize: '1.8rem', fontWeight: 800, color: 'var(--accent-em)' }}>
            {formatRp(total)}
          </div>
        </div>

        <div style={{ marginBottom: 16 }}>
          <div
            style={{
              fontSize: '0.78rem',
              color: 'var(--text-3)',
              marginBottom: 8,
              textTransform: 'uppercase',
              letterSpacing: '0.05em',
            }}
          >
            Metode Bayar
          </div>
          <div className="pay-method-tabs">
            {methods.map(m => (
              <button
                key={m}
                className={`pay-tab ${method === m ? 'active' : ''}`}
                onClick={() => {
                  setMethod(m);
                  if (m !== 'cash') setAmountStr(String(total));
                }}
              >
                {m.toUpperCase()}
              </button>
            ))}
          </div>
        </div>

        {method === 'cash' && (
          <>
            <div style={{ marginBottom: 8 }}>
              <div style={{ fontSize: '0.78rem', color: 'var(--text-3)', marginBottom: 6 }}>
                Jumlah Diterima
              </div>
              <div
                style={{
                  background: 'var(--bg-elevated)',
                  borderRadius: 8,
                  padding: '10px 14px',
                  fontSize: '1.4rem',
                  fontWeight: 700,
                  color: 'var(--text-1)',
                  minHeight: 50,
                  display: 'flex',
                  alignItems: 'center',
                  border: '1px solid var(--border-md)',
                }}
              >
                {amountStr ? (
                  formatRp(amount)
                ) : (
                  <span style={{ color: 'var(--text-3)' }}>Masukkan jumlah</span>
                )}
              </div>
            </div>

            <div className="numpad">
              {['7', '8', '9', '4', '5', '6', '1', '2', '3', '000', '0', '⌫'].map(k => (
                <button
                  key={k}
                  className="num-btn"
                  onClick={() => numPress(k)}
                  style={k === '⌫' ? { color: 'var(--accent-rd)' } : undefined}
                >
                  {k}
                </button>
              ))}
            </div>

            {amount > 0 && (
              <div
                style={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  marginTop: 12,
                  padding: '10px 14px',
                  background: change >= 0 ? 'rgba(16,185,129,0.08)' : 'rgba(239,68,68,0.08)',
                  borderRadius: 8,
                }}
              >
                <span style={{ color: 'var(--text-2)', fontSize: '0.9rem' }}>Kembalian</span>
                <span
                  style={{
                    fontWeight: 700,
                    color: change >= 0 ? 'var(--accent-em)' : 'var(--accent-rd)',
                    fontSize: '0.9rem',
                  }}
                >
                  {change >= 0 ? formatRp(change) : '⚠ Kurang ' + formatRp(-change)}
                </span>
              </div>
            )}
          </>
        )}

        {method !== 'cash' && (
          <div
            style={{
              padding: '20px',
              textAlign: 'center',
              color: 'var(--text-2)',
              fontSize: '0.9rem',
            }}
          >
            Konfirmasi pembayaran {method.toUpperCase()} sebesar{' '}
            <strong style={{ color: 'var(--text-1)' }}>{formatRp(total)}</strong>
          </div>
        )}

        <button
          className="btn btn-primary btn-lg"
          style={{ width: '100%', marginTop: 16 }}
          disabled={!canConfirm || loading}
          onClick={() => onConfirm(method, method === 'cash' ? amount : total)}
        >
          {loading ? <Loader2 size={18} className="loading-spin" /> : <CheckCircle2 size={18} />}
          {loading ? 'Memproses...' : 'Konfirmasi Pembayaran'}
        </button>
      </div>
    </div>
  );
}

// ── Receipt Modal ─────────────────────────────────────────────────────────────
function ReceiptModal({ txn, onClose }: { txn: Transaction; onClose: () => void }) {
  const handlePrint = () => {
    const lines = txn.items
      .map(
        item => `
      <div class="item">
        <div class="item-name">${item.product_name}</div>
        <div class="item-row">
          <span>${item.quantity} x ${formatRp(item.unit_price)}</span>
          <span>${formatRp(item.subtotal)}</span>
        </div>
      </div>`
      )
      .join('');

    const html = `<!DOCTYPE html>
<html lang="id">
<head>
  <meta charset="UTF-8" />
  <title>Struk - ${txn.id.slice(0, 8).toUpperCase()}</title>
  <style>
    * { margin: 0; padding: 0; box-sizing: border-box; }
    body { font-family: 'Courier New', Courier, monospace; font-size: 12px; color: #000; background: #fff; width: 80mm; padding: 8px; }
    .center { text-align: center; }
    .bold { font-weight: 700; }
    .divider { border-top: 1px dashed #999; margin: 8px 0; }
    .row { display: flex; justify-content: space-between; margin: 2px 0; }
    .total-row { display: flex; justify-content: space-between; font-weight: 800; font-size: 14px; margin-top: 4px; }
    .item { margin-bottom: 6px; }
    .item-name { font-weight: 600; }
    .item-row { display: flex; justify-content: space-between; font-size: 11px; color: #444; }
    .muted { color: #555; font-size: 11px; }
    @page { size: 80mm auto; margin: 0; }
  </style>
</head>
<body>
  <div class="center bold" style="font-size:14px;margin-bottom:4px;">MoedahPOS</div>
  <div class="center muted">${new Date(txn.created_at).toLocaleString('id-ID')}</div>
  <div class="center muted">Kasir: ${txn.cashier_name ?? '-'}</div>
  ${txn.customer_name ? `<div class="center muted">Pelanggan: ${txn.customer_name}</div>` : ''}
  <div class="divider"></div>
  ${lines}
  <div class="divider"></div>
  <div class="row muted"><span>Subtotal</span><span>${formatRp(txn.subtotal)}</span></div>
  ${txn.discount_amt > 0 ? `<div class="row muted"><span>Diskon</span><span>-${formatRp(txn.discount_amt)}</span></div>` : ''}
  <div class="row muted"><span>PPN</span><span>${formatRp(txn.tax_amt)}</span></div>
  <div class="total-row"><span>TOTAL</span><span>${formatRp(txn.total)}</span></div>
  <div class="row muted" style="margin-top:4px;"><span>Bayar (${txn.payment_method.toUpperCase()})</span><span>${formatRp(txn.payment_amount)}</span></div>
  ${txn.change_amount > 0 ? `<div class="row muted"><span>Kembalian</span><span>${formatRp(txn.change_amount)}</span></div>` : ''}
  <div class="divider"></div>
  <div class="center muted">Terima kasih telah berbelanja!<br/>No. Transaksi: ${txn.id.slice(0, 8).toUpperCase()}</div>
</body>
</html>`;

    const pw = window.open('', '_blank', 'width=360,height=600');
    if (!pw) return;
    pw.document.write(html);
    pw.document.close();
    pw.onload = () => {
      pw.focus();
      pw.print();
      pw.onafterprint = () => pw.close();
    };
    setTimeout(() => {
      try {
        pw.focus();
        pw.print();
      } catch {
        /* already handled */
      }
    }, 500);
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-box" style={{ maxWidth: 340 }} onClick={e => e.stopPropagation()}>
        <div
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            marginBottom: 16,
          }}
        >
          <h2 style={{ fontWeight: 700, fontSize: '1rem' }}>Struk Pembayaran</h2>
          <button onClick={onClose} className="btn btn-ghost btn-sm">
            <X size={16} />
          </button>
        </div>

        {/* Screen preview — CSS vars are fine here */}
        <div
          style={{ fontFamily: "'Courier New', monospace", fontSize: '0.82rem', lineHeight: 1.6 }}
        >
          <div
            style={{
              textAlign: 'center',
              marginBottom: 12,
              borderBottom: '1px dashed var(--border-md)',
              paddingBottom: 10,
            }}
          >
            <div style={{ fontWeight: 800, fontSize: '1rem' }}>MoedahPOS</div>
            <div style={{ color: 'var(--text-2)', fontSize: '0.75rem' }}>
              {new Date(txn.created_at).toLocaleString('id-ID')}
            </div>
            <div style={{ color: 'var(--text-2)', fontSize: '0.75rem' }}>
              Kasir: {txn.cashier_name}
            </div>
            {txn.customer_name && (
              <div style={{ color: 'var(--text-2)', fontSize: '0.75rem' }}>
                Pelanggan: {txn.customer_name}
              </div>
            )}
          </div>

          {txn.items.map((item, i) => (
            <div key={i} style={{ marginBottom: 6 }}>
              <div style={{ fontWeight: 600 }}>{item.product_name}</div>
              <div
                style={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  color: 'var(--text-2)',
                  fontSize: '0.78rem',
                }}
              >
                <span>
                  {item.quantity} x {formatRp(item.unit_price)}
                </span>
                <span>{formatRp(item.subtotal)}</span>
              </div>
            </div>
          ))}

          <div style={{ borderTop: '1px dashed var(--border-md)', marginTop: 10, paddingTop: 10 }}>
            <div
              style={{
                display: 'flex',
                justifyContent: 'space-between',
                color: 'var(--text-2)',
                fontSize: '0.8rem',
              }}
            >
              <span>Subtotal</span>
              <span>{formatRp(txn.subtotal)}</span>
            </div>
            {txn.discount_amt > 0 && (
              <div
                style={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  color: 'var(--accent-rd)',
                  fontSize: '0.8rem',
                }}
              >
                <span>Diskon</span>
                <span>-{formatRp(txn.discount_amt)}</span>
              </div>
            )}
            <div
              style={{
                display: 'flex',
                justifyContent: 'space-between',
                color: 'var(--text-2)',
                fontSize: '0.8rem',
              }}
            >
              <span>PPN</span>
              <span>{formatRp(txn.tax_amt)}</span>
            </div>
            <div
              style={{
                display: 'flex',
                justifyContent: 'space-between',
                fontWeight: 800,
                fontSize: '1rem',
                marginTop: 4,
                color: 'var(--accent-em)',
              }}
            >
              <span>TOTAL</span>
              <span>{formatRp(txn.total)}</span>
            </div>
            <div
              style={{
                display: 'flex',
                justifyContent: 'space-between',
                color: 'var(--text-2)',
                fontSize: '0.8rem',
                marginTop: 4,
              }}
            >
              <span>Bayar ({txn.payment_method.toUpperCase()})</span>
              <span>{formatRp(txn.payment_amount)}</span>
            </div>
            {txn.change_amount > 0 && (
              <div
                style={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  color: 'var(--accent-em)',
                  fontSize: '0.8rem',
                }}
              >
                <span>Kembalian</span>
                <span>{formatRp(txn.change_amount)}</span>
              </div>
            )}
          </div>

          <div
            style={{
              textAlign: 'center',
              marginTop: 12,
              color: 'var(--text-3)',
              fontSize: '0.75rem',
              borderTop: '1px dashed var(--border-md)',
              paddingTop: 10,
            }}
          >
            Terima kasih telah berbelanja!
            <br />
            No. Transaksi: {txn.id.slice(0, 8).toUpperCase()}
          </div>
        </div>

        <div style={{ display: 'flex', gap: 8, marginTop: 16 }}>
          <button className="btn btn-secondary" style={{ flex: 1 }} onClick={onClose}>
            Tutup
          </button>
          <button className="btn btn-primary" style={{ flex: 1 }} onClick={handlePrint}>
            <Printer size={15} /> Cetak
          </button>
        </div>
      </div>
    </div>
  );
}

// ── Main POS Page ─────────────────────────────────────────────────────────────
export default function POSPage() {
  const { user, selectedStore } = useAuth();
  const isRestaurant = selectedStore?.store_type === 'restaurant';
  const { saveTransaction } = useOfflineTransaction(selectedStore?.store_id || '', user?.id || '');

  // ── Restaurant table selection state ────────────────────────────────────────
  const [tables, setTables] = useState<RestaurantTable[]>([]);
  const [tablesLoading, setTablesLoading] = useState(false);
  const [tablesDraftMap, setTablesDraftMap] = useState<Record<string, Transaction | null>>({});
  const [selectedTable, setSelectedTable] = useState<RestaurantTable | null>(null);
  const [isTakeAway, setIsTakeAway] = useState(false); // restaurant take-away (no table)
  const [activeDraft, setActiveDraft] = useState<Transaction | null>(null);
  const [holdLoading, setHoldLoading] = useState(false);

  // Retail state
  const [products, setProducts] = useState<Product[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);

  // Restaurant state
  const [menuItems, setMenuItems] = useState<MenuItem[]>([]);
  const [menuCategories, setMenuCategories] = useState<string[]>([]);

  const [cart, dispatch] = useReducer(cartReducer, []);
  const [search, setSearch] = useState('');
  const [activeCat, setActiveCat] = useState('all');
  const [loading, setLoading] = useState(true);
  const [showPayment, setShowPayment] = useState(false);
  const [payLoading, setPayLoading] = useState(false);
  const [receipt, setReceipt] = useState<Transaction | null>(null);
  const [error, setError] = useState('');
  const [holdError, setHoldError] = useState('');

  // Discount panel expand/collapse state
  const [expandedDiscountId, setExpandedDiscountId] = useState<string | null>(null);
  const [cartDiscountOpen, setCartDiscountOpen] = useState(false);
  const [isMobileCartOpen, setIsMobileCartOpen] = useState(false);
  const [discountErrors, setDiscountErrors] = useState<Record<string, string>>({});
  const autoCollapseRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  // Cart-level discount state
  const [cartDiscountType, setCartDiscountType] = useState<'PERCENTAGE' | 'FIXED'>('PERCENTAGE');
  const [cartDiscountValue, setCartDiscountValue] = useState(0);

  // Customer picker
  const [selectedCustomer, setSelectedCustomer] = useState<Customer | null>(null);
  const [custSearch, setCustSearch] = useState('');
  const [custResults, setCustResults] = useState<Customer[]>([]);
  const [custOpen, setCustOpen] = useState(false);
  const custRef = useRef<HTMLDivElement>(null);

  // Loyalty
  const loyalty = useLoyalty();

  const storeId = selectedStore?.store_id;

  // Fetch loyalty balance whenever a customer is selected
  useEffect(() => {
    if (storeId && selectedCustomer) {
      loyalty.fetchBalance(storeId, selectedCustomer.id);
    } else {
      loyalty.reset();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [storeId, selectedCustomer?.id]);

  // Customer search
  const searchCustomers = useCallback(
    async (q: string) => {
      if (!storeId || q.length < 1) {
        setCustResults([]);
        return;
      }
      try {
        const r = await customersApi.search(storeId, q);
        setCustResults((r.data as Customer[]) ?? []);
      } catch {
        setCustResults([]);
      }
    },
    [storeId]
  );

  useEffect(() => {
    const t = setTimeout(() => searchCustomers(custSearch), 250);
    return () => clearTimeout(t);
  }, [custSearch, searchCustomers]);

  useEffect(() => {
    const h = (e: MouseEvent) => {
      if (custRef.current && !custRef.current.contains(e.target as Node)) setCustOpen(false);
    };
    document.addEventListener('mousedown', h);
    return () => document.removeEventListener('mousedown', h);
  }, []);

  // Reset cart and refetch when store / mode changes; load tables for restaurant
  useEffect(() => {
    if (!storeId) return;
    dispatch({ type: 'CLEAR' });
    setActiveCat('all');
    setSearch('');
    setLoading(true);
    setSelectedTable(null);
    setActiveDraft(null);

    if (isRestaurant) {
      // Load tables + menu items in parallel
      setTablesLoading(true);
      Promise.all([tablesApi.list(storeId), menuItemsApi.list(storeId)])
        .then(([tRes, mRes]) => {
          const tbls = tRes.data as RestaurantTable[];
          setTables(tbls.filter(t => t.is_active !== false));
          // For each table, check if there's an open draft (so we can show the order summary badge)
          tbls
            .filter(t => t.is_active !== false)
            .forEach(async t => {
              try {
                const dr = await transactionsApi.getDraftByTable(storeId, t.id);
                setTablesDraftMap(prev => ({ ...prev, [t.id]: dr.data as Transaction | null }));
              } catch {
                /* ignore */
              }
            });
          const items = mRes.data as MenuItem[];
          setMenuItems(items);
          const cats = Array.from(
            new Set(items.map(i => i.category_name ?? 'Lainnya').filter(Boolean))
          );
          setMenuCategories(cats);
        })
        .catch(console.error)
        .finally(() => {
          setLoading(false);
          setTablesLoading(false);
        });
    } else {
      Promise.all([
        productsApi.list(storeId, { per_page: 200 }),
        productsApi.listCategories(storeId),
      ])
        .then(([p, c]) => {
          setProducts((p.data as PaginatedData<Product>).data ?? []);
          setCategories(c.data as Category[]);
        })
        .catch(console.error)
        .finally(() => setLoading(false));
    }
  }, [storeId, isRestaurant]);

  // When a table is selected in restaurant mode, load its draft order
  const handleSelectTable = useCallback(
    async (table: RestaurantTable) => {
      if (!storeId) return;
      setSelectedTable(table);
      dispatch({ type: 'CLEAR' });
      setActiveDraft(null);
      setHoldError('');
      setError('');
      try {
        const res = await transactionsApi.getDraftByTable(storeId, table.id);
        const draft = res.data as Transaction | null;
        setActiveDraft(draft);
        if (draft?.items) {
          // Restore cart from draft items (map back to PosCartItem & aggregate split rows)
          const aggregatedItems = new Map<string, AggregatedDraftItem>();
          draft.items.forEach(item => {
            const id = item.menu_item_id ?? item.product_id ?? item.id;
            if (aggregatedItems.has(id)) {
              const existing = aggregatedItems.get(id);
              if (existing) existing.quantity += item.quantity;
            } else {
              aggregatedItems.set(id, { ...item, parsedId: id });
            }
          });

          aggregatedItems.forEach(item => {
            const fakeMenuItem: MenuItem = {
              id: item.parsedId,
              store_id: storeId,
              name: item.product_name,
              description: '',
              sell_price: item.unit_price,
              cost_price: item.cost_price || 0,
              tax_rate: item.tax_rate,
              is_active: true,
              category_name: '',
              ingredients: [],
              created_at: '',
              updated_at: '',
            };
            dispatch({ type: 'ADD_MENU', item: fakeMenuItem });
            // Set the correct aggregate quantity
            if (item.quantity > 1) {
              dispatch({ type: 'SET_QTY', id: fakeMenuItem.id, qty: item.quantity });
            }
          });
        }
      } catch {
        /* no draft is fine */
      }
    },
    [storeId]
  );

  // Back to table selection (restaurant)
  const handleBackToTables = useCallback(() => {
    setSelectedTable(null);
    setIsTakeAway(false);
    setActiveDraft(null);
    dispatch({ type: 'CLEAR' });
    setError('');
    setHoldError('');
    // Re-fetch tables to refresh occupied status
    if (!storeId) return;
    tablesApi
      .list(storeId)
      .then(res => {
        const tbls = res.data as RestaurantTable[];
        setTables(tbls.filter(t => t.is_active !== false));
        tbls
          .filter(t => t.is_active !== false)
          .forEach(async t => {
            try {
              const dr = await transactionsApi.getDraftByTable(storeId, t.id);
              setTablesDraftMap(prev => ({ ...prev, [t.id]: dr.data as Transaction | null }));
            } catch {
              /* ignore */
            }
          });
      })
      .catch(console.error);
  }, [storeId]);

  // Hold order (create or update draft)
  const handleHoldOrder = useCallback(async () => {
    if (!storeId || !selectedTable || cart.length === 0) return;
    setHoldLoading(true);
    setHoldError('');
    try {
      const items = (cart as PosCartItem[]).map(i => ({
        product_id: i.menuItemId ? '' : i.product.id,
        menu_item_id: i.menuItemId ?? '',
        quantity: i.quantity,
        discount_pct: i.discountType === 'PERCENTAGE' ? i.discountValue : 0,
        discount_type: i.discountType,
        discount_value: i.discountValue,
      }));
      if (activeDraft) {
        await transactionsApi.updateDraft(storeId, activeDraft.id, {
          items,
          cart_discount_type: cartDiscountType,
          cart_discount_value: cartDiscountValue,
        });
      } else {
        await transactionsApi.createDraft(storeId, {
          table_id: selectedTable.id,
          items,
          cart_discount_type: cartDiscountType,
          cart_discount_value: cartDiscountValue,
        });
      }
      // Mark table occupied in UI
      await tablesApi.updateStatus(storeId, selectedTable.id, 'occupied');
      handleBackToTables();
    } catch (err) {
      if (err instanceof ApiError) setHoldError(err.message);
      else setHoldError('Gagal menyimpan pesanan');
    } finally {
      setHoldLoading(false);
    }
  }, [
    storeId,
    selectedTable,
    activeDraft,
    cart,
    handleBackToTables,
    cartDiscountType,
    cartDiscountValue,
  ]);

  // ── Filtered lists ──────────────────────────────────────────────────────────
  const filteredProducts = products.filter(p => {
    if (!p.is_active) return false;
    const matchSearch =
      p.name.toLowerCase().includes(search.toLowerCase()) ||
      p.sku.toLowerCase().includes(search.toLowerCase());
    const matchCat = activeCat === 'all' || p.category_id === activeCat;
    return matchSearch && matchCat;
  });

  const filteredMenuItems = menuItems.filter(m => {
    const matchSearch = m.name.toLowerCase().includes(search.toLowerCase());
    const catName = m.category_name ?? 'Lainnya';
    const matchCat = activeCat === 'all' || catName === activeCat;
    return matchSearch && matchCat;
  });

  // ── Totals ───────────────────────────────────────────────────────────────────
  const subtotalAfterItems = cart.reduce((s, i) => s + i.subtotal, 0); // after item discounts
  const taxAmt = cart.reduce((s, i) => s + i.taxAmt, 0);
  const itemCount = cart.reduce((s, i) => s + i.quantity, 0);

  // Cart-level discount amount (for display only; backend re-computes authoritative values)
  const cartDiscountAmt =
    cartDiscountType === 'PERCENTAGE'
      ? (subtotalAfterItems * cartDiscountValue) / 100
      : Math.min(cartDiscountValue, subtotalAfterItems);

  const subtotal = Math.max(0, subtotalAfterItems - cartDiscountAmt); // post-cart-discount net
  const total = subtotal + taxAmt;

  // Total discount across both levels (item + cart)
  const totalItemDiscount = cart.reduce(
    (s, i) => s + (i.originalPrice - i.unitPrice) * i.quantity,
    0
  );
  const totalDiscount = totalItemDiscount + cartDiscountAmt;

  // ── Checkout (retail + restaurant direct pay) ────────────────────────────────
  const handleConfirmPayment = useCallback(
    async (method: string, amount: number) => {
      if (!storeId) return;
      setPayLoading(true);
      setError('');
      try {
        let res;
        if (isRestaurant && activeDraft) {
          // Pay via the draft endpoint
          res = await transactionsApi.payDraft(storeId, activeDraft.id, {
            payment_method: method,
            payment_amount: amount,
            customer_name: selectedCustomer?.name ?? '',
            customer_phone: selectedCustomer?.phone ?? '',
          });
          // Table goes back to available
          if (selectedTable) {
            await tablesApi.updateStatus(storeId, selectedTable.id, 'available').catch(() => {});
          }
          setReceipt(res.data as Transaction);
        } else {
          // OFFLINE-FIRST RETAIL CHECKOUT
          const txItems = (cart as PosCartItem[]).map(i => ({
            product_id: i.menuItemId ? null : i.product.id,
            menu_item_id: i.menuItemId ?? null,
            product_name: i.product.name,
            sku: i.product.sku,
            quantity: i.quantity,
            original_price: i.originalPrice,
            unit_price: i.unitPrice,
            cost_price: i.product.cost_price,
            discount_pct: i.discountType === 'PERCENTAGE' ? i.discountValue : 0,
            discount_type: i.discountType,
            discount_value: i.discountValue,
            cart_discount_allocated: 0,
            tax_rate: i.product.tax_rate ?? 0,
            subtotal: i.subtotal,
            status: 'completed',
          }));

          const payloadData = {
            table_id: null,
            customer_name: selectedCustomer?.name ?? '',
            customer_phone: selectedCustomer?.phone ?? '',
            subtotal: subtotalAfterItems,
            discount_amt: cartDiscountAmt,
            tax_amt: taxAmt,
            total: total,
            payment_method: method,
            payment_amount: amount,
            change_amount: amount - total,
            status: 'completed',
            notes: '',
            cart_discount_type: cartDiscountType,
            cart_discount_value: cartDiscountValue,
            items: txItems as any,
          };

          const result = await saveTransaction(payloadData);

          setReceipt({
            id: result.transactionId,
            store_id: storeId,
            cashier_id: user?.id || '',
            cashier_name: user?.name,
            ...payloadData,
            created_at: new Date().toISOString(),
          } as unknown as Transaction);
          // Earn loyalty points for the customer (fire-and-forget)
          if (selectedCustomer) {
            loyalty
              .earnPoints(storeId, selectedCustomer.id, result.transactionId, total)
              .catch(() => {});
          }
        }

        setSelectedCustomer(null);
        setCustSearch('');
        loyalty.reset();
        setShowPayment(false);
        setCartDiscountValue(0);
        setCartDiscountType('PERCENTAGE');
        setCartDiscountOpen(false);
        setExpandedDiscountId(null);
        setDiscountErrors({});
        if (isRestaurant) {
          handleBackToTables();
        } else {
          dispatch({ type: 'CLEAR' });
          setActiveDraft(null);
        }
      } catch (err) {
        if (err instanceof ApiError) setError(err.message);
        else setError('Gagal memproses pembayaran');
      } finally {
        setPayLoading(false);
      }
    },
    [
      storeId,
      cart,
      selectedCustomer,
      isRestaurant,
      activeDraft,
      selectedTable,
      handleBackToTables,
      cartDiscountType,
      cartDiscountValue,
    ]
  );

  if (!selectedStore) {
    return (
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          height: '100vh',
        }}
      >
        <div className="empty-state">
          <ShoppingCart size={40} />
          <p className="type-body-sm">Pilih toko di sidebar untuk memulai</p>
        </div>
      </div>
    );
  }

  // ── Restaurant: table selection screen ─────────────────────────────────────
  if (isRestaurant && !selectedTable && !isTakeAway) {
    return (
      <>
        <div
          style={{ padding: '28px 32px', minHeight: '100vh', background: 'var(--bg-app)' }}
          className="reveal-animate"
        >
          {/* Header */}
          <div className="flex items-center justify-between" style={{ marginBottom: 32 }}>
            <div>
              <h1 className="page-title" style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                <div
                  style={{
                    background: 'rgba(8,132,246,0.1)',
                    color: 'var(--brand)',
                    padding: 8,
                    borderRadius: 12,
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                  }}
                >
                  <UtensilsCrossed size={22} />
                </div>
                Pilih Meja
              </h1>
              <p className="page-subtitle" style={{ marginLeft: 42 }}>
                {selectedStore.store_name}
              </p>
            </div>

            {/* Take Away shortcut button */}
            <button
              onClick={() => {
                setIsTakeAway(true);
                dispatch({ type: 'CLEAR' });
                setError('');
                setHoldError('');
              }}
              className="btn shadow-hover"
              style={{
                background: 'rgba(245,158,11,0.08)',
                border: '1px solid rgba(245,158,11,0.25)',
                color: '#d97706',
                fontWeight: 700,
                gap: 8,
                padding: '10px 20px',
                borderRadius: 14,
                transition: 'all 0.2s ease',
              }}
            >
              <ShoppingBag size={18} />
              Bungkus / Take Away
            </button>
          </div>

          {tablesLoading ? (
            <div style={{ display: 'flex', justifyContent: 'center', padding: 80 }}>
              <Loader2 size={40} className="loading-spin" style={{ color: 'var(--brand)' }} />
            </div>
          ) : tables.length === 0 ? (
            <div className="empty-state" style={{ marginTop: 40 }}>
              <UtensilsCrossed size={48} />
              <p style={{ fontSize: '1.1rem', fontWeight: 500 }}>Belum ada meja yang terdaftar.</p>
              <p style={{ color: 'var(--text-3)', fontSize: '0.9rem' }}>
                Tambahkan konfigurasi meja Anda di menu Manajemen Meja.
              </p>
            </div>
          ) : (
            <div className="pos-table-grid">
              {tables.map((table, i) => {
                const draft = tablesDraftMap[table.id];
                const hasOrder = !!draft;
                const isOccupied = table.status === 'occupied' || hasOrder;
                const isUnavailable = table.status === 'unavailable';

                const statusClass = isUnavailable
                  ? 'unavailable'
                  : isOccupied
                    ? 'occupied'
                    : 'available';

                return (
                  <button
                    key={table.id}
                    className={`pos-table-card ${statusClass} reveal-animate shadow-hover`}
                    onClick={() => !isUnavailable && handleSelectTable(table)}
                    style={{ animationDelay: `${i * 0.04}s` }}
                  >
                    <div className="pos-table-icon-wrapper">
                      <UtensilsCrossed size={24} strokeWidth={2.5} />
                    </div>

                    <div className="pos-table-number">Meja {table.table_number}</div>

                    <div className="pos-table-capacity">
                      <Users size={12} strokeWidth={2.5} />
                      <span>{table.capacity} Kursi</span>
                    </div>

                    <div className={`status-${statusClass}`}>
                      <div className="pos-table-status-pill">
                        {isUnavailable
                          ? 'Offline'
                          : isOccupied
                            ? hasOrder
                              ? 'Draf Aktif'
                              : 'Terisi'
                            : 'Tersedia'}
                      </div>
                    </div>

                    {hasOrder && draft && (
                      <div
                        style={{
                          marginTop: 6,
                          fontSize: '0.68rem',
                          color: 'var(--accent-em)',
                          fontWeight: 700,
                          background: 'rgba(8,132,246,0.06)',
                          padding: '2px 8px',
                          borderRadius: 6,
                        }}
                      >
                        {draft.items.length} item · {formatRp(draft.total)}
                      </div>
                    )}
                  </button>
                );
              })}
            </div>
          )}
        </div>
        {receipt && <ReceiptModal txn={receipt} onClose={() => setReceipt(null)} />}
      </>
    );
  }

  // ── Render (Retail + Restaurant order screen) ───────────────────────────────
  return (
    <div className="pos-layout">
      {/* ── LEFT: Catalog ── */}
      <div className="pos-catalog">
        {/* ── Catalog header row: mode badge + search ── */}
        <div className="flex items-center gap-2 reveal-animate" style={{ marginBottom: 10 }}>
          {isRestaurant && (selectedTable || isTakeAway) && (
            <button
              onClick={handleBackToTables}
              className="btn btn-ghost btn-sm"
              style={{ padding: '5px 8px', gap: 4, flexShrink: 0 }}
            >
              <ArrowLeft size={14} />
              <span style={{ fontSize: '0.8rem' }}>Meja</span>
            </button>
          )}
          {/* Mode badge */}
          <span
            style={{
              display: 'inline-flex',
              alignItems: 'center',
              gap: 5,
              padding: '4px 12px',
              borderRadius: 100,
              fontSize: '0.75rem',
              fontWeight: 600,
              flexShrink: 0,
              background: isRestaurant ? 'rgba(251,146,60,0.10)' : 'rgba(8,132,246,0.10)',
              color: isRestaurant ? '#fb923c' : '#0884f6',
              border: isRestaurant
                ? '1px solid rgba(251,146,60,0.25)'
                : '1px solid rgba(8,132,246,0.20)',
            }}
          >
            {isRestaurant ? <UtensilsCrossed size={12} /> : <ShoppingBag size={12} />}
            {isRestaurant
              ? isTakeAway
                ? 'Take Away'
                : `Meja ${selectedTable?.table_number ?? ''}`
              : 'Retail'}
          </span>
          {/* Search — takes remaining space */}
          <div style={{ flex: 1, position: 'relative' }}>
            <Search
              size={15}
              style={{
                position: 'absolute',
                left: 11,
                top: '50%',
                transform: 'translateY(-50%)',
                color: 'var(--text-3)',
                pointerEvents: 'none',
              }}
            />
            <input
              className="input"
              style={{ paddingLeft: 34 }}
              placeholder={isRestaurant ? 'Cari menu...' : 'Cari produk atau SKU...'}
              value={search}
              onChange={e => setSearch(e.target.value)}
            />
          </div>
        </div>

        {/* Category Tabs */}
        <div className="category-tabs reveal-animate" style={{ animationDelay: '0.1s' }}>
          <button
            className={`cat-tab ${activeCat === 'all' ? 'active' : ''}`}
            onClick={() => setActiveCat('all')}
          >
            Semua
          </button>
          {isRestaurant
            ? menuCategories.map(c => (
                <button
                  key={c}
                  className={`cat-tab ${activeCat === c ? 'active' : ''}`}
                  onClick={() => setActiveCat(c)}
                >
                  {c}
                </button>
              ))
            : categories.map(c => (
                <button
                  key={c.id}
                  className={`cat-tab ${activeCat === c.id ? 'active' : ''}`}
                  onClick={() => setActiveCat(c.id)}
                >
                  {c.name}
                </button>
              ))}
        </div>

        {/* Error */}
        {error && (
          <div
            style={{
              background: 'rgba(239,68,68,0.12)',
              border: '1px solid rgba(239,68,68,0.3)',
              borderRadius: 8,
              padding: '10px 14px',
              color: '#f87171',
              fontSize: '0.85rem',
              display: 'flex',
              justifyContent: 'space-between',
              marginBottom: 10,
            }}
          >
            {error}
            <button
              onClick={() => setError('')}
              style={{ background: 'none', border: 'none', color: '#f87171', cursor: 'pointer' }}
            >
              <X size={14} />
            </button>
          </div>
        )}

        {/* Grid */}
        {loading ? (
          <div style={{ display: 'flex', justifyContent: 'center', padding: 48 }}>
            <Loader2 size={28} className="loading-spin" style={{ color: 'var(--accent-em)' }} />
          </div>
        ) : isRestaurant ? (
          /* ── Restaurant: Menu Items ── */
          filteredMenuItems.length === 0 ? (
            <div className="empty-state">
              <UtensilsCrossed size={32} />
              <p>Tidak ada menu ditemukan</p>
            </div>
          ) : (
            <div className="product-grid">
              {filteredMenuItems.map((m, i) => {
                const inCart = cart.find(i => i.menuItemId === m.id);
                return (
                  <div
                    key={m.id}
                    className="product-card reveal-animate"
                    onClick={() => dispatch({ type: 'ADD_MENU', item: m })}
                    style={{ animationDelay: `${0.2 + i * 0.012}s` }}
                  >
                    {/* In-cart quantity badge */}
                    {inCart && (
                      <div
                        style={{
                          position: 'absolute',
                          top: 8,
                          right: 8,
                          background: 'var(--brand)',
                          color: '#fff',
                          borderRadius: '50%',
                          width: 20,
                          height: 20,
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'center',
                          fontSize: '0.68rem',
                          fontWeight: 700,
                          boxShadow: '0 2px 6px rgba(8,132,246,0.4)',
                        }}
                      >
                        {inCart.quantity}
                      </div>
                    )}
                    <div className="product-icon">🍽️</div>
                    <div className="product-name">{m.name}</div>
                    {m.category_name && <div className="product-sku">{m.category_name}</div>}
                    <div className="product-price">{formatRp(m.sell_price)}</div>
                    {m.ingredients && m.ingredients.length > 0 && (
                      <div className="product-sku" style={{ marginTop: 2 }}>
                        🧂 {m.ingredients.length} bahan
                      </div>
                    )}
                    {/* Add to cart button */}
                    <button
                      className="btn btn-primary btn-sm"
                      style={{
                        marginTop: 'auto',
                        width: '100%',
                        justifyContent: 'center',
                        fontSize: '0.75rem',
                      }}
                      onClick={e => {
                        e.stopPropagation();
                        dispatch({ type: 'ADD_MENU', item: m });
                      }}
                    >
                      <Plus size={13} />
                      Tambah
                    </button>
                  </div>
                );
              })}
            </div>
          )
        ) : /* ── Retail: Products ── */
        filteredProducts.length === 0 ? (
          <div className="empty-state">
            <Search size={32} />
            <p>Tidak ada produk ditemukan</p>
          </div>
        ) : (
          <div className="product-grid">
            {filteredProducts.map((p, i) => {
              const inCart = cart.find(i => i.product.id === p.id && !i.menuItemId);
              const outOfStock = (p.stock_qty ?? 1) <= 0;
              return (
                <div
                  key={p.id}
                  className={`product-card ${outOfStock ? 'out-of-stock' : ''} reveal-animate`}
                  onClick={() => !outOfStock && dispatch({ type: 'ADD_PRODUCT', product: p })}
                  style={{ animationDelay: `${0.2 + i * 0.012}s` }}
                >
                  {/* In-cart quantity badge */}
                  {inCart && (
                    <div
                      style={{
                        position: 'absolute',
                        top: 8,
                        right: 8,
                        background: 'var(--brand)',
                        color: '#fff',
                        borderRadius: '50%',
                        width: 20,
                        height: 20,
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        fontSize: '0.68rem',
                        fontWeight: 700,
                        boxShadow: '0 2px 6px rgba(8,132,246,0.4)',
                      }}
                    >
                      {inCart.quantity}
                    </div>
                  )}
                  <div className="product-icon">{productEmoji(p.name)}</div>
                  <div className="product-name">{p.name}</div>
                  <div className="product-sku">{p.sku}</div>
                  <div className="product-price">{formatRp(p.sell_price)}</div>
                  <div className="product-stock">
                    {outOfStock ? (
                      <span className="badge badge-red">Habis</span>
                    ) : p.stock_qty !== undefined && p.stock_qty <= 5 ? (
                      <span className="badge badge-amber">
                        {p.stock_qty} {p.unit}
                      </span>
                    ) : (
                      <span className="badge badge-green">
                        {p.stock_qty ?? '–'} {p.unit}
                      </span>
                    )}
                  </div>
                  {/* Add to cart button */}
                  {!outOfStock && (
                    <button
                      className="btn btn-primary btn-sm"
                      style={{
                        marginTop: 'auto',
                        width: '100%',
                        justifyContent: 'center',
                        fontSize: '0.75rem',
                      }}
                      onClick={e => {
                        e.stopPropagation();
                        dispatch({ type: 'ADD_PRODUCT', product: p });
                      }}
                    >
                      <Plus size={13} />
                      Tambah
                    </button>
                  )}
                </div>
              );
            })}
          </div>
        )}
      </div>

      {/* Mobile Floating Cart Button */}
      {!isMobileCartOpen && (
        <div className="md:hidden fixed bottom-4 left-4 right-4 z-40">
          <button
            className="checkout-btn"
            style={{
              width: '100%',
              padding: '14px',
              borderRadius: '12px',
              boxShadow: '0 8px 30px rgba(8,132,246,0.4)',
              justifyContent: 'center',
            }}
            onClick={() => setIsMobileCartOpen(true)}
          >
            <ShoppingBag size={18} />
            {isRestaurant ? 'Pesanan' : 'Keranjang'} ({itemCount}) - {formatRp(total)}
          </button>
        </div>
      )}

      {/* ── RIGHT: Cart ── */}
      <div className={`pos-cart ${isMobileCartOpen ? 'open' : ''}`}>
        <div className="cart-header">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <ShoppingCart size={17} style={{ color: 'var(--brand)' }} />
              <span style={{ fontWeight: 700, fontSize: '0.9375rem', color: 'var(--text-1)' }}>
                {isRestaurant ? 'Pesanan' : 'Keranjang'}
              </span>
              {itemCount > 0 && (
                <span className="badge badge-blue" style={{ fontVariantNumeric: 'tabular-nums' }}>
                  {itemCount}
                </span>
              )}
            </div>
            <div className="flex items-center gap-1">
              <button
                className="btn btn-ghost btn-sm md:hidden"
                onClick={() => setIsMobileCartOpen(false)}
              >
                <X size={16} />
              </button>
              {cart.length > 0 && (
                <button
                  className="btn btn-ghost btn-sm"
                  style={{ color: 'var(--accent-rd)', fontSize: '0.75rem', gap: 4 }}
                  onClick={() => dispatch({ type: 'CLEAR' })}
                >
                  <Trash2 size={12} />
                  Kosongkan
                </button>
              )}
            </div>
          </div>
        </div>
        {/* ── Customer Picker — inside cart, always visible ── */}
        <div
          ref={custRef}
          style={{
            padding: '10px 12px',
            borderBottom: '1px solid var(--border)',
            background: 'var(--bg-elevated)',
            position: 'relative',
          }}
        >
          <div
            style={{
              fontSize: '0.68rem',
              fontWeight: 700,
              color: 'var(--text-3)',
              textTransform: 'uppercase',
              letterSpacing: '0.08em',
              marginBottom: 6,
              display: 'flex',
              alignItems: 'center',
              gap: 4,
            }}
          >
            <UserRound size={11} /> Customer
          </div>
          {selectedCustomer ? (
            <div
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: 8,
                background: 'rgba(16,185,129,0.1)',
                border: '1.5px solid rgba(16,185,129,0.4)',
                borderRadius: 8,
                padding: '7px 10px',
              }}
            >
              <div
                style={{
                  width: 28,
                  height: 28,
                  borderRadius: '50%',
                  background: 'linear-gradient(135deg, #10b981, #059669)',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  fontWeight: 800,
                  fontSize: '0.75rem',
                  color: '#fff',
                  flexShrink: 0,
                }}
              >
                {selectedCustomer.name.charAt(0).toUpperCase()}
              </div>
              <div style={{ flex: 1, minWidth: 0 }}>
                <div style={{ fontWeight: 700, fontSize: '0.85rem', color: '#10b981' }}>
                  {selectedCustomer.name}
                </div>
                {selectedCustomer.phone && (
                  <div style={{ fontSize: '0.72rem', color: 'var(--text-3)' }}>
                    {selectedCustomer.phone}
                  </div>
                )}
                {/* Loyalty balance badge */}
                {loyalty.balance && (
                  <div
                    style={{
                      display: 'inline-flex',
                      alignItems: 'center',
                      gap: 3,
                      marginTop: 2,
                      background: 'rgba(251,191,36,0.15)',
                      border: '1px solid rgba(251,191,36,0.4)',
                      borderRadius: 4,
                      padding: '1px 6px',
                      fontSize: '0.7rem',
                      fontWeight: 700,
                      color: '#f59e0b',
                    }}
                  >
                    <Star size={10} style={{ fill: '#f59e0b', color: '#f59e0b' }} />
                    {loyalty.balance.balance.toLocaleString('id-ID')} pts
                    {loyalty.balance.tier && (
                      <span style={{ fontWeight: 400, color: 'var(--text-3)', marginLeft: 3 }}>
                        · {loyalty.balance.tier.name}
                      </span>
                    )}
                  </div>
                )}
              </div>
              <button
                style={{
                  background: 'none',
                  border: 'none',
                  cursor: 'pointer',
                  color: 'var(--text-3)',
                  padding: 4,
                  borderRadius: 4,
                  display: 'flex',
                  alignItems: 'center',
                }}
                onClick={() => {
                  setSelectedCustomer(null);
                  setCustSearch('');
                }}
                title="Hapus customer"
              >
                <X size={13} />
              </button>
            </div>
          ) : (
            <div style={{ position: 'relative' }}>
              <UserRound
                size={13}
                style={{
                  position: 'absolute',
                  left: 10,
                  top: '50%',
                  transform: 'translateY(-50%)',
                  color: 'var(--text-3)',
                  pointerEvents: 'none',
                }}
              />
              <input
                className="input"
                style={{ paddingLeft: 32, fontSize: '0.83rem', height: 36 }}
                placeholder="Cari nama / telepon customer..."
                value={custSearch}
                onChange={e => {
                  setCustSearch(e.target.value);
                  setCustOpen(true);
                }}
                onFocus={() => setCustOpen(true)}
              />
              {custOpen && custResults.length > 0 && (
                <div
                  style={{
                    position: 'absolute',
                    top: '100%',
                    left: 0,
                    right: 0,
                    zIndex: 100,
                    background: 'var(--bg-card)',
                    border: '1px solid var(--border)',
                    borderRadius: 8,
                    boxShadow: '0 8px 32px rgba(0,0,0,0.35)',
                    marginTop: 4,
                    overflow: 'hidden',
                  }}
                >
                  {custResults.map((c: Customer) => (
                    <button
                      key={c.id}
                      onClick={() => {
                        setSelectedCustomer(c);
                        setCustSearch('');
                        setCustOpen(false);
                      }}
                      style={{
                        width: '100%',
                        padding: '8px 12px',
                        textAlign: 'left',
                        background: 'none',
                        border: 'none',
                        cursor: 'pointer',
                        display: 'flex',
                        alignItems: 'center',
                        gap: 8,
                      }}
                    >
                      <div
                        style={{
                          width: 30,
                          height: 30,
                          borderRadius: '50%',
                          background: 'linear-gradient(135deg, var(--accent-in), var(--accent-em))',
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'center',
                          fontWeight: 700,
                          fontSize: '0.78rem',
                          color: '#fff',
                          flexShrink: 0,
                        }}
                      >
                        {c.name.charAt(0).toUpperCase()}
                      </div>
                      <div>
                        <div style={{ fontWeight: 600, fontSize: '0.85rem' }}>{c.name}</div>
                        {c.phone && (
                          <div style={{ fontSize: '0.73rem', color: 'var(--text-3)' }}>
                            {c.phone}
                          </div>
                        )}
                      </div>
                    </button>
                  ))}
                </div>
              )}
            </div>
          )}
        </div>

        <div className="cart-items">
          {cart.length === 0 ? (
            <div className="empty-state" style={{ paddingTop: 48 }}>
              {isRestaurant ? (
                <UtensilsCrossed size={36} style={{ color: 'var(--text-3)' }} />
              ) : (
                <ShoppingCart size={36} style={{ color: 'var(--text-3)' }} />
              )}
              <p style={{ fontSize: '0.85rem' }}>
                {isRestaurant ? 'Klik menu untuk memesan' : 'Klik produk untuk menambahkan'}
              </p>
            </div>
          ) : (
            (cart as PosCartItem[]).map(item => {
              const isDiscounted = item.unitPrice < item.originalPrice;
              const isExpanded = expandedDiscountId === item.product.id;
              const discBadge =
                item.discountType === 'PERCENTAGE' && item.discountValue > 0
                  ? `-${item.discountValue}%`
                  : item.discountType === 'FIXED' && item.discountValue > 0
                    ? `-${formatRp(item.discountValue)}`
                    : item.discountType === 'OVERRIDE' && item.discountValue > 0
                      ? `= ${formatRp(item.discountValue)}`
                      : null;
              const itemError = discountErrors[item.product.id] ?? '';
              const costPrice = item.product.cost_price ?? 0;
              const costLabel = item.menuItemId ? 'HPP' : 'Harga Beli';
              const isBelowCost = costPrice > 0 && item.unitPrice < costPrice;

              const pctPresets = [5, 10, 15, 20];
              const fixedPresets = [5000, 10000];

              const handleQuickApply = (type: DiscountType, value: number) => {
                if (type === 'PERCENTAGE' && value > 100) return;
                setDiscountErrors(prev => ({ ...prev, [item.product.id]: '' }));
                dispatch({
                  type: 'SET_DISCOUNT',
                  id: item.product.id,
                  discountType: type,
                  discountValue: value,
                });
                if (autoCollapseRef.current) clearTimeout(autoCollapseRef.current);
                autoCollapseRef.current = setTimeout(() => {
                  setExpandedDiscountId(prev => (prev === item.product.id ? null : prev));
                }, 1200);
              };

              const handleInputChange = (val: number) => {
                let error = '';
                let safeVal = val;
                if (val < 0) {
                  safeVal = 0;
                  error = 'Nilai tidak boleh negatif';
                } else if (item.discountType === 'PERCENTAGE' && val > 100) {
                  safeVal = 100;
                  error = 'Diskon maksimal 100%';
                } else if (item.discountType === 'FIXED' && val > item.originalPrice) {
                  safeVal = item.originalPrice;
                  error = 'Diskon melebihi harga';
                }
                // Warn if final price would be below cost
                const simFinal =
                  item.discountType === 'PERCENTAGE'
                    ? item.originalPrice * (1 - safeVal / 100)
                    : item.originalPrice - safeVal;
                if (costPrice > 0 && simFinal < costPrice && !error) {
                  error = `⚠ Harga jual di bawah ${costLabel} (${formatRp(costPrice)})`;
                }
                setDiscountErrors(prev => ({ ...prev, [item.product.id]: error }));
                dispatch({
                  type: 'SET_DISCOUNT',
                  id: item.product.id,
                  discountType: item.discountType,
                  discountValue: safeVal,
                });
              };

              return (
                <div
                  key={item.product.id}
                  className={`cart-item ${isDiscounted ? 'is-discounted' : ''}`}
                >
                  {/* Remove Button (Absolute positioned for minimalism) */}
                  <button
                    className="btn btn-ghost"
                    style={{
                      position: 'absolute',
                      right: 4,
                      top: 4,
                      padding: '4px',
                      color: 'var(--text-3)',
                      opacity: 0.6,
                    }}
                    onClick={() => {
                      dispatch({ type: 'REMOVE', id: item.product.id });
                      if (expandedDiscountId === item.product.id) setExpandedDiscountId(null);
                    }}
                  >
                    <X size={14} />
                  </button>

                  <div className="cart-item-header">
                    <div className="cart-item-name">
                      {item.menuItemId && (
                        <span style={{ fontSize: '0.7rem', color: '#fb923c', marginRight: 5 }}>
                          🍽
                        </span>
                      )}
                      {item.product.name}
                    </div>
                    <div className="cart-item-total">
                      {isDiscounted && (
                        <div
                          style={{
                            fontSize: '0.7rem',
                            color: 'var(--text-3)',
                            textDecoration: 'line-through',
                            fontWeight: 500,
                            lineHeight: 1,
                            marginBottom: 2,
                          }}
                        >
                          {formatRp(item.originalPrice * item.quantity)}
                        </div>
                      )}
                      {formatRp(item.subtotal)}
                    </div>
                  </div>

                  <div className="cart-item-body">
                    <div className="qty-ctrl">
                      <button
                        className="qty-btn"
                        onClick={() =>
                          dispatch({
                            type: 'SET_QTY',
                            id: item.product.id,
                            qty: item.quantity - 1,
                          })
                        }
                      >
                        <Minus size={13} />
                      </button>
                      <span className="qty-val">{item.quantity}</span>
                      <button
                        className="qty-btn"
                        onClick={() =>
                          dispatch({
                            type: 'SET_QTY',
                            id: item.product.id,
                            qty: item.quantity + 1,
                          })
                        }
                      >
                        <Plus size={13} />
                      </button>
                    </div>

                    <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                      {/* Detailed Price Context */}
                      <div className="cart-item-details" style={{ textAlign: 'right' }}>
                        <div>
                          {formatRp(item.unitPrice)} × {item.quantity}
                        </div>
                        {item.product.tax_rate > 0 && (
                          <div style={{ fontSize: '0.65rem', opacity: 0.8 }}>
                            + PPN {item.product.tax_rate}%
                          </div>
                        )}
                        {costPrice > 0 && (
                          <div
                            style={{
                              fontSize: '0.65rem',
                              color: isBelowCost ? '#f87171' : 'var(--text-3)',
                              fontWeight: isBelowCost ? 700 : 500,
                            }}
                          >
                            {costLabel}: {formatRp(costPrice)}
                          </div>
                        )}
                      </div>

                      {/* Discount Trigger */}
                      <button
                        className={`btn-discount-trigger ${isDiscounted ? 'active' : ''}`}
                        onClick={() => setExpandedDiscountId(isExpanded ? null : item.product.id)}
                      >
                        <Tag size={10} />
                        {isDiscounted && discBadge ? discBadge : 'Diskon'}
                      </button>
                    </div>
                  </div>

                  {/* Collapsible discount panel */}
                  <div
                    style={{
                      overflow: 'hidden',
                      maxHeight: isExpanded ? 160 : 0,
                      opacity: isExpanded ? 1 : 0,
                      marginTop: isExpanded ? 6 : 0,
                      transition:
                        'max-height 0.28s cubic-bezier(0.4,0,0.2,1), opacity 0.2s ease, margin 0.2s ease',
                    }}
                  >
                    <div
                      style={{
                        display: 'flex',
                        flexDirection: 'column',
                        gap: 6,
                        padding: '8px 10px',
                        background: 'var(--bg-card)',
                        border: '1px solid var(--border-md)',
                        borderRadius: 8,
                      }}
                    >
                      {/* Type toggle + quick buttons row */}
                      <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
                        {/* % / Rp toggle group */}
                        <div
                          style={{
                            display: 'flex',
                            gap: 0,
                            background: 'var(--bg-elevated)',
                            borderRadius: 6,
                            padding: 2,
                            border: '1px solid var(--border)',
                          }}
                        >
                          {(['PERCENTAGE', 'FIXED'] as DiscountType[]).map(dt => (
                            <button
                              key={dt}
                              onClick={() => {
                                setDiscountErrors(prev => ({ ...prev, [item.product.id]: '' }));
                                dispatch({
                                  type: 'SET_DISCOUNT',
                                  id: item.product.id,
                                  discountType: dt,
                                  discountValue: 0,
                                });
                              }}
                              style={{
                                flex: 1,
                                padding: '4px 10px',
                                border: 'none',
                                borderRadius: 5,
                                fontSize: '0.68rem',
                                fontWeight: 700,
                                cursor: 'pointer',
                                transition: 'all 0.15s ease',
                                background:
                                  item.discountType === dt ? 'var(--accent-em)' : 'transparent',
                                color: item.discountType === dt ? '#fff' : 'var(--text-3)',
                                boxShadow:
                                  item.discountType === dt
                                    ? '0 1px 4px rgba(8,132,246,0.3)'
                                    : 'none',
                              }}
                            >
                              {dt === 'PERCENTAGE' ? '%' : 'Rp'}
                            </button>
                          ))}
                        </div>
                        {/* Quick preset buttons */}
                        <div style={{ display: 'flex', gap: 4, flex: 1, flexWrap: 'wrap' }}>
                          {item.discountType === 'PERCENTAGE'
                            ? pctPresets.map(v => {
                                const isSelected =
                                  item.discountType === 'PERCENTAGE' && item.discountValue === v;
                                return (
                                  <button
                                    key={v}
                                    onClick={() => handleQuickApply('PERCENTAGE', v)}
                                    style={{
                                      padding: '4px 8px',
                                      borderRadius: 6,
                                      fontSize: '0.68rem',
                                      fontWeight: 600,
                                      cursor: 'pointer',
                                      transition: 'all 0.12s ease',
                                      background: isSelected
                                        ? 'rgba(8,132,246,0.15)'
                                        : 'var(--bg-elevated)',
                                      color: isSelected ? 'var(--accent-em)' : 'var(--text-2)',
                                      border: isSelected
                                        ? '1px solid rgba(8,132,246,0.4)'
                                        : '1px solid var(--border)',
                                      whiteSpace: 'nowrap',
                                    }}
                                  >
                                    -{v}%
                                  </button>
                                );
                              })
                            : fixedPresets.map(v => {
                                const isSelected =
                                  item.discountType === 'FIXED' && item.discountValue === v;
                                return (
                                  <button
                                    key={v}
                                    onClick={() => handleQuickApply('FIXED', v)}
                                    style={{
                                      padding: '4px 8px',
                                      borderRadius: 6,
                                      fontSize: '0.68rem',
                                      fontWeight: 600,
                                      cursor: 'pointer',
                                      transition: 'all 0.12s ease',
                                      background: isSelected
                                        ? 'rgba(8,132,246,0.15)'
                                        : 'var(--bg-elevated)',
                                      color: isSelected ? 'var(--accent-em)' : 'var(--text-2)',
                                      border: isSelected
                                        ? '1px solid rgba(8,132,246,0.4)'
                                        : '1px solid var(--border)',
                                      whiteSpace: 'nowrap',
                                    }}
                                  >
                                    -{v / 1000}rb
                                  </button>
                                );
                              })}
                        </div>
                      </div>
                      {/* Input row */}
                      <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
                        <input
                          type="number"
                          min={0}
                          max={item.discountType === 'PERCENTAGE' ? 100 : item.originalPrice}
                          step={item.discountType === 'PERCENTAGE' ? 1 : 500}
                          placeholder={
                            item.discountType === 'PERCENTAGE' ? 'Masukkan %' : 'Masukkan Rp'
                          }
                          value={item.discountValue === 0 ? '' : item.discountValue}
                          onChange={e => handleInputChange(parseFloat(e.target.value) || 0)}
                          style={{
                            flex: 1,
                            minWidth: 0,
                            padding: '6px 8px',
                            borderRadius: 6,
                            border: '1px solid var(--border-md)',
                            background: 'var(--bg-elevated)',
                            color: 'var(--text-1)',
                            fontSize: '0.78rem',
                            fontWeight: 600,
                            outline: 'none',
                          }}
                        />
                        {item.discountValue > 0 && (
                          <button
                            onClick={() => {
                              setDiscountErrors(prev => ({ ...prev, [item.product.id]: '' }));
                              dispatch({
                                type: 'SET_DISCOUNT',
                                id: item.product.id,
                                discountType: item.discountType,
                                discountValue: 0,
                              });
                            }}
                            style={{
                              padding: '5px 10px',
                              borderRadius: 5,
                              fontSize: '0.68rem',
                              fontWeight: 600,
                              cursor: 'pointer',
                              border: 'none',
                              background: 'rgba(239,68,68,0.1)',
                              color: 'var(--accent-rd)',
                              transition: 'all 0.12s ease',
                              whiteSpace: 'nowrap',
                            }}
                          >
                            Hapus
                          </button>
                        )}
                      </div>
                      {/* Error */}
                      {itemError && (
                        <div
                          style={{
                            fontSize: '0.68rem',
                            color: 'var(--accent-rd)',
                            fontWeight: 500,
                            paddingTop: 2,
                          }}
                        >
                          {itemError}
                        </div>
                      )}
                    </div>
                  </div>
                </div>
              );
            })
          )}
        </div>

        <div className="cart-footer">
          {/* Subtotal before cart discount */}
          <div className="cart-total-row">
            <span style={{ color: 'var(--text-2)', fontSize: '0.8125rem' }}>Subtotal</span>
            <span>{formatRp(subtotalAfterItems)}</span>
          </div>

          {/* Item-level discount row */}
          {totalItemDiscount > 0 && (
            <div className="cart-total-row" style={{ color: 'var(--accent-rd)' }}>
              <span>Diskon Item</span>
              <span>-{formatRp(totalItemDiscount)}</span>
            </div>
          )}

          {/* Cart-level discount — collapsible */}
          {cart.length > 0 && (
            <div
              style={{
                padding: '6px 0',
                borderTop: '1px dashed var(--border)',
                borderBottom: '1px dashed var(--border)',
                margin: '2px 0',
              }}
            >
              {/* Trigger button */}
              <button
                onClick={() => setCartDiscountOpen(prev => !prev)}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'space-between',
                  width: '100%',
                  padding: '6px 4px',
                  border: 'none',
                  background: 'transparent',
                  cursor: 'pointer',
                  color: 'var(--text-2)',
                  fontSize: '0.78rem',
                  fontWeight: 600,
                  borderRadius: 6,
                  transition: 'all 0.15s ease',
                }}
              >
                <span style={{ display: 'flex', alignItems: 'center', gap: 5 }}>
                  <Tag size={12} />
                  Diskon Keranjang
                  {cartDiscountAmt > 0 && (
                    <span
                      style={{
                        marginLeft: 4,
                        display: 'inline-flex',
                        alignItems: 'center',
                        padding: '2px 7px',
                        borderRadius: 5,
                        fontSize: '0.65rem',
                        fontWeight: 700,
                        background: 'rgba(239,68,68,0.12)',
                        color: 'var(--accent-rd)',
                      }}
                    >
                      -{formatRp(cartDiscountAmt)}
                    </span>
                  )}
                </span>
                <ChevronDown
                  size={14}
                  style={{
                    transition: 'transform 0.2s ease',
                    transform: cartDiscountOpen ? 'rotate(180deg)' : 'rotate(0deg)',
                    color: 'var(--text-3)',
                  }}
                />
              </button>

              {/* Collapsible panel */}
              <div
                style={{
                  overflow: 'hidden',
                  maxHeight: cartDiscountOpen ? 160 : 0,
                  opacity: cartDiscountOpen ? 1 : 0,
                  transition: 'max-height 0.28s cubic-bezier(0.4,0,0.2,1), opacity 0.2s ease',
                }}
              >
                <div
                  style={{
                    display: 'flex',
                    flexDirection: 'column',
                    gap: 6,
                    padding: '8px 4px 4px',
                  }}
                >
                  {/* Type toggle + quick buttons */}
                  <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
                    {/* % / Rp toggle group */}
                    <div
                      style={{
                        display: 'flex',
                        gap: 0,
                        background: 'var(--bg-elevated)',
                        borderRadius: 6,
                        padding: 2,
                        border: '1px solid var(--border)',
                      }}
                    >
                      {(['PERCENTAGE', 'FIXED'] as const).map(dt => (
                        <button
                          key={dt}
                          onClick={() => {
                            setCartDiscountType(dt);
                            setCartDiscountValue(0);
                          }}
                          style={{
                            flex: 1,
                            padding: '4px 10px',
                            border: 'none',
                            borderRadius: 5,
                            fontSize: '0.68rem',
                            fontWeight: 700,
                            cursor: 'pointer',
                            transition: 'all 0.15s ease',
                            background:
                              cartDiscountType === dt ? 'var(--accent-em)' : 'transparent',
                            color: cartDiscountType === dt ? '#fff' : 'var(--text-3)',
                            boxShadow:
                              cartDiscountType === dt ? '0 1px 4px rgba(8,132,246,0.3)' : 'none',
                          }}
                        >
                          {dt === 'PERCENTAGE' ? '%' : 'Rp'}
                        </button>
                      ))}
                    </div>
                    {/* Quick preset buttons */}
                    <div style={{ display: 'flex', gap: 4, flex: 1, flexWrap: 'wrap' }}>
                      {cartDiscountType === 'PERCENTAGE'
                        ? [5, 10, 15, 20].map(v => {
                            const isSelected =
                              cartDiscountType === 'PERCENTAGE' && cartDiscountValue === v;
                            return (
                              <button
                                key={v}
                                onClick={() => {
                                  setCartDiscountType('PERCENTAGE');
                                  setCartDiscountValue(v);
                                }}
                                style={{
                                  padding: '4px 8px',
                                  borderRadius: 6,
                                  fontSize: '0.68rem',
                                  fontWeight: 600,
                                  cursor: 'pointer',
                                  transition: 'all 0.12s ease',
                                  background: isSelected
                                    ? 'rgba(8,132,246,0.15)'
                                    : 'var(--bg-elevated)',
                                  color: isSelected ? 'var(--accent-em)' : 'var(--text-2)',
                                  border: isSelected
                                    ? '1px solid rgba(8,132,246,0.4)'
                                    : '1px solid var(--border)',
                                  whiteSpace: 'nowrap',
                                }}
                              >
                                -{v}%
                              </button>
                            );
                          })
                        : [5000, 10000].map(v => {
                            const isSelected =
                              cartDiscountType === 'FIXED' && cartDiscountValue === v;
                            return (
                              <button
                                key={v}
                                onClick={() => {
                                  setCartDiscountType('FIXED');
                                  setCartDiscountValue(v);
                                }}
                                style={{
                                  padding: '4px 8px',
                                  borderRadius: 6,
                                  fontSize: '0.68rem',
                                  fontWeight: 600,
                                  cursor: 'pointer',
                                  transition: 'all 0.12s ease',
                                  background: isSelected
                                    ? 'rgba(8,132,246,0.15)'
                                    : 'var(--bg-elevated)',
                                  color: isSelected ? 'var(--accent-em)' : 'var(--text-2)',
                                  border: isSelected
                                    ? '1px solid rgba(8,132,246,0.4)'
                                    : '1px solid var(--border)',
                                  whiteSpace: 'nowrap',
                                }}
                              >
                                -{v / 1000}rb
                              </button>
                            );
                          })}
                    </div>
                  </div>
                  {/* Input row */}
                  <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
                    <input
                      type="number"
                      min={0}
                      max={cartDiscountType === 'PERCENTAGE' ? 100 : undefined}
                      step={cartDiscountType === 'PERCENTAGE' ? 1 : 1000}
                      placeholder={cartDiscountType === 'PERCENTAGE' ? 'Masukkan %' : 'Masukkan Rp'}
                      value={cartDiscountValue === 0 ? '' : cartDiscountValue}
                      onChange={e => {
                        let val = parseFloat(e.target.value) || 0;
                        if (val < 0) val = 0;
                        if (cartDiscountType === 'PERCENTAGE' && val > 100) val = 100;
                        setCartDiscountValue(val);
                      }}
                      style={{
                        flex: 1,
                        minWidth: 0,
                        padding: '6px 8px',
                        borderRadius: 6,
                        border: '1px solid var(--border-md)',
                        background: 'var(--bg-elevated)',
                        color: 'var(--text-1)',
                        fontSize: '0.78rem',
                        fontWeight: 600,
                        outline: 'none',
                      }}
                    />
                    {cartDiscountValue > 0 && (
                      <button
                        onClick={() => setCartDiscountValue(0)}
                        style={{
                          padding: '5px 10px',
                          borderRadius: 5,
                          fontSize: '0.68rem',
                          fontWeight: 600,
                          cursor: 'pointer',
                          border: 'none',
                          background: 'rgba(239,68,68,0.1)',
                          color: 'var(--accent-rd)',
                          transition: 'all 0.12s ease',
                          whiteSpace: 'nowrap',
                        }}
                      >
                        Hapus
                      </button>
                    )}
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* PPN */}
          <div className="cart-total-row">
            <span style={{ color: 'var(--text-2)', fontSize: '0.8125rem' }}>PPN</span>
            <span>{formatRp(taxAmt)}</span>
          </div>

          {/* Grand total */}
          <div className="cart-total-row grand">
            <span>Total</span>
            <span style={{ color: 'var(--brand)' }}>{formatRp(total)}</span>
          </div>

          {/* Savings banner */}
          {totalDiscount > 0 && (
            <div
              style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
                gap: 8,
                padding: '8px 12px',
                borderRadius: 8,
                background: 'linear-gradient(135deg, rgba(16,185,129,0.1), rgba(16,185,129,0.04))',
                border: '1px solid rgba(16,185,129,0.2)',
              }}
            >
              <span
                style={{
                  fontSize: '0.75rem',
                  fontWeight: 600,
                  color: '#10b981',
                  display: 'flex',
                  alignItems: 'center',
                  gap: 5,
                }}
              >
                💰 Anda hemat
              </span>
              <span style={{ fontSize: '0.82rem', fontWeight: 800, color: '#10b981' }}>
                {formatRp(totalDiscount)}
              </span>
            </div>
          )}

          {/* Hold error */}
          {holdError && (
            <div style={{ fontSize: '0.8rem', color: '#f87171', padding: '4px 0' }}>
              {holdError}
            </div>
          )}

          {/* Loyalty points preview */}
          {selectedCustomer && cart.length > 0 && (
            <div
              style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
                padding: '6px 10px',
                background: 'rgba(251,191,36,0.08)',
                border: '1px solid rgba(251,191,36,0.25)',
                borderRadius: 7,
                marginBottom: 4,
              }}
            >
              <div
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 5,
                  fontSize: '0.78rem',
                  color: '#f59e0b',
                }}
              >
                <Star size={12} style={{ fill: '#f59e0b', color: '#f59e0b' }} />
                Poin yang akan didapat
              </div>
              <span style={{ fontWeight: 700, fontSize: '0.82rem', color: '#f59e0b' }}>
                +
                {loyalty
                  .previewEarnings(total, undefined, selectedStore?.loyalty_points_per_rupiah)
                  .toLocaleString('id-ID')}{' '}
                pts
              </span>
            </div>
          )}

          {/* Restaurant dine-in: Hold + Pay buttons */}
          {isRestaurant && selectedTable ? (
            <div style={{ display: 'flex', gap: 8 }}>
              <button
                className="btn btn-secondary"
                style={{ flex: 1, gap: 6 }}
                disabled={cart.length === 0 || holdLoading}
                onClick={handleHoldOrder}
              >
                {holdLoading ? <Loader2 size={14} className="loading-spin" /> : <Clock size={14} />}
                {activeDraft ? 'Perbarui' : 'Tahan'}
              </button>
              <button
                className="checkout-btn"
                style={{ flex: 2 }}
                disabled={cart.length === 0}
                onClick={() => setShowPayment(true)}
              >
                <CheckCircle2 size={16} />
                Bayar {cart.length > 0 ? formatRp(total) : ''}
              </button>
            </div>
          ) : (
            /* Retail + Take Away: single Pay button */
            <button
              className="checkout-btn"
              disabled={cart.length === 0}
              onClick={() => setShowPayment(true)}
            >
              <CheckCircle2 size={18} />
              {isRestaurant ? 'Proses Pesanan' : 'Bayar'} {cart.length > 0 ? formatRp(total) : ''}
            </button>
          )}
        </div>
      </div>

      {/* ── Modals ── */}
      {showPayment && (
        <Portal>
          <PaymentModal
            total={total}
            onClose={() => setShowPayment(false)}
            onConfirm={handleConfirmPayment}
            loading={payLoading}
          />
        </Portal>
      )}
      {receipt && (
        <Portal>
          <ReceiptModal txn={receipt} onClose={() => setReceipt(null)} />
        </Portal>
      )}
    </div>
  );
}
