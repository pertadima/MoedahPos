'use client';

import { useEffect, useState } from 'react';
import { useParams, useSearchParams } from 'next/navigation';
import { getPODocument, type PODocumentData, type Termin } from '@/lib/api/termins';
import { useAuth } from '@/lib/auth/AuthContext';

// ─── Helpers ──────────────────────────────────────────────────────────────────

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

// ─── Page ─────────────────────────────────────────────────────────────────────

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

  return (
    <>
      {/* Print-specific global styles */}
      <style>{`
        @media print {
          .no-print { display: none !important; }
          body { background: white; }
        }
        body { margin: 0; font-family: 'Segoe UI', Arial, sans-serif; background: #f3f4f6; color: #000; }
        .page { max-width: 800px; margin: 0 auto; background: white; padding: 48px; box-shadow: 0 4px 24px rgba(0,0,0,0.08); }
        table { width: 100%; border-collapse: collapse; }
        th, td { padding: 8px 12px; text-align: left; border-bottom: 1px solid #e5e7eb; font-size: 13px; color: #000; }
        th { background: #f9fafb; font-weight: 600; color: #000; border-top: 2px solid #000; border-bottom: 2px solid #000; }
        .badge { display: inline-block; padding: 2px 10px; border-radius: 12px; font-size: 11px; font-weight: 600; border: 1px solid #ccc; background: #f9fafb; }
      `}</style>

      {/* Print button — hidden when printing */}
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
        {/* Header */}
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
            <div style={{ fontWeight: 700, fontSize: 15, color: '#000' }}>No. PO: {po.po_number}</div>
            <div style={{ color: '#000', fontSize: 13, marginTop: 2 }}>
              Tanggal: {formatDate(po.created_at)}
            </div>
          </div>
        </div>

        {/* Divider */}
        <div style={{ borderTop: '3px solid #000', marginBottom: 24 }} />

        {/* PO + Supplier Info */}
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
            <div style={{ fontSize: 11, fontWeight: 700, marginBottom: 4 }}>
              SUPPLIER
            </div>
            <div style={{ fontWeight: 700, fontSize: 15 }}>
              {supplier_name || '— (tidak ada supplier)'}
            </div>
          </div>
          <div>
            <div style={{ fontSize: 11, fontWeight: 700, marginBottom: 4 }}>
              TOKO
            </div>
            <div style={{ fontWeight: 700, fontSize: 15 }}>{data.store_name || '—'}</div>
          </div>
        </div>

        {/* Debt Summary */}
        <div
          style={{
            background: '#fff',
            border: '2px solid #000',
            borderRadius: 6,
            padding: '16px 20px',
            marginBottom: 28,
            display: 'grid',
            gridTemplateColumns: 'repeat(4, 1fr)',
            gap: 12,
            color: '#000'
          }}
        >
          {[
            ['Total PO', formatIDR(po.total_amount)],
            ['Total Termin', formatIDR(debt_summary.total_termin)],
            ['Total Dibayar', formatIDR(debt_summary.total_paid)],
            ['Sisa Hutang', formatIDR(debt_summary.remaining_debt)],
          ].map(([label, value]) => (
            <div key={label}>
              <div style={{ fontSize: 11, fontWeight: 700, marginBottom: 2 }}>{label}</div>
              <div style={{ fontWeight: 700, fontSize: 15 }}>{value}</div>
            </div>
          ))}
        </div>

        {/* Termin Table */}
        <div style={{ marginBottom: 32 }}>
          <div style={{ fontWeight: 700, fontSize: 14, color: '#000', marginBottom: 10 }}>
            Jadwal Termin Pembayaran
          </div>
          <table>
            <thead>
              <tr>
                <th>Termin</th>
                <th>Jatuh Tempo</th>
                <th style={{ textAlign: 'right' }}>Jumlah</th>
                <th style={{ textAlign: 'right' }}>Dibayar</th>
                <th style={{ textAlign: 'right' }}>Sisa</th>
                <th>Status</th>
                <th>Catatan</th>
              </tr>
            </thead>
            <tbody>
              {termins.map(t => {
                const { label, color } = statusLabel(t);
                return (
                  <tr key={t.id}>
                    <td style={{ fontWeight: 600 }}>Termin {t.termin_number}</td>
                    <td>{formatDate(t.due_date)}</td>
                    <td style={{ textAlign: 'right' }}>{formatIDR(t.amount)}</td>
                    <td style={{ textAlign: 'right' }}>
                      {formatIDR(t.amount_paid)}
                    </td>
                    <td style={{ textAlign: 'right', fontWeight: 600 }}>
                      {formatIDR(t.amount_due)}
                    </td>
                    <td>
                      <span className="badge">
                        {label}
                      </span>
                    </td>
                    <td style={{ fontSize: 12 }}>{t.notes || '—'}</td>
                  </tr>
                );
              })}
              {/* Total row */}
              <tr style={{ background: '#f9fafb', fontWeight: 700 }}>
                <td colSpan={2}>Total</td>
                <td style={{ textAlign: 'right' }}>{formatIDR(debt_summary.total_termin)}</td>
                <td style={{ textAlign: 'right' }}>
                  {formatIDR(debt_summary.total_paid)}
                </td>
                <td style={{ textAlign: 'right' }}>
                  {formatIDR(debt_summary.remaining_debt)}
                </td>
                <td colSpan={2} />
              </tr>
            </tbody>
          </table>
        </div>

        {/* Payment History (receipt & invoice) */}
        {docType !== 'termin_agreement' && termins.some(t => t.payments.length > 0) && (
          <div style={{ marginBottom: 32 }}>
            <div style={{ fontWeight: 700, fontSize: 14, color: '#000', marginBottom: 10 }}>
              Riwayat Pembayaran
            </div>
            <table>
              <thead>
                <tr>
                  <th>Termin</th>
                  <th>Tanggal Bayar</th>
                  <th>Metode</th>
                  <th style={{ textAlign: 'right' }}>Jumlah</th>
                  <th>Catatan</th>
                  <th>Dicatat Oleh</th>
                </tr>
              </thead>
              <tbody>
                {termins.flatMap(t =>
                  t.payments.map(p => (
                    <tr key={p.id}>
                      <td>Termin {t.termin_number}</td>
                      <td>{formatDate(p.payment_date)}</td>
                      <td style={{ textTransform: 'capitalize' }}>{p.payment_method}</td>
                      <td style={{ textAlign: 'right', fontWeight: 600 }}>
                        {formatIDR(p.amount_paid)}
                      </td>
                      <td style={{ fontSize: 12 }}>{p.notes || '—'}</td>
                      <td style={{ fontSize: 12 }}>{p.recorded_by_name}</td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        )}

        {/* Signature Block */}
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
          {['Pihak Supplier', 'Pihak Pembeli'].map(label => (
            <div key={label} style={{ textAlign: 'center' }}>
              <div style={{ fontSize: 12, fontWeight: 700, marginBottom: 64 }}>{label}</div>
              <div
                style={{
                  borderTop: '1px solid #000',
                  paddingTop: 6,
                  fontSize: 12,
                }}
              >
                Tanda Tangan &amp; Nama
              </div>
            </div>
          ))}
        </div>

        {/* Footer */}
        <div
          style={{
            marginTop: 32,
            textAlign: 'center',
            fontSize: 11,
            color: '#000',
          }}
        >
          Dokumen ini dibuat secara otomatis oleh sistem MoedahPOS · {po.po_number}
        </div>
      </div>
    </>
  );
}
