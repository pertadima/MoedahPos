'use client';

import { useEffect, useState } from 'react';
import { useParams, useSearchParams } from 'next/navigation';
import { getPODocument } from '@/lib/api/termins';
import type { PODocumentData, Termin } from '@/types';
import { useAuth } from '@/lib/auth/AuthContext';

function formatIDR(n: number) {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(n);
}

function formatDate(s: string) {
  return new Date(s).toLocaleDateString('id-ID', {
    day: '2-digit',
    month: 'long',
    year: 'numeric',
  });
}

function statusLabel(t: Termin) {
  if (t.status === 'paid') return { label: 'Lunas', color: '#000' };
  if (t.status === 'overdue' || t.is_overdue) return { label: 'Jatuh Tempo', color: '#000' };
  if (t.status === 'partial') return { label: 'Sebagian', color: '#000' };
  return { label: 'Belum Bayar', color: '#000' };
}

const DOC_TITLES: Record<string, string> = {
  invoice: 'INVOICE PEMBELIAN',
  receipt: 'BUKTI PEMBAYARAN',
  termin_agreement: 'PERJANJIAN TERMIN PEMBAYARAN',
};

export default function PODocumentPage() {
  const params = useParams();
  const searchParams = useSearchParams();
  const { selectedStore } = useAuth();

  const poId = params?.id as string;
  const docType = (searchParams?.get('type') ?? 'invoice') as
    | 'invoice'
    | 'receipt'
    | 'termin_agreement';

  const [data, setData] = useState<PODocumentData | null>(null);
  const [error, setError] = useState('');

  useEffect(() => {
    if (!selectedStore || !poId) return;
    getPODocument(selectedStore.store_id, poId, docType)
      .then(setData)
      .catch(() => setError('Gagal memuat data dokumen.'));
  }, [selectedStore, poId, docType]);

  if (error) {
    return (
      <div style={{ padding: 32, fontFamily: 'sans-serif', color: '#dc2626' }}>
        <strong>Error:</strong> {error}
      </div>
    );
  }

  if (!data) {
    return (
      <div style={{ padding: 32, fontFamily: 'sans-serif', color: '#6b7280' }}>Memuat dokumen…</div>
    );
  }

  const { po, debt_summary, termins, supplier_name } = data;
  const docTitle = DOC_TITLES[docType] ?? 'DOKUMEN';
  const isReceipt = docType === 'receipt';
  const isAgreement = docType === 'termin_agreement';

  const stampDate = new Date().toLocaleDateString('id-ID', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
  });
  const stampTime = new Date().toLocaleTimeString('id-ID', {
    hour: '2-digit',
    minute: '2-digit',
  });

  return (
    <>
      <style>{`
        @media print {
          .no-print { display: none !important; }
          body { background: white; }
        }
        body { margin: 0; font-family: 'Segoe UI', Arial, sans-serif; background: #f3f4f6; color: #000; }
        .page { max-width: 800px; margin: 0 auto; background: white; padding: 48px; box-shadow: 0 4px 24px rgba(0,0,0,0.08); }
        table { width: 100%; border-collapse: collapse; }
        th, td { padding: 8px 12px; text-align: left; border-bottom: 1px solid #e5e7eb; font-size: 13px; color: #000; }
        th { background: #f9fafb; fontWeight: 600; color: #000; border-top: 2px solid #000; border-bottom: 2px solid #000; }
        .badge { display: inline-block; padding: 2px 10px; border-radius: 12px; font-size: 11px; font-weight: 600; border: 1px solid #ccc; background: #f9fafb; }
      `}</style>

      <div
        className="no-print"
        style={{
          position: 'fixed',
          top: 16,
          right: 16,
          display: 'flex',
          gap: 8,
          zIndex: 10,
        }}
      >
        <button
          id="btn-print-doc"
          onClick={() => window.print()}
          style={{
            padding: '9px 20px',
            background: '#2563eb',
            color: '#fff',
            border: 'none',
            borderRadius: 8,
            cursor: 'pointer',
            fontWeight: 600,
            fontSize: 14,
          }}
        >
          🖨 Cetak
        </button>
        <button
          id="btn-close-doc"
          onClick={() => window.close()}
          style={{
            padding: '9px 20px',
            background: '#6b7280',
            color: '#fff',
            border: 'none',
            borderRadius: 8,
            cursor: 'pointer',
            fontWeight: 600,
            fontSize: 14,
          }}
        >
          Tutup
        </button>
      </div>

      <div className="page">
        <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 32 }}>
          <div>
            <div
              style={{
                fontSize: 22,
                fontWeight: 800,
                color: '#000',
                letterSpacing: '-0.5px',
                marginBottom: 4,
              }}
            >
              {docTitle}
            </div>
            <div style={{ color: '#000', fontSize: 13 }}>
              Digenerate: {formatDate(data.generated_at)}
            </div>
          </div>
          <div style={{ textAlign: 'right' }}>
            <div style={{ fontWeight: 700, fontSize: 15, color: '#000' }}>
              No. PO: {po.po_number}
            </div>
            <div style={{ color: '#000', fontSize: 13, marginTop: 2 }}>
              Tanggal: {formatDate(po.created_at)}
            </div>
          </div>
        </div>

        <div style={{ borderTop: '3px solid #000', marginBottom: 24 }} />

        <div
          style={{
            display: 'grid',
            gridTemplateColumns: '1fr 1fr',
            gap: 24,
            marginBottom: 28,
            color: '#000',
          }}
        >
          <div>
            <div style={{ fontSize: 11, fontWeight: 700, marginBottom: 4 }}>SUPPLIER</div>
            <div style={{ fontWeight: 700, fontSize: 15 }}>
              {supplier_name || '— (tidak ada supplier)'}
            </div>
          </div>
          <div>
            <div style={{ fontSize: 11, fontWeight: 700, marginBottom: 4 }}>TOKO</div>
            <div style={{ fontWeight: 700, fontSize: 15 }}>{data.store_name || '—'}</div>
          </div>
        </div>

        <div
          style={{
            background: isReceipt ? '#f9fafb' : '#fff',
            border: '1px solid #e5e7eb',
            borderRadius: 8,
            padding: '12px 16px',
            marginBottom: 28,
            color: '#000',
          }}
        >
          {[
            { label: 'Total Tagihan', value: formatIDR(po.total_amount) },
            { label: 'Sudah Dibayar', value: formatIDR(debt_summary.total_paid) },
            { label: 'Sisa Hutang', value: formatIDR(debt_summary.remaining_debt) },
          ].map(label => (
            <div key={label.label} style={{ fontSize: 13, marginBottom: 4 }}>
              <span>{label.label}: </span>
              <strong>{label.value}</strong>
            </div>
          ))}
        </div>

        <div style={{ marginBottom: 32 }}>
          <div style={{ fontWeight: 700, fontSize: 14, color: '#000', marginBottom: 10 }}>
            Rincian Item
          </div>
          <table>
            <thead>
              <tr>
                <th style={{ width: '40%' }}>Item</th>
                <th style={{ textAlign: 'right' }}>SKU</th>
                <th style={{ textAlign: 'right' }}>Qty</th>
                <th style={{ textAlign: 'right' }}>Harga</th>
                <th style={{ textAlign: 'right' }}>Subtotal</th>
              </tr>
            </thead>
            <tbody>
              {po.items?.map(item => (
                <tr key={item.id}>
                  <td style={{ borderBottom: '1px solid #e5e7eb', padding: '6px 8px' }}>
                    {item.product_name}
                  </td>
                  <td
                    style={{
                      textAlign: 'right',
                      fontSize: 12,
                      color: '#666',
                      borderBottom: '1px solid #e5e7eb',
                      padding: '6px 8px',
                    }}
                  >
                    {item.product_sku || '—'}
                  </td>
                  <td
                    style={{
                      textAlign: 'right',
                      borderBottom: '1px solid #e5e7eb',
                      padding: '6px 8px',
                    }}
                  >
                    {item.quantity} {item.unit}
                  </td>
                  <td
                    style={{
                      textAlign: 'right',
                      borderBottom: '1px solid #e5e7eb',
                      padding: '6px 8px',
                    }}
                  >
                    {formatIDR(item.unit_cost)}
                  </td>
                  <td
                    style={{
                      textAlign: 'right',
                      borderBottom: '1px solid #e5e7eb',
                      padding: '6px 8px',
                    }}
                  >
                    {formatIDR(item.subtotal)}
                  </td>
                </tr>
              ))}
              <tr>
                <td
                  colSpan={4}
                  style={{
                    textAlign: 'right',
                    fontWeight: 700,
                    padding: '8px 8px',
                    borderTop: '2px solid #000',
                  }}
                >
                  Total
                </td>
                <td
                  style={{
                    textAlign: 'right',
                    fontWeight: 700,
                    padding: '8px 8px',
                    borderTop: '2px solid #000',
                  }}
                >
                  {formatIDR(po.total_amount)}
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        {isReceipt && (
          <div style={{ marginBottom: 32 }}>
            <div style={{ fontWeight: 700, fontSize: 14, color: '#000', marginBottom: 10 }}>
              Riwayat Pembayaran
            </div>
            <table>
              <thead>
                <tr>
                  <th>Tanggal</th>
                  <th>Oleh</th>
                  <th style={{ textAlign: 'right' }}>Jumlah</th>
                </tr>
              </thead>
              <tbody>
                {termins.map(t =>
                  t.payments.map(r => (
                    <tr key={r.id}>
                      <td>{formatDate(r.paid_at ?? r.payment_date ?? '')}</td>
                      <td>{r.paid_by_name || r.recorded_by_name || '—'}</td>
                      <td style={{ textAlign: 'right' }}>
                        {formatIDR(r.amount ?? r.amount_paid ?? 0)}
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        )}

        {isAgreement && (
          <div style={{ marginBottom: 32 }}>
            <div style={{ fontWeight: 700, fontSize: 14, color: '#000', marginBottom: 10 }}>
              Jadwal Termin
            </div>
            <table>
              <thead>
                <tr>
                  <th>Termin</th>
                  <th>Jatuh Tempo</th>
                  <th style={{ textAlign: 'right' }}>Jumlah</th>
                  <th style={{ textAlign: 'center' }}>Status</th>
                </tr>
              </thead>
              <tbody>
                {termins.map(t => (
                  <tr key={t.id}>
                    <td>Termin {t.termin_number}</td>
                    <td>{formatDate(t.due_date)}</td>
                    <td style={{ textAlign: 'right' }}>{formatIDR(t.amount)}</td>
                    <td style={{ textAlign: 'center' }}>
                      <span className="badge">{statusLabel(t).label}</span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

        {isAgreement && (
          <div
            style={{
              border: '1px solid #e5e7eb',
              borderRadius: 8,
              padding: '16px 20px',
              marginBottom: 32,
              color: '#000',
            }}
          >
            <div style={{ fontWeight: 700, marginBottom: 8 }}>Ketentuan Perjanjian</div>
            <div>1. Pihak pembeli wajib melunasi setiap termin sesuai tanggal jatuh tempo.</div>
            <div>
              2. Keterlambatan pembayaran dapat dikenakan kebijakan tambahan sesuai kesepakatan.
            </div>
            <div>
              3. Dokumen ini berlaku sebagai bukti kesepakatan pembayaran bertahap atas PO ini.
            </div>
          </div>
        )}

        <div
          style={{
            display: 'grid',
            gridTemplateColumns: '1fr 1fr',
            gap: 32,
            marginTop: 40,
            paddingTop: 24,
            borderTop: '2px solid #000',
            color: '#000',
          }}
        >
          <div style={{ textAlign: 'center' }}>
            <div style={{ fontSize: 12, fontWeight: 700, marginBottom: 64 }}>
              Pihak Supplier / Pemberi Barang
            </div>
            <div style={{ borderTop: '1px solid #000', paddingTop: 6, fontSize: 12 }}>
              Tanda Tangan &amp; Nama
            </div>
          </div>

          <div style={{ textAlign: 'center' }}>
            <div style={{ fontSize: 12, fontWeight: 700, marginBottom: 4 }}>
              Pihak Pembeli / Penerima
            </div>
            <div style={{ position: 'relative', width: 148, height: 148, margin: '0 auto' }}>
              <svg width="148" height="148" style={{ position: 'absolute', top: 0, left: 0 }}>
                <circle cx="74" cy="74" r="70" fill="none" stroke="#ef4444" strokeWidth="2.5" />
                <circle
                  cx="74"
                  cy="74"
                  r="64"
                  fill="none"
                  stroke="#ef4444"
                  strokeWidth="0.5"
                  strokeDasharray="3,3"
                />
              </svg>
              <div
                style={{
                  position: 'absolute',
                  top: 0,
                  left: 0,
                  width: '100%',
                  height: '100%',
                  display: 'flex',
                  flexDirection: 'column',
                  alignItems: 'center',
                  justifyContent: 'center',
                  gap: 2,
                }}
              >
                <div
                  style={{
                    fontSize: '0.85rem',
                    fontWeight: 800,
                    color: '#ef4444',
                    textAlign: 'center',
                    lineHeight: 1.2,
                    opacity: 0.75,
                  }}
                >
                  {data.store_name || 'MoedahPOS'}
                </div>
                <div
                  style={{
                    width: 80,
                    height: 1,
                    background: '#ef4444',
                    opacity: 0.75,
                    margin: '4px 0',
                  }}
                />
                <div
                  style={{
                    fontSize: '0.58rem',
                    color: '#ef4444',
                    textAlign: 'center',
                    opacity: 0.75,
                  }}
                >
                  {stampDate}
                </div>
                <div
                  style={{
                    fontSize: '0.58rem',
                    color: '#ef4444',
                    textAlign: 'center',
                    opacity: 0.75,
                  }}
                >
                  {stampTime}
                </div>
              </div>
            </div>
          </div>
        </div>

        <div
          style={{
            marginTop: 32,
            textAlign: 'center',
            fontSize: 11,
            color: '#000',
          }}
        >
          Halaman 1 dari 1 · Dokumen ini dibuat secara otomatis oleh sistem MoedahPOS ·{' '}
          {po.po_number}
        </div>
      </div>
    </>
  );
}
