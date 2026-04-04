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
  Coffee,
  Clock,
  Users,
} from 'lucide-react';
import { useAuth } from '@/lib/auth/AuthContext';
import { productsApi } from '@/lib/api/products';
import { menuItemsApi, customersApi, tablesApi } from '@/lib/api/store-apis';
import { transactionsApi } from '@/lib/api/transactions';
import { formatRp, productEmoji } from '@/lib/utils';
import type {
  Product,
  Category,
  CartItem,
  Transaction,
  MenuItem,
  Customer,
  RestaurantTable,
} from '@/types';
import { ApiError } from '@/lib/api/client';

// ── Unified cart item type ────────────────────────────────────────────────────
// We re-use CartItem but may hold either a product or a menu item.
// menuItemId is set for restaurant items.
interface PosCartItem extends CartItem {
  menuItemId?: string;
}

type CartAction =
  | { type: 'ADD_PRODUCT'; product: Product }
  | { type: 'ADD_MENU'; item: MenuItem }
  | { type: 'REMOVE'; id: string }
  | { type: 'SET_QTY'; id: string; qty: number }
  | { type: 'CLEAR' };

function makeFromProduct(product: Product, qty = 1): PosCartItem {
  const net = product.sell_price * qty;
  const tax = net * (product.tax_rate / 100);
  return {
    product,
    quantity: qty,
    discount_pct: 0,
    unitPrice: product.sell_price,
    subtotal: net,
    taxAmt: tax,
  };
}

function makeFromMenu(item: MenuItem, qty = 1): PosCartItem {
  const price = item.sell_price;
  const taxRate = item.tax_rate ?? 0;
  const net = price * qty;
  const tax = net * (taxRate / 100);
  // Coerce MenuItem into Product shape enough for the cart
  const fakeProduct: Product = {
    id: item.id,
    name: item.name,
    sku: 'MENU',
    unit: 'porsi',
    description: '',
    sell_price: price,
    cost_price: 0,
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
    unitPrice: price,
    subtotal: net,
    taxAmt: tax,
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
        if (i.menuItemId) {
          const net = i.unitPrice * action.qty;
          const tax = net * (i.product.tax_rate / 100);
          return { ...i, quantity: action.qty, subtotal: net, taxAmt: tax };
        }
        return makeFromProduct(i.product, action.qty);
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
  const { selectedStore } = useAuth();
  const isRestaurant = selectedStore?.store_type === 'restaurant';

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

  // Customer picker
  const [selectedCustomer, setSelectedCustomer] = useState<Customer | null>(null);
  const [custSearch, setCustSearch] = useState('');
  const [custResults, setCustResults] = useState<Customer[]>([]);
  const [custOpen, setCustOpen] = useState(false);
  const custRef = useRef<HTMLDivElement>(null);

  const storeId = selectedStore?.store_id;

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
          const tbls: RestaurantTable[] = (tRes.data as any) ?? [];
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
          const items: MenuItem[] = (mRes.data as any) ?? [];
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
          setProducts((p.data as any).data ?? []);
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
          // Restore cart from draft items (map back to PosCartItem)
          draft.items.forEach(item => {
            const fakeMenuItem: MenuItem = {
              id: item.product_id ?? item.id,
              store_id: storeId,
              name: item.product_name,
              description: '',
              sell_price: item.unit_price,
              tax_rate: item.tax_rate,
              is_active: true,
              category_name: '',
              ingredients: [],
              created_at: '',
              updated_at: '',
            };
            dispatch({ type: 'ADD_MENU', item: fakeMenuItem });
            // Set the correct quantity
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
        const tbls: RestaurantTable[] = (res.data as any) ?? [];
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
        discount_pct: 0,
      }));
      if (activeDraft) {
        await transactionsApi.updateDraft(storeId, activeDraft.id, { items });
      } else {
        await transactionsApi.createDraft(storeId, { table_id: selectedTable.id, items });
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
  }, [storeId, selectedTable, activeDraft, cart, handleBackToTables]);

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

  // ── Totals ──────────────────────────────────────────────────────────────────
  const subtotal = cart.reduce((s, i) => s + i.subtotal, 0);
  const taxAmt = cart.reduce((s, i) => s + i.taxAmt, 0);
  const total = subtotal + taxAmt;
  const itemCount = cart.reduce((s, i) => s + i.quantity, 0);

  // ── Checkout (retail + restaurant direct pay) ────────────────────────────────
  const handleConfirmPayment = useCallback(
    async (method: string, amount: number) => {
      if (!storeId) return;
      setPayLoading(true);
      setError('');
      try {
        const items = (cart as PosCartItem[]).map(i => ({
          product_id: i.menuItemId ? '' : i.product.id,
          menu_item_id: i.menuItemId ?? '',
          quantity: i.quantity,
          discount_pct: 0,
        }));
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
        } else {
          res = await transactionsApi.checkout(storeId, {
            payment_method: method,
            payment_amount: amount,
            customer_name: selectedCustomer?.name ?? '',
            customer_phone: selectedCustomer?.phone ?? '',
            items,
          });
        }
        setReceipt(res.data as Transaction);
        setSelectedCustomer(null);
        setCustSearch('');
        setShowPayment(false);
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
    [storeId, cart, selectedCustomer, isRestaurant, activeDraft, selectedTable, handleBackToTables]
  );

  if (!selectedStore) {
    return (
      <div
        style={{
          marginLeft: 220,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          height: '100vh',
        }}
      >
        <div className="empty-state">
          <ShoppingCart size={40} />
          <p>Pilih toko di sidebar untuk memulai</p>
        </div>
      </div>
    );
  }

  // ── Restaurant: table selection screen ─────────────────────────────────────
  if (isRestaurant && !selectedTable && !isTakeAway) {
    return (
      <div style={{ marginLeft: 220, padding: '24px 28px', minHeight: '100vh' }}>
        {/* Header */}
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            marginBottom: 24,
          }}
        >
          <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
            <UtensilsCrossed size={22} style={{ color: 'var(--brand)' }} />
            <div>
              <h1 style={{ fontSize: '1.3rem', fontWeight: 800, margin: 0 }}>Pilih Meja</h1>
              <p style={{ color: 'var(--text-2)', fontSize: '0.85rem', margin: 0 }}>
                {selectedStore.store_name}
              </p>
            </div>
          </div>

          {/* Take Away shortcut button */}
          <button
            onClick={() => {
              setIsTakeAway(true);
              dispatch({ type: 'CLEAR' });
              setError('');
              setHoldError('');
            }}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: 8,
              background: 'linear-gradient(135deg, rgba(255,167,36,0.15), rgba(255,167,36,0.05))',
              border: '1.5px solid rgba(255,167,36,0.5)',
              borderRadius: 12,
              padding: '10px 18px',
              cursor: 'pointer',
              fontWeight: 700,
              fontSize: '0.9rem',
              color: '#FFA724',
              boxShadow: '0 2px 12px rgba(255,167,36,0.15)',
              transition: 'all 0.18s',
            }}
          >
            <ShoppingBag size={16} />
            Take Away
          </button>
        </div>

        {tablesLoading ? (
          <div style={{ display: 'flex', justifyContent: 'center', padding: 64 }}>
            <Loader2 size={32} className="loading-spin" style={{ color: 'var(--brand)' }} />
          </div>
        ) : tables.length === 0 ? (
          <div className="empty-state">
            <Coffee size={40} />
            <p>Belum ada meja. Tambahkan meja di menu Manajemen Meja.</p>
          </div>
        ) : (
          <div
            style={{
              display: 'grid',
              gridTemplateColumns: 'repeat(auto-fill, minmax(160px, 1fr))',
              gap: 16,
            }}
          >
            {tables.map(table => {
              const draft = tablesDraftMap[table.id];
              const hasOrder = !!draft;
              const isOccupied = table.status === 'occupied' || hasOrder;
              const isUnavailable = table.status === 'unavailable';

              return (
                <button
                  key={table.id}
                  onClick={() => !isUnavailable && handleSelectTable(table)}
                  style={{
                    background: isUnavailable
                      ? 'var(--bg-elevated)'
                      : isOccupied
                        ? 'linear-gradient(135deg, rgba(255,167,36,0.15), rgba(255,167,36,0.05))'
                        : 'linear-gradient(135deg, rgba(8,132,246,0.12), rgba(8,132,246,0.04))',
                    border: isUnavailable
                      ? '1.5px solid var(--border)'
                      : isOccupied
                        ? '1.5px solid rgba(255,167,36,0.5)'
                        : '1.5px solid rgba(8,132,246,0.35)',
                    borderRadius: 16,
                    padding: '20px 16px',
                    cursor: isUnavailable ? 'not-allowed' : 'pointer',
                    opacity: isUnavailable ? 0.5 : 1,
                    textAlign: 'center',
                    transition: 'all 0.18s',
                    display: 'flex',
                    flexDirection: 'column',
                    alignItems: 'center',
                    gap: 8,
                    position: 'relative',
                  }}
                >
                  {/* Status dot */}
                  <div
                    style={{
                      width: 10,
                      height: 10,
                      borderRadius: '50%',
                      background: isUnavailable
                        ? 'var(--text-3)'
                        : isOccupied
                          ? '#FFA724'
                          : '#22c55e',
                      position: 'absolute',
                      top: 12,
                      right: 12,
                      boxShadow:
                        isOccupied && !isUnavailable ? '0 0 8px rgba(255,167,36,0.6)' : 'none',
                    }}
                  />

                  <div style={{ fontSize: '2rem' }}>🪑</div>
                  <div style={{ fontWeight: 800, fontSize: '1.05rem', color: 'var(--text-1)' }}>
                    Meja {table.table_number}
                  </div>
                  <div
                    style={{
                      fontSize: '0.75rem',
                      color: 'var(--text-2)',
                      display: 'flex',
                      alignItems: 'center',
                      gap: 4,
                    }}
                  >
                    <Users size={11} /> {table.capacity} kursi
                  </div>

                  {hasOrder && draft ? (
                    <div
                      style={{
                        background: 'rgba(255,167,36,0.18)',
                        borderRadius: 8,
                        padding: '4px 8px',
                        fontSize: '0.72rem',
                        color: '#FFA724',
                        fontWeight: 600,
                        display: 'flex',
                        alignItems: 'center',
                        gap: 4,
                      }}
                    >
                      <Clock size={10} />
                      {draft.items.length} item · {formatRp(draft.total)}
                    </div>
                  ) : (
                    <div
                      style={{
                        fontSize: '0.72rem',
                        fontWeight: 600,
                        color: isUnavailable ? 'var(--text-3)' : 'var(--brand)',
                      }}
                    >
                      {isUnavailable ? 'Tidak Tersedia' : 'Tersedia'}
                    </div>
                  )}
                </button>
              );
            })}
          </div>
        )}
      </div>
    );
  }

  // ── Render (Retail + Restaurant order screen) ───────────────────────────────
  return (
    <div className="pos-layout">
      {/* ── LEFT: Catalog ── */}
      <div className="pos-catalog">
        {/* Mode badge + table back button (restaurant only) */}
        <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 10 }}>
          {isRestaurant && (selectedTable || isTakeAway) && (
            <button
              onClick={handleBackToTables}
              className="btn btn-ghost btn-sm"
              style={{ padding: '4px 8px', gap: 4 }}
            >
              <ArrowLeft size={13} /> Meja
            </button>
          )}
          <div
            style={{
              display: 'inline-flex',
              alignItems: 'center',
              gap: 5,
              padding: '4px 10px',
              borderRadius: 8,
              fontSize: '0.75rem',
              fontWeight: 600,
              background: isRestaurant ? 'rgba(251,146,60,0.12)' : 'rgba(16,185,129,0.12)',
              color: isRestaurant ? '#fb923c' : '#10b981',
            }}
          >
            {isRestaurant ? <UtensilsCrossed size={13} /> : <ShoppingBag size={13} />}
            {isRestaurant
              ? isTakeAway
                ? 'Take Away — Pesanan'
                : `Meja ${selectedTable?.table_number ?? ''} — Pesanan`
              : 'Mode Retail — Produk'}
          </div>
        </div>

        {/* Search */}
        <div style={{ position: 'relative', marginBottom: 10 }}>
          <Search
            size={16}
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
            placeholder={isRestaurant ? 'Cari menu...' : 'Cari produk atau SKU...'}
            value={search}
            onChange={e => setSearch(e.target.value)}
          />
        </div>

        {/* Category Tabs */}
        <div className="category-tabs">
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
              {filteredMenuItems.map(m => {
                const inCart = cart.find(i => i.menuItemId === m.id);
                return (
                  <div
                    key={m.id}
                    className="product-card"
                    onClick={() => dispatch({ type: 'ADD_MENU', item: m })}
                  >
                    {inCart && (
                      <div
                        style={{
                          position: 'absolute',
                          top: 6,
                          right: 6,
                          background: '#fb923c',
                          color: '#fff',
                          borderRadius: '50%',
                          width: 20,
                          height: 20,
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'center',
                          fontSize: '0.7rem',
                          fontWeight: 700,
                        }}
                      >
                        {inCart.quantity}
                      </div>
                    )}
                    <div className="product-icon">🍽️</div>
                    <div className="product-name">{m.name}</div>
                    {m.category_name && <div className="product-sku">{m.category_name}</div>}
                    <div className="product-price">{formatRp(m.sell_price)}</div>
                    {m.description && (
                      <div
                        style={{
                          fontSize: '0.68rem',
                          color: 'var(--text-3)',
                          lineHeight: 1.3,
                          marginTop: 2,
                          overflow: 'hidden',
                          display: '-webkit-box',
                          WebkitLineClamp: 2,
                          WebkitBoxOrient: 'vertical',
                        }}
                      >
                        {m.description}
                      </div>
                    )}
                    {m.ingredients && m.ingredients.length > 0 && (
                      <div style={{ fontSize: '0.65rem', color: 'var(--text-3)', marginTop: 3 }}>
                        🧂 {m.ingredients.length} bahan
                      </div>
                    )}
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
            {filteredProducts.map(p => {
              const inCart = cart.find(i => i.product.id === p.id && !i.menuItemId);
              const outOfStock = (p.stock_qty ?? 1) <= 0;
              return (
                <div
                  key={p.id}
                  className={`product-card ${outOfStock ? 'out-of-stock' : ''}`}
                  onClick={() => !outOfStock && dispatch({ type: 'ADD_PRODUCT', product: p })}
                >
                  {inCart && (
                    <div
                      style={{
                        position: 'absolute',
                        top: 6,
                        right: 6,
                        background: 'var(--accent-em)',
                        color: '#fff',
                        borderRadius: '50%',
                        width: 20,
                        height: 20,
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        fontSize: '0.7rem',
                        fontWeight: 700,
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
                </div>
              );
            })}
          </div>
        )}
      </div>

      {/* ── RIGHT: Cart ── */}
      <div className="pos-cart">
        <div className="cart-header">
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
              <ShoppingCart size={18} style={{ color: 'var(--accent-em)' }} />
              <span style={{ fontWeight: 700, fontSize: '0.95rem' }}>
                {isRestaurant ? 'Pesanan' : 'Keranjang'}
              </span>
              {itemCount > 0 && <span className="badge badge-green">{itemCount} item</span>}
            </div>
            {cart.length > 0 && (
              <button className="btn btn-ghost btn-sm" onClick={() => dispatch({ type: 'CLEAR' })}>
                <Trash2 size={13} /> Kosongkan
              </button>
            )}
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
            (cart as PosCartItem[]).map(item => (
              <div key={item.product.id} className="cart-item">
                <div
                  style={{
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'flex-start',
                  }}
                >
                  <div className="cart-item-name" style={{ flex: 1, paddingRight: 8 }}>
                    {item.menuItemId && (
                      <span style={{ fontSize: '0.65rem', color: '#fb923c', marginRight: 4 }}>
                        🍽
                      </span>
                    )}
                    {item.product.name}
                  </div>
                  <button
                    className="btn btn-ghost btn-sm"
                    style={{ padding: '2px 4px', color: 'var(--accent-rd)' }}
                    onClick={() => dispatch({ type: 'REMOVE', id: item.product.id })}
                  >
                    <X size={13} />
                  </button>
                </div>
                <div className="cart-item-row">
                  <div className="qty-ctrl">
                    <button
                      className="qty-btn"
                      onClick={() =>
                        dispatch({ type: 'SET_QTY', id: item.product.id, qty: item.quantity - 1 })
                      }
                    >
                      <Minus size={12} />
                    </button>
                    <span className="qty-val">{item.quantity}</span>
                    <button
                      className="qty-btn"
                      onClick={() =>
                        dispatch({ type: 'SET_QTY', id: item.product.id, qty: item.quantity + 1 })
                      }
                    >
                      <Plus size={12} />
                    </button>
                  </div>
                  <span style={{ fontWeight: 700, fontSize: '0.85rem', color: 'var(--accent-em)' }}>
                    {formatRp(item.subtotal)}
                  </span>
                </div>
                <div style={{ fontSize: '0.72rem', color: 'var(--text-3)' }}>
                  {formatRp(item.unitPrice)} × {item.quantity}
                  {item.product.tax_rate > 0 && ` · PPN ${item.product.tax_rate}%`}
                </div>
              </div>
            ))
          )}
        </div>

        <div className="cart-footer">
          <div className="cart-total-row">
            <span className="text-2">Subtotal</span>
            <span>{formatRp(subtotal)}</span>
          </div>
          <div className="cart-total-row">
            <span className="text-2">PPN</span>
            <span>{formatRp(taxAmt)}</span>
          </div>
          <div className="cart-total-row grand">
            <span>Total</span>
            <span style={{ color: 'var(--accent-em)' }}>{formatRp(total)}</span>
          </div>

          {/* Hold error */}
          {holdError && (
            <div style={{ fontSize: '0.8rem', color: '#f87171', padding: '4px 0' }}>
              {holdError}
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
        <PaymentModal
          total={total}
          onClose={() => setShowPayment(false)}
          onConfirm={handleConfirmPayment}
          loading={payLoading}
        />
      )}
      {receipt && <ReceiptModal txn={receipt} onClose={() => setReceipt(null)} />}
    </div>
  );
}
