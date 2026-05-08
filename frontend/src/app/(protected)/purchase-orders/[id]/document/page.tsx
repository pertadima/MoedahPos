'use client';

import { useParams, useSearchParams } from 'next/navigation';
import {
  PDFDownloadLink,
  PDFViewer,
  Document,
  Page,
  View,
  Text,
  StyleSheet,
  Font,
  Svg,
  Rect,
} from '@react-pdf/renderer';
import { getPODocument } from '@/lib/api/termins';
import type { PODocumentData, Termin } from '@/types';
import { useAuth } from '@/lib/auth/AuthContext';
import { useEffect, useState } from 'react';

Font.register({
  family: 'Helvetica',
  fonts: [
    { src: 'https://fonts.gstatic.com/s/roboto/v30/KFOmCnqEu92Fr1Me5Q.ttf' },
    {
      src: 'https://fonts.gstatic.com/s/roboto/v30/KFOlCnqEu92Fr1MmEU9fBBc9.ttf',
      fontWeight: 'bold',
    },
  ],
});

function fmtIDR(n: number) {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(n);
}

function fmtDate(s: string) {
  return new Date(s).toLocaleDateString('id-ID', {
    day: '2-digit',
    month: 'long',
    year: 'numeric',
  });
}

function sLabel(t: Termin) {
  if (t.status === 'paid') return 'Lunas';
  if (t.status === 'overdue' || t.is_overdue) return 'Jatuh Tempo';
  if (t.status === 'partial') return 'Sebagian';
  return 'Belum Bayar';
}

const DOC_TITLES: Record<string, string> = {
  invoice: 'INVOICE PEMBELIAN',
  receipt: 'BUKTI PEMBAYARAN',
  termin_agreement: 'PERJANJIAN TERMIN PEMBAYARAN',
};

const s = StyleSheet.create({
  page: {
    padding: 40,
    fontFamily: 'Helvetica',
    fontSize: 10,
    color: '#000',
    backgroundColor: '#fff',
  },
  headerWrap: { marginBottom: 14 },
  header: { flexDirection: 'row', justifyContent: 'space-between', marginBottom: 14 },
  docTitle: { fontSize: 16, fontWeight: 'bold' },
  genDate: { fontSize: 9, color: '#666', marginTop: 2 },
  poNumber: { fontSize: 12, fontWeight: 'bold', textAlign: 'right' },
  poDate: { fontSize: 9, marginTop: 2, textAlign: 'right' },
  divider: { borderTopWidth: 2, borderTopColor: '#000', marginBottom: 12 },
  twoCol: { flexDirection: 'row', gap: 16, marginBottom: 10 },
  col: { flex: 1 },
  label: { fontSize: 8, fontWeight: 'bold', color: '#888', marginBottom: 1, letterSpacing: 0.5 },
  value: { fontSize: 10, fontWeight: 'bold' },
  summaryBox: {
    borderWidth: 1,
    borderColor: '#e5e7eb',
    borderRadius: 4,
    padding: '8px 14px',
    marginBottom: 12,
    backgroundColor: '#f9fafb',
  },
  summaryRow: { fontSize: 9, marginBottom: 1 },
  summaryVal: { fontWeight: 'bold' },
  sectionTitle: { fontSize: 11, fontWeight: 'bold', marginBottom: 4 },
  table: { width: '100%', marginBottom: 12, borderWidth: 1, borderColor: '#000' },
  th: {
    flexDirection: 'row',
    backgroundColor: '#1e293b',
    paddingVertical: 4,
    paddingHorizontal: 5,
    borderBottomWidth: 1,
    borderBottomColor: '#000',
  },
  thItem: { flex: 1, fontSize: 8, fontWeight: 'bold', color: '#fff' },
  thRight: { width: 90, fontSize: 8, fontWeight: 'bold', color: '#fff', textAlign: 'right' },
  thCenter: { width: 80, fontSize: 8, fontWeight: 'bold', color: '#fff', textAlign: 'center' },
  tr: {
    flexDirection: 'row',
    borderBottomWidth: 1,
    borderBottomColor: '#cbd5e1',
    paddingVertical: 3,
    paddingHorizontal: 5,
  },
  td: { flex: 1, fontSize: 9 },
  tdRight: { width: 90, fontSize: 9, textAlign: 'right' },
  tdCenter: { width: 80, fontSize: 9, textAlign: 'center' },
  tdSku: { width: 70, fontSize: 8, color: '#888', textAlign: 'right' },
  trTotal: {
    flexDirection: 'row',
    borderTopWidth: 1.5,
    borderTopColor: '#000',
    borderBottomWidth: 1,
    borderBottomColor: '#000',
    paddingVertical: 3,
    paddingHorizontal: 5,
    backgroundColor: '#f1f5f9',
  },
  tdTotalLabel: { flex: 1, fontSize: 9, fontWeight: 'bold', textAlign: 'right' },
  tdTotal: { width: 90, fontSize: 9, fontWeight: 'bold', textAlign: 'right' },
  agreementBox: {
    borderWidth: 1,
    borderColor: '#e5e7eb',
    borderRadius: 4,
    padding: '10px 14px',
    marginBottom: 12,
  },
  agreeTitle: { fontSize: 9, fontWeight: 'bold', marginBottom: 4 },
  agreeItem: { fontSize: 8, marginBottom: 2 },
  signatureSection: {
    flexDirection: 'row',
    gap: 32,
    marginTop: 20,
    paddingTop: 10,
    borderTopWidth: 2,
    borderTopColor: '#000',
  },
  sigBox: { flex: 1, alignItems: 'center', justifyContent: 'space-between' },
  sigLabel: { fontSize: 8, fontWeight: 'bold', color: '#888', marginBottom: 36 },
  sigLine: {
    borderTopWidth: 1,
    borderTopColor: '#000',
    paddingTop: 2,
    fontSize: 8,
    width: 120,
    textAlign: 'center',
  },
  stamp: { alignItems: 'center', marginBottom: 36 },
  footer: {
    position: 'absolute',
    bottom: 20,
    left: 40,
    right: 40,
    borderTopWidth: 1,
    borderTopColor: '#e5e7eb',
    paddingTop: 4,
    alignItems: 'center',
  },
  footerText: { fontSize: 7, color: '#888' },
});

function PODocumentPDF({ data }: { data: PODocumentData }) {
  const { po, debt_summary, termins, supplier_name } = data;
  const docType = data.doc_type;
  const isReceipt = docType === 'receipt';
  const isAgreement = docType === 'termin_agreement';
  const isReceived = po.status === 'received';
  const docTitle = DOC_TITLES[docType] ?? 'DOKUMEN';
  const hasItems = (po.items ?? []).length > 0;
  const hasPayments = termins.some(t => t.payments.length > 0);
  const hasTermins = termins.length > 0;
  const totalAmount =
    po.items?.reduce((acc, item) => acc + item.subtotal, 0) ?? po.total_amount ?? 0;

  return (
    <Document>
      <Page size="A4" style={s.page}>
        <View style={s.headerWrap} fixed>
          <View style={s.header}>
            <View>
              <Text style={s.docTitle}>{docTitle}</Text>
              <Text style={s.genDate}>Digenerate: {fmtDate(data.generated_at)}</Text>
            </View>
            <View>
              <Text style={s.poNumber}>No. PO: {po.po_number}</Text>
              <Text style={s.poDate}>Tanggal: {fmtDate(po.created_at)}</Text>
            </View>
          </View>
          <View style={s.divider} />
        </View>

        <View>
          <View style={s.twoCol}>
            <View style={s.col}>
              <Text style={s.label}>SUPPLIER</Text>
              <Text style={s.value}>{supplier_name || '—'}</Text>
            </View>
            <View style={s.col}>
              <Text style={s.label}>TOKO</Text>
              <Text style={s.value}>{data.store_name || '—'}</Text>
            </View>
          </View>

          <View style={s.summaryBox}>
            <Text style={s.summaryRow}>
              Total Tagihan:{' '}
              <Text style={s.summaryVal}>{fmtIDR(po.total_amount ?? totalAmount)}</Text>
            </Text>
            <Text style={s.summaryRow}>
              Sudah Dibayar: <Text style={s.summaryVal}>{fmtIDR(debt_summary.total_paid)}</Text>
            </Text>
            <Text style={s.summaryRow}>
              Sisa Hutang: <Text style={s.summaryVal}>{fmtIDR(debt_summary.remaining_debt)}</Text>
            </Text>
          </View>

          {hasItems && (
            <>
              <Text style={s.sectionTitle}>Rincian Item</Text>
              <View style={s.table}>
                <View style={s.th}>
                  <Text style={s.thItem}>Item</Text>
                  <Text style={{ width: 70, fontSize: 8, fontWeight: 'bold', textAlign: 'right' }}>
                    SKU
                  </Text>
                  <Text style={s.thRight}>Qty</Text>
                  <Text style={s.thRight}>Harga</Text>
                  <Text style={s.thRight}>Subtotal</Text>
                </View>
                {(po.items ?? []).map(item => (
                  <View key={item.id} style={s.tr}>
                    <Text style={s.td}>{item.product_name}</Text>
                    <Text style={s.tdSku}>{item.product_sku || '—'}</Text>
                    <Text style={s.tdRight}>
                      {item.quantity} {item.unit}
                    </Text>
                    <Text style={s.tdRight}>{fmtIDR(item.unit_cost)}</Text>
                    <Text style={s.tdRight}>{fmtIDR(item.subtotal)}</Text>
                  </View>
                ))}
                <View style={s.trTotal}>
                  <Text style={s.tdTotalLabel}>Total</Text>
                  <Text style={s.tdTotal}>{fmtIDR(totalAmount)}</Text>
                </View>
              </View>
            </>
          )}

          {isReceipt && hasPayments && (
            <>
              <Text style={s.sectionTitle}>Riwayat Pembayaran</Text>
              <View style={s.table}>
                <View style={s.th}>
                  <Text style={s.thItem}>Tanggal</Text>
                  <Text style={{ flex: 1, fontSize: 8, fontWeight: 'bold' }}>Oleh</Text>
                  <Text style={s.thRight}>Jumlah</Text>
                </View>
                {termins.map(t =>
                  t.payments.map(r => (
                    <View key={r.id} style={s.tr}>
                      <Text style={s.td}>{fmtDate(r.paid_at ?? r.payment_date ?? '')}</Text>
                      <Text style={{ flex: 1, fontSize: 9 }}>
                        {r.paid_by_name || r.recorded_by_name || '—'}
                      </Text>
                      <Text style={s.tdRight}>{fmtIDR(r.amount ?? r.amount_paid ?? 0)}</Text>
                    </View>
                  ))
                )}
              </View>
            </>
          )}

          {isAgreement && hasTermins && (
            <>
              <Text style={s.sectionTitle}>Jadwal Termin</Text>
              <View style={s.table}>
                <View style={s.th}>
                  <Text style={s.thItem}>Termin</Text>
                  <Text style={{ flex: 1, fontSize: 8, fontWeight: 'bold' }}>Jatuh Tempo</Text>
                  <Text style={s.thRight}>Jumlah</Text>
                  <Text style={s.thCenter}>Status</Text>
                </View>
                {termins.map(t => (
                  <View key={t.id} style={s.tr}>
                    <Text style={s.td}>Termin {t.termin_number}</Text>
                    <Text style={{ flex: 1, fontSize: 9 }}>{fmtDate(t.due_date)}</Text>
                    <Text style={s.tdRight}>{fmtIDR(t.amount)}</Text>
                    <Text style={s.tdCenter}>{sLabel(t)}</Text>
                  </View>
                ))}
              </View>
            </>
          )}

          {isAgreement && (
            <View style={s.agreementBox}>
              <Text style={s.agreeTitle}>Ketentuan Perjanjian</Text>
              <Text style={s.agreeItem}>
                1. Pihak pembeli wajib melunasi setiap termin sesuai tanggal jatuh tempo.
              </Text>
              <Text style={s.agreeItem}>
                2. Keterlambatan pembayaran dapat dikenakan kebijakan tambahan sesuai kesepakatan.
              </Text>
              <Text style={s.agreeItem}>
                3. Dokumen ini berlaku sebagai bukti kesepakatan pembayaran bertahap atas PO ini.
              </Text>
            </View>
          )}

          <View style={s.signatureSection}>
            <View style={s.sigBox}>
              <Text style={s.sigLabel}>Pihak Supplier / Pemberi Barang</Text>
              <Text style={s.sigLine}>Tanda Tangan &amp; Nama</Text>
            </View>
            <View style={s.sigBox}>
              <Text style={s.sigLabel}>Pihak Pembeli / Penerima</Text>
{isReceived && (
                <View style={s.stamp}>
                  <Svg width={110} height={60} viewBox="0 0 110 60">
                    <Rect x={1} y={1} width={108} height={58} rx={6} fill="none" stroke="#ef4444" strokeWidth={2} />
                    <Rect x={4} y={4} width={102} height={52} rx={4} fill="none" stroke="#ef4444" strokeWidth={0.5} />
                  </Svg>
                  <View style={{ position: 'absolute', top: 0, left: 0, right: 0, bottom: 0, alignItems: 'center', justifyContent: 'center' }}>
                    <Text style={{ fontSize: 11, fontWeight: 'bold', color: '#ef4444' }}>VERIFIED</Text>
                    <Text style={{ fontSize: 6, fontWeight: 'bold', color: '#ef4444', marginTop: 1 }}>{data.store_name || 'MoedahPOS'}</Text>
                    <Text style={{ fontSize: 5, color: '#ef4444', marginTop: 0.5 }}>
                      {new Date().toLocaleDateString('id-ID', { day: '2-digit', month: 'short', year: 'numeric' })} {new Date().toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' })}
                    </Text>
                  </View>
                </View>
              )}
              <Text style={s.sigLine}>Tanda Tangan &amp; Nama</Text>
            </View>
          </View>
        </View>

        <View style={s.footer} fixed>
          <Text style={s.footerText}>
            Dokumen ini dibuat secara otomatis oleh sistem MoedahPOS · {po.po_number}
          </Text>
        </View>
      </Page>
    </Document>
  );
}

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
  const [ready, setReady] = useState(false);

  useEffect(() => {
    if (!selectedStore || !poId) return;
    getPODocument(selectedStore.store_id, poId, docType)
      .then(d => {
        setData(d);
        setReady(true);
      })
      .catch(() => setError('Gagal memuat data dokumen.'));
  }, [selectedStore, poId, docType]);

  if (error)
    return (
      <div style={{ padding: 32, color: '#dc2626', fontFamily: 'sans-serif' }}>
        <strong>Error:</strong> {error}
      </div>
    );
  if (!ready || !data)
    return (
      <div style={{ padding: 32, color: '#6b7280', fontFamily: 'sans-serif' }}>
        Memuat dokumen...
      </div>
    );

  const docTitle = DOC_TITLES[docType] ?? 'Dokumen';

  return (
    <div
      style={{
        minHeight: '100vh',
        background: '#f3f4f6',
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        padding: '24px 16px',
        gap: 12,
        fontFamily: 'sans-serif',
      }}
    >
      <div style={{ display: 'flex', gap: 8, alignItems: 'center', width: 675 }}>
        <PDFDownloadLink
          document={<PODocumentPDF data={data} />}
          fileName={`${docTitle.replace(/\s+/g, '_')}_${data.po.po_number}.pdf`}
          style={{ textDecoration: 'none' }}
        >
          {({ loading }) => (
            <button
              disabled={loading}
              style={{
                padding: '8px 18px',
                background: loading ? '#9ca3af' : '#2563eb',
                color: '#fff',
                border: 'none',
                borderRadius: 8,
                cursor: loading ? 'not-allowed' : 'pointer',
                fontWeight: 600,
                fontSize: 13,
              }}
            >
              {loading ? 'Menyiapkan...' : '⬇ Download PDF'}
            </button>
          )}
        </PDFDownloadLink>
        <button
          onClick={() => window.close()}
          style={{
            padding: '8px 18px',
            background: '#6b7280',
            color: '#fff',
            border: 'none',
            borderRadius: 8,
            cursor: 'pointer',
            fontWeight: 600,
            fontSize: 13,
          }}
        >
          Tutup
        </button>
      </div>

      <div
        style={{
          background: '#fff',
          boxShadow: '0 4px 24px rgba(0,0,0,0.10)',
          borderRadius: 4,
          overflow: 'hidden',
        }}
      >
        <PDFViewer width={675} height={842} style={{ border: 'none' }}>
          <PODocumentPDF data={data} />
        </PDFViewer>
      </div>
    </div>
  );
}
