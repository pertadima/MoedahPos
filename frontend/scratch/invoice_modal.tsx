interface InvoiceModalProps {
  po: PurchaseOrder;
  store: Store | null;
  onClose: () => void;
}
function InvoiceModal({ po, store, onClose }: InvoiceModalProps) {
  return (
    <Portal>
      <div className="modal-overlay no-print" style={{ zIndex: 5000 }} onClick={onClose} />
      <div
        style={{
          position: 'fixed',
          inset: 0,
          zIndex: 5001,
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
  </div>
  <style>{`@media print{body>*:not(#portal-root){display:none!important}.no-print{display:none!important}#po-invoice{position:fixed!important;inset:0!important;max-height:none!important;border-radius:0!important;box-shadow:none!important;overflow:visible!important}}`}</style>
  </Portal>
  );
}
