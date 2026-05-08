'use client';

import { useEffect, useRef, useState } from 'react';
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
  if (t.status === 'paid') return { label: 'Lunas' };
  if (t.status === 'overdue' || t.is_overdue) return { label: 'Jatuh Tempo' };
  if (t.status === 'partial') return { label: 'Sebagian' };
  return { label: 'Belum Bayar' };
}

const DOC_TITLES: Record<string, string> = {
  invoice: 'INVOICE PEMBELIAN',
  receipt: 'BUKTI PEMBAYARAN',
  termin_agreement: 'PERJANJIAN TERMIN PEMBAYARAN',
};

function DocumentContent({ data }: { data: PODocumentData }) {
  const { po, debt_summary, termins, supplier_name } = data;
  const docType = data.doc_type;
  const isReceipt = docType === 'receipt';
  const isAgreement = docType === 'termin_agreement';
  const docTitle = DOC_TITLES[docType] ?? 'DOKUMEN';
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
    <div
      style={{
        width: 595,
        padding: 40,
        background: '#fff',
        fontFamily: "'Segoe UI', Arial, sans-serif",
        color: '#000',
        fontSize: 12,
      }}
    >
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 24 }}>
        <div>
          <div style={{ fontSize: 18, fontWeight: 800, marginBottom: 4 }}>{docTitle}</div>
          <div style={{ fontSize: 11 }}>Digenerate: {formatDate(data.generated_at)}</div>
        </div>
        <div style={{ textAlign: 'right' }}>
          <div style={{ fontWeight: 700, fontSize: 13 }}>No. PO: {po.po_number}</div>
          <div style={{ fontSize: 11, marginTop: 2 }}>Tanggal: {formatDate(po.created_at)}</div>
        </div>
      </div>

      <div style={{ borderTop: '2px solid #000', marginBottom: 20 }} />

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16, marginBottom: 20 }}>
        <div>
          <div style={{ fontSize: 9, fontWeight: 700, marginBottom: 2 }}>SUPPLIER</div>
          <div style={{ fontWeight: 700 }}>{supplier_name || '—'}</div>
        </div>
        <div>
          <div style={{ fontSize: 9, fontWeight: 700, marginBottom: 2 }}>TOKO</div>
          <div style={{ fontWeight: 700 }}>{data.store_name || '—'}</div>
        </div>
      </div>

      <div
        style={{
          background: isReceipt ? '#f9fafb' : '#fff',
          border: '1px solid #e5e7eb',
          borderRadius: 4,
          padding: '10px 14px',
          marginBottom: 20,
        }}
      >
        {[
          { label: 'Total Tagihan', value: formatIDR(po.total_amount) },
          { label: 'Sudah Dibayar', value: formatIDR(debt_summary.total_paid) },
          { label: 'Sisa Hutang', value: formatIDR(debt_summary.remaining_debt) },
        ].map(item => (
          <div key={item.label} style={{ fontSize: 11, marginBottom: 2 }}>
            <span>{item.label}: </span>
            <strong>{item.value}</strong>
          </div>
        ))}
      </div>

      <div style={{ marginBottom: 20 }}>
        <div style={{ fontWeight: 700, fontSize: 12, marginBottom: 6 }}>Rincian Item</div>
        <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: 11 }}>
          <thead>
            <tr>
              <th
                style={{
                  textAlign: 'left',
                  borderBottom: '1px solid #000',
                  padding: '4px 6px',
                  fontWeight: 700,
                  fontSize: 9,
                }}
              >
                Item
              </th>
              <th
                style={{
                  textAlign: 'right',
                  borderBottom: '1px solid #000',
                  padding: '4px 6px',
                  fontWeight: 700,
                  fontSize: 9,
                }}
              >
                SKU
              </th>
              <th
                style={{
                  textAlign: 'right',
                  borderBottom: '1px solid #000',
                  padding: '4px 6px',
                  fontWeight: 700,
                  fontSize: 9,
                }}
              >
                Qty
              </th>
              <th
                style={{
                  textAlign: 'right',
                  borderBottom: '1px solid #000',
                  padding: '4px 6px',
                  fontWeight: 700,
                  fontSize: 9,
                }}
              >
                Harga
              </th>
              <th
                style={{
                  textAlign: 'right',
                  borderBottom: '1px solid #000',
                  padding: '4px 6px',
                  fontWeight: 700,
                  fontSize: 9,
                }}
              >
                Subtotal
              </th>
            </tr>
          </thead>
          <tbody>
            {(po.items ?? []).map(item => (
              <tr key={item.id}>
                <td style={{ borderBottom: '1px solid #e5e7eb', padding: '4px 6px' }}>
                  {item.product_name}
                </td>
                <td
                  style={{
                    textAlign: 'right',
                    borderBottom: '1px solid #e5e7eb',
                    padding: '4px 6px',
                    fontSize: 10,
                    color: '#666',
                  }}
                >
                  {item.product_sku || '—'}
                </td>
                <td
                  style={{
                    textAlign: 'right',
                    borderBottom: '1px solid #e5e7eb',
                    padding: '4px 6px',
                  }}
                >
                  {item.quantity} {item.unit}
                </td>
                <td
                  style={{
                    textAlign: 'right',
                    borderBottom: '1px solid #e5e7eb',
                    padding: '4px 6px',
                  }}
                >
                  {formatIDR(item.unit_cost)}
                </td>
                <td
                  style={{
                    textAlign: 'right',
                    borderBottom: '1px solid #e5e7eb',
                    padding: '4px 6px',
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
                  padding: '6px 6px 4px',
                  borderTop: '1.5px solid #000',
                }}
              >
                Total
              </td>
              <td
                style={{
                  textAlign: 'right',
                  fontWeight: 700,
                  padding: '6px 6px 4px',
                  borderTop: '1.5px solid #000',
                }}
              >
                {formatIDR(po.total_amount)}
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      {isReceipt && (
        <div style={{ marginBottom: 20 }}>
          <div style={{ fontWeight: 700, fontSize: 12, marginBottom: 6 }}>Riwayat Pembayaran</div>
          <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: 11 }}>
            <thead>
              <tr>
                <th
                  style={{
                    textAlign: 'left',
                    borderBottom: '1px solid #000',
                    padding: '4px 6px',
                    fontWeight: 700,
                    fontSize: 9,
                  }}
                >
                  Tanggal
                </th>
                <th
                  style={{
                    textAlign: 'left',
                    borderBottom: '1px solid #000',
                    padding: '4px 6px',
                    fontWeight: 700,
                    fontSize: 9,
                  }}
                >
                  Oleh
                </th>
                <th
                  style={{
                    textAlign: 'right',
                    borderBottom: '1px solid #000',
                    padding: '4px 6px',
                    fontWeight: 700,
                    fontSize: 9,
                  }}
                >
                  Jumlah
                </th>
              </tr>
            </thead>
            <tbody>
              {termins.map(t =>
                t.payments.map(r => (
                  <tr key={r.id}>
                    <td style={{ borderBottom: '1px solid #e5e7eb', padding: '4px 6px' }}>
                      {formatDate(r.paid_at ?? r.payment_date ?? '')}
                    </td>
                    <td style={{ borderBottom: '1px solid #e5e7eb', padding: '4px 6px' }}>
                      {r.paid_by_name || r.recorded_by_name || '—'}
                    </td>
                    <td
                      style={{
                        textAlign: 'right',
                        borderBottom: '1px solid #e5e7eb',
                        padding: '4px 6px',
                      }}
                    >
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
        <div style={{ marginBottom: 20 }}>
          <div style={{ fontWeight: 700, fontSize: 12, marginBottom: 6 }}>Jadwal Termin</div>
          <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: 11 }}>
            <thead>
              <tr>
                <th
                  style={{
                    textAlign: 'left',
                    borderBottom: '1px solid #000',
                    padding: '4px 6px',
                    fontWeight: 700,
                    fontSize: 9,
                  }}
                >
                  Termin
                </th>
                <th
                  style={{
                    textAlign: 'left',
                    borderBottom: '1px solid #000',
                    padding: '4px 6px',
                    fontWeight: 700,
                    fontSize: 9,
                  }}
                >
                  Jatuh Tempo
                </th>
                <th
                  style={{
                    textAlign: 'right',
                    borderBottom: '1px solid #000',
                    padding: '4px 6px',
                    fontWeight: 700,
                    fontSize: 9,
                  }}
                >
                  Jumlah
                </th>
                <th
                  style={{
                    textAlign: 'center',
                    borderBottom: '1px solid #000',
                    padding: '4px 6px',
                    fontWeight: 700,
                    fontSize: 9,
                  }}
                >
                  Status
                </th>
              </tr>
            </thead>
            <tbody>
              {termins.map(t => (
                <tr key={t.id}>
                  <td style={{ borderBottom: '1px solid #e5e7eb', padding: '4px 6px' }}>
                    Termin {t.termin_number}
                  </td>
                  <td style={{ borderBottom: '1px solid #e5e7eb', padding: '4px 6px' }}>
                    {formatDate(t.due_date)}
                  </td>
                  <td
                    style={{
                      textAlign: 'right',
                      borderBottom: '1px solid #e5e7eb',
                      padding: '4px 6px',
                    }}
                  >
                    {formatIDR(t.amount)}
                  </td>
                  <td
                    style={{
                      textAlign: 'center',
                      borderBottom: '1px solid #e5e7eb',
                      padding: '4px 6px',
                    }}
                  >
                    {statusLabel(t).label}
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
            borderRadius: 4,
            padding: '12px 16px',
            marginBottom: 20,
          }}
        >
          <div style={{ fontWeight: 700, marginBottom: 6 }}>Ketentuan Perjanjian</div>
          <div style={{ fontSize: 10, marginBottom: 2 }}>
            1. Pihak pembeli wajib melunasi setiap termin sesuai tanggal jatuh tempo.
          </div>
          <div style={{ fontSize: 10, marginBottom: 2 }}>
            2. Keterlambatan pembayaran dapat dikenakan kebijakan tambahan sesuai kesepakatan.
          </div>
          <div style={{ fontSize: 10 }}>
            3. Dokumen ini berlaku sebagai bukti kesepakatan pembayaran bertahap atas PO ini.
          </div>
        </div>
      )}

      <div
        style={{
          display: 'grid',
          gridTemplateColumns: '1fr 1fr',
          gap: 32,
          marginTop: 32,
          paddingTop: 16,
          borderTop: '2px solid #000',
        }}
      >
        <div style={{ textAlign: 'center' }}>
          <div style={{ fontSize: 10, fontWeight: 700, marginBottom: 48 }}>
            Pihak Supplier / Pemberi Barang
          </div>
          <div style={{ borderTop: '1px solid #000', paddingTop: 4, fontSize: 10 }}>
            Tanda Tangan &amp; Nama
          </div>
        </div>

        <div style={{ textAlign: 'center' }}>
          <div style={{ fontSize: 10, fontWeight: 700, marginBottom: 4 }}>
            Pihak Pembeli / Penerima
          </div>
          <div style={{ position: 'relative', width: 120, height: 120, margin: '0 auto' }}>
            <svg width="120" height="120" style={{ position: 'absolute', top: 0, left: 0 }}>
              <circle cx="60" cy="60" r="56" fill="none" stroke="#ef4444" strokeWidth="2" />
              <circle
                cx="60"
                cy="60"
                r="50"
                fill="none"
                stroke="#ef4444"
                strokeWidth="0.5"
                strokeDasharray="2,2"
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
                gap: 1,
              }}
            >
              <div style={{ fontSize: 11, fontWeight: 800, color: '#ef4444', opacity: 0.8 }}>
                {data.store_name || 'MoedahPOS'}
              </div>
              <div style={{ width: 60, height: 1, background: '#ef4444', opacity: 0.8 }} />
              <div style={{ fontSize: 9, color: '#ef4444', opacity: 0.8 }}>{stampDate}</div>
              <div style={{ fontSize: 9, color: '#ef4444', opacity: 0.8 }}>{stampTime}</div>
            </div>
          </div>
        </div>
      </div>

      <div style={{ marginTop: 24, textAlign: 'center', fontSize: 9, color: '#666' }}>
        Halaman 1 dari 1 · Dokumen ini dibuat secara otomatis oleh sistem MoedahPOS · {po.po_number}
      </div>
    </div>
  );
}

export default function PODocumentPage() {
  const params = useParams();
  const searchParams = useSearchParams();
  const { selectedStore } = useAuth();
  const docRef = useRef<HTMLDivElement>(null);
  const [pdfUrl, setPdfUrl] = useState<string | null>(null);
  const [generating, setGenerating] = useState(false);

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

  const generatePDF = async () => {
    if (!docRef.current || !data) return;
    setGenerating(true);
    try {
      const [{ default: html2canvas }, { jsPDF }] = await Promise.all([
        import('html2canvas'),
        import('jspdf'),
      ]);

      const canvas = await html2canvas(docRef.current, {
        scale: 2,
        useCORS: true,
        backgroundColor: '#ffffff',
        logging: false,
      } as unknown as Parameters<typeof html2canvas>[1]);

      const pdf = new jsPDF({ orientation: 'portrait', unit: 'pt', format: 'a4' });

      const pdfWidth = pdf.internal.pageSize.getWidth();
      const pdfHeight = pdf.internal.pageSize.getHeight();
      const canvasWidth = canvas.width;
      const canvasHeight = canvas.height;
      const ratio = pdfWidth / canvasWidth;
      const sliceHeight = pdfHeight;
      const totalPages = Math.ceil((canvasHeight * ratio) / sliceHeight);

      for (let i = 0; i < totalPages; i++) {
        if (i > 0) pdf.addPage();
        const srcY = (i * sliceHeight) / ratio;
        const srcH = Math.min(sliceHeight / ratio, canvasHeight - srcY);
        const pageCanvas = document.createElement('canvas');
        pageCanvas.width = canvasWidth;
        pageCanvas.height = srcH;
        const ctx = pageCanvas.getContext('2d');
        if (!ctx) continue;
        ctx.drawImage(canvas, 0, srcY, canvasWidth, srcH, 0, 0, canvasWidth, srcH);
        const pageImg = pageCanvas.toDataURL('image/jpeg', 0.95);
        pdf.addImage(pageImg, 'JPEG', 0, 0, pdfWidth, srcH * ratio);
      }

      const blob = pdf.output('blob');
      const url = URL.createObjectURL(blob);
      setPdfUrl(url);
    } catch (err) {
      console.error('PDF generation failed:', err);
    } finally {
      setGenerating(false);
    }
  };

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

  return (
    <div
      style={{
        minHeight: '100vh',
        background: '#f3f4f6',
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        padding: '24px 16px',
        gap: 16,
      }}
    >
      <div
        style={{
          display: 'flex',
          gap: 10,
          alignItems: 'center',
          width: 675,
        }}
      >
        <button
          onClick={generatePDF}
          disabled={generating}
          style={{
            padding: '9px 20px',
            background: generating ? '#9ca3af' : '#2563eb',
            color: '#fff',
            border: 'none',
            borderRadius: 8,
            cursor: generating ? 'not-allowed' : 'pointer',
            fontWeight: 600,
            fontSize: 14,
            display: 'flex',
            alignItems: 'center',
            gap: 6,
          }}
        >
          {generating ? 'Membuat PDF...' : '⬇ Download PDF'}
        </button>
        <button
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

      <div
        style={{
          background: '#fff',
          boxShadow: '0 4px 24px rgba(0,0,0,0.08)',
          borderRadius: 4,
          overflow: 'hidden',
        }}
      >
        <div ref={docRef}>
          <DocumentContent data={data} />
        </div>
      </div>

      {pdfUrl && (
        <iframe
          src={pdfUrl}
          style={{
            width: 675,
            height: 900,
            border: 'none',
            borderRadius: 4,
            boxShadow: '0 4px 24px rgba(0,0,0,0.12)',
          }}
          title="PDF Preview"
        />
      )}
    </div>
  );
}
