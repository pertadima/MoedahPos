'use client';

import { useEffect, useReducer, useState, useCallback } from 'react';
import { Search, ShoppingCart, Trash2, Printer, X, Minus, Plus, Loader2, CheckCircle2 } from 'lucide-react';
import { useAuth } from '@/lib/auth/AuthContext';
import { productsApi } from '@/lib/api/products';
import { transactionsApi } from '@/lib/api/transactions';
import { formatRp, productEmoji } from '@/lib/utils';
import type { Product, Category, CartItem, Transaction } from '@/types';
import { ApiError } from '@/lib/api/client';

// ── Cart Reducer ─────────────────────────────────────────────────────────────
type CartAction =
  | { type: 'ADD'; product: Product }
  | { type: 'REMOVE'; productId: string }
  | { type: 'SET_QTY'; productId: string; qty: number }
  | { type: 'CLEAR' };

function makeCartItem(product: Product, qty = 1): CartItem {
  const net = product.sell_price * qty;
  const tax = net * (product.tax_rate / 100);
  return { product, quantity: qty, discount_pct: 0, unitPrice: product.sell_price, subtotal: net, taxAmt: tax };
}
function cartReducer(state: CartItem[], action: CartAction): CartItem[] {
  switch (action.type) {
    case 'ADD': {
      const idx = state.findIndex(i => i.product.id === action.product.id);
      if (idx >= 0) {
        return state.map((item, i) => i === idx ? makeCartItem(item.product, item.quantity + 1) : item);
      }
      return [...state, makeCartItem(action.product)];
    }
    case 'REMOVE': return state.filter(i => i.product.id !== action.productId);
    case 'SET_QTY': {
      if (action.qty < 1) return state.filter(i => i.product.id !== action.productId);
      return state.map(i => i.product.id === action.productId ? makeCartItem(i.product, action.qty) : i);
    }
    case 'CLEAR': return [];
    default: return state;
  }
}

// ── Payment Modal ─────────────────────────────────────────────────────────────
function PaymentModal({
  total, onClose, onConfirm, loading
}: {
  total: number; onClose: () => void;
  onConfirm: (method: string, amount: number) => void; loading: boolean;
}) {
  const methods = ['cash', 'qris', 'card', 'transfer'];
  const [method, setMethod] = useState('cash');
  const [amountStr, setAmountStr] = useState('');

  const amount = parseFloat(amountStr) || 0;
  const change = amount - total;
  const canConfirm = method !== 'cash' || amount >= total;

  const numPress = (val: string) => {
    if (val === '⌫') { setAmountStr(s => s.slice(0, -1)); return; }
    if (val === '000') { setAmountStr(s => (s + '000').replace(/^0+(\d)/, '$1')); return; }
    setAmountStr(s => (s + val).replace(/^0+(\d)/, '$1'));
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-box" onClick={e => e.stopPropagation()}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 20 }}>
          <h2 style={{ fontWeight: 700, fontSize: '1.1rem' }}>Proses Pembayaran</h2>
          <button onClick={onClose} className="btn btn-ghost btn-sm"><X size={16} /></button>
        </div>

        {/* Total */}
        <div style={{ background: 'rgba(16,185,129,0.1)', borderRadius: 10, padding: '14px 18px', marginBottom: 18, textAlign: 'center' }}>
          <div style={{ fontSize: '0.8rem', color: 'var(--text-2)', marginBottom: 2 }}>Total Pembayaran</div>
          <div style={{ fontSize: '1.8rem', fontWeight: 800, color: 'var(--accent-em)' }}>{formatRp(total)}</div>
        </div>

        {/* Method */}
        <div style={{ marginBottom: 16 }}>
          <div style={{ fontSize: '0.78rem', color: 'var(--text-3)', marginBottom: 8, textTransform: 'uppercase', letterSpacing: '0.05em' }}>Metode Bayar</div>
          <div className="pay-method-tabs">
            {methods.map(m => (
              <button key={m} className={`pay-tab ${method === m ? 'active' : ''}`} onClick={() => { setMethod(m); if (m !== 'cash') setAmountStr(String(total)); }}>
                {m.toUpperCase()}
              </button>
            ))}
          </div>
        </div>

        {/* Cash numpad */}
        {method === 'cash' && (
          <>
            <div style={{ marginBottom: 8 }}>
              <div style={{ fontSize: '0.78rem', color: 'var(--text-3)', marginBottom: 6 }}>Jumlah Diterima</div>
              <div style={{
                background: 'var(--bg-elevated)', borderRadius: 8, padding: '10px 14px',
                fontSize: '1.4rem', fontWeight: 700, color: 'var(--text-1)',
                minHeight: 50, display: 'flex', alignItems: 'center',
                border: '1px solid var(--border-md)',
              }}>
                {amountStr ? formatRp(amount) : <span style={{ color: 'var(--text-3)' }}>Masukkan jumlah</span>}
              </div>
            </div>

            <div className="numpad">
              {['7','8','9','4','5','6','1','2','3','000','0','⌫'].map(k => (
                <button key={k} className="num-btn" onClick={() => numPress(k)}
                  style={k === '⌫' ? { color: 'var(--accent-rd)' } : undefined}>
                  {k}
                </button>
              ))}
            </div>

            {amount > 0 && (
              <div style={{ display: 'flex', justifyContent: 'space-between', marginTop: 12, padding: '10px 14px', background: change >= 0 ? 'rgba(16,185,129,0.08)' : 'rgba(239,68,68,0.08)', borderRadius: 8 }}>
                <span style={{ color: 'var(--text-2)', fontSize: '0.9rem' }}>Kembalian</span>
                <span style={{ fontWeight: 700, color: change >= 0 ? 'var(--accent-em)' : 'var(--accent-rd)', fontSize: '0.9rem' }}>
                  {change >= 0 ? formatRp(change) : '⚠ Kurang ' + formatRp(-change)}
                </span>
              </div>
            )}
          </>
        )}

        {method !== 'cash' && (
          <div style={{ padding: '20px', textAlign: 'center', color: 'var(--text-2)', fontSize: '0.9rem' }}>
            Konfirmasi pembayaran {method.toUpperCase()} sebesar <strong style={{ color: 'var(--text-1)' }}>{formatRp(total)}</strong>
          </div>
        )}

        <button
          className="btn btn-primary btn-lg" style={{ width: '100%', marginTop: 16 }}
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
  const handlePrint = () => window.print();

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-box" style={{ maxWidth: 340 }} onClick={e => e.stopPropagation()}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
          <h2 style={{ fontWeight: 700, fontSize: '1rem' }}>Struk Pembayaran</h2>
          <button onClick={onClose} className="btn btn-ghost btn-sm"><X size={16} /></button>
        </div>

        {/* Receipt content (also printed) */}
        <div id="receipt-content" style={{ fontFamily: "'Courier New', monospace", fontSize: '0.82rem', lineHeight: 1.6 }}>
          <div style={{ textAlign: 'center', marginBottom: 12, borderBottom: '1px dashed var(--border-md)', paddingBottom: 10 }}>
            <div style={{ fontWeight: 800, fontSize: '1rem' }}>MoedahPOS</div>
            <div style={{ color: 'var(--text-2)', fontSize: '0.75rem' }}>{new Date(txn.created_at).toLocaleString('id-ID')}</div>
            <div style={{ color: 'var(--text-2)', fontSize: '0.75rem' }}>Kasir: {txn.cashier_name}</div>
            {txn.customer_name && <div style={{ color: 'var(--text-2)', fontSize: '0.75rem' }}>Pelanggan: {txn.customer_name}</div>}
          </div>

          {txn.items.map((item, i) => (
            <div key={i} style={{ marginBottom: 6 }}>
              <div style={{ fontWeight: 600 }}>{item.product_name}</div>
              <div style={{ display: 'flex', justifyContent: 'space-between', color: 'var(--text-2)', fontSize: '0.78rem' }}>
                <span>{item.quantity} x {formatRp(item.unit_price)}</span>
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

// ── Main POS Page ─────────────────────────────────────────────────────────────
export default function POSPage() {
  const { selectedStore } = useAuth();
  const [products, setProducts] = useState<Product[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [cart, dispatch] = useReducer(cartReducer, []);
  const [search, setSearch] = useState('');
  const [activeCat, setActiveCat] = useState('all');
  const [loading, setLoading] = useState(true);
  const [showPayment, setShowPayment] = useState(false);
  const [payLoading, setPayLoading] = useState(false);
  const [receipt, setReceipt] = useState<Transaction | null>(null);
  const [error, setError] = useState('');

  const storeId = selectedStore?.store_id;

  useEffect(() => {
    if (!storeId) return;
    setLoading(true);
    Promise.all([
      productsApi.list(storeId, { per_page: 200 }),
      productsApi.listCategories(storeId),
    ]).then(([p, c]) => {
      setProducts((p.data as any).data ?? []);
      setCategories(c.data as Category[]);
    }).catch(console.error)
      .finally(() => setLoading(false));
  }, [storeId]);

  const filtered = products.filter(p => {
    if (!p.is_active) return false;
    const matchSearch = p.name.toLowerCase().includes(search.toLowerCase()) ||
                        p.sku.toLowerCase().includes(search.toLowerCase());
    const matchCat = activeCat === 'all' || p.category_id === activeCat;
    return matchSearch && matchCat;
  });

  // Totals
  const subtotal = cart.reduce((s, i) => s + i.subtotal, 0);
  const taxAmt = cart.reduce((s, i) => s + i.taxAmt, 0);
  const total = subtotal + taxAmt;
  const itemCount = cart.reduce((s, i) => s + i.quantity, 0);

  const handleConfirmPayment = useCallback(async (method: string, amount: number) => {
    if (!storeId) return;
    setPayLoading(true);
    setError('');
    try {
      const res = await transactionsApi.checkout(storeId, {
        payment_method: method,
        payment_amount: amount,
        items: cart.map(i => ({ product_id: i.product.id, quantity: i.quantity, discount_pct: 0 })),
      });
      setReceipt(res.data as Transaction);
      dispatch({ type: 'CLEAR' });
      setShowPayment(false);
    } catch (err) {
      if (err instanceof ApiError) setError(err.message);
      else setError('Gagal memproses pembayaran');
    } finally {
      setPayLoading(false);
    }
  }, [storeId, cart]);

  if (!selectedStore) {
    return (
      <div style={{ marginLeft: 220, display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100vh' }}>
        <div className="empty-state"><ShoppingCart size={40} /><p>Pilih toko di sidebar untuk memulai</p></div>
      </div>
    );
  }

  return (
    <div className="pos-layout">
      {/* ── LEFT: Product Catalog ── */}
      <div className="pos-catalog">
        {/* Search */}
        <div style={{ position: 'relative' }}>
          <Search size={16} style={{ position: 'absolute', left: 12, top: '50%', transform: 'translateY(-50%)', color: 'var(--text-3)' }} />
          <input
            className="input" style={{ paddingLeft: 36 }}
            placeholder="Cari produk atau SKU..."
            value={search} onChange={e => setSearch(e.target.value)}
          />
        </div>

        {/* Category Tabs */}
        <div className="category-tabs">
          <button className={`cat-tab ${activeCat === 'all' ? 'active' : ''}`} onClick={() => setActiveCat('all')}>
            Semua
          </button>
          {categories.map(c => (
            <button key={c.id} className={`cat-tab ${activeCat === c.id ? 'active' : ''}`} onClick={() => setActiveCat(c.id)}>
              {c.name}
            </button>
          ))}
        </div>

        {/* Error */}
        {error && (
          <div style={{ background: 'rgba(239,68,68,0.12)', border: '1px solid rgba(239,68,68,0.3)', borderRadius: 8, padding: '10px 14px', color: '#f87171', fontSize: '0.85rem', display: 'flex', justifyContent: 'space-between' }}>
            {error}
            <button onClick={() => setError('')} style={{ background: 'none', border: 'none', color: '#f87171', cursor: 'pointer' }}><X size={14} /></button>
          </div>
        )}

        {/* Product Grid */}
        {loading ? (
          <div style={{ display: 'flex', justifyContent: 'center', padding: 48 }}>
            <Loader2 size={28} className="loading-spin" style={{ color: 'var(--accent-em)' }} />
          </div>
        ) : filtered.length === 0 ? (
          <div className="empty-state">
            <Search size={32} />
            <p>Tidak ada produk ditemukan</p>
          </div>
        ) : (
          <div className="product-grid">
            {filtered.map(p => {
              const inCart = cart.find(i => i.product.id === p.id);
              const outOfStock = (p.stock_qty ?? 1) <= 0;
              return (
                <div
                  key={p.id}
                  className={`product-card ${outOfStock ? 'out-of-stock' : ''}`}
                  onClick={() => !outOfStock && dispatch({ type: 'ADD', product: p })}
                >
                  {inCart && (
                    <div style={{
                      position: 'absolute', top: 6, right: 6,
                      background: 'var(--accent-em)', color: '#fff',
                      borderRadius: '50%', width: 20, height: 20,
                      display: 'flex', alignItems: 'center', justifyContent: 'center',
                      fontSize: '0.7rem', fontWeight: 700,
                    }}>
                      {inCart.quantity}
                    </div>
                  )}
                  <div className="product-icon">{productEmoji(p.name)}</div>
                  <div className="product-name">{p.name}</div>
                  <div className="product-sku">{p.sku}</div>
                  <div className="product-price">{formatRp(p.sell_price)}</div>
                  <div className="product-stock">
                    {outOfStock
                      ? <span className="badge badge-red">Habis</span>
                      : p.stock_qty !== undefined && p.stock_qty <= 5
                        ? <span className="badge badge-amber">{p.stock_qty} {p.unit}</span>
                        : <span className="badge badge-green">{p.stock_qty ?? '–'} {p.unit}</span>
                    }
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>

      {/* ── RIGHT: Cart ── */}
      <div className="pos-cart">
        {/* Cart Header */}
        <div className="cart-header">
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
              <ShoppingCart size={18} style={{ color: 'var(--accent-em)' }} />
              <span style={{ fontWeight: 700, fontSize: '0.95rem' }}>Keranjang</span>
              {itemCount > 0 && (
                <span className="badge badge-green">{itemCount} item</span>
              )}
            </div>
            {cart.length > 0 && (
              <button className="btn btn-ghost btn-sm" onClick={() => dispatch({ type: 'CLEAR' })}>
                <Trash2 size={13} /> Kosongkan
              </button>
            )}
          </div>
        </div>

        {/* Cart Items */}
        <div className="cart-items">
          {cart.length === 0 ? (
            <div className="empty-state" style={{ paddingTop: 48 }}>
              <ShoppingCart size={36} style={{ color: 'var(--text-3)' }} />
              <p style={{ fontSize: '0.85rem' }}>Klik produk untuk menambahkan</p>
            </div>
          ) : (
            cart.map(item => (
              <div key={item.product.id} className="cart-item">
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                  <div className="cart-item-name" style={{ flex: 1, paddingRight: 8 }}>{item.product.name}</div>
                  <button className="btn btn-ghost btn-sm" style={{ padding: '2px 4px', color: 'var(--accent-rd)' }}
                    onClick={() => dispatch({ type: 'REMOVE', productId: item.product.id })}>
                    <X size={13} />
                  </button>
                </div>
                <div className="cart-item-row">
                  <div className="qty-ctrl">
                    <button className="qty-btn" onClick={() => dispatch({ type: 'SET_QTY', productId: item.product.id, qty: item.quantity - 1 })}>
                      <Minus size={12} />
                    </button>
                    <span className="qty-val">{item.quantity}</span>
                    <button className="qty-btn" onClick={() => dispatch({ type: 'SET_QTY', productId: item.product.id, qty: item.quantity + 1 })}>
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

        {/* Cart Footer */}
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
          <button
            className="checkout-btn"
            disabled={cart.length === 0}
            onClick={() => setShowPayment(true)}
          >
            <CheckCircle2 size={18} />
            Bayar {cart.length > 0 ? formatRp(total) : ''}
          </button>
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
      {receipt && (
        <ReceiptModal txn={receipt} onClose={() => setReceipt(null)} />
      )}

      {/* Hidden print DOM */}
      {receipt && (
        <div id="receipt-root" style={{ display: 'none' }}>
          <div id="receipt-content-print" style={{ fontFamily: 'Courier New, monospace', fontSize: 12, color: '#000' }}>
            {/* Mirrors receipt content for print CSS targeting */}
          </div>
        </div>
      )}
    </div>
  );
}
