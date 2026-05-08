'use client';

interface SystemStampProps {
  label?: string;
  poNumber?: string;
  storeName?: string;
  timestamp?: string;
  variant?: 'supplier' | 'buyer';
}

export default function SystemStamp({
  label,
  poNumber,
  storeName,
  timestamp,
  variant = 'supplier',
}: SystemStampProps) {
  const role = variant === 'supplier' ? 'Pemberi Barang' : 'Penerima / Toko';
  const ts = timestamp
    ? new Date(timestamp).toLocaleString('id-ID', {
        day: '2-digit',
        month: 'short',
        year: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
      })
    : new Date().toLocaleString('id-ID', {
        day: '2-digit',
        month: 'short',
        year: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
      });

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
      {label && (
        <span style={{ fontSize: '0.72rem', fontWeight: 600, color: 'var(--text-2)' }}>
          {label}
        </span>
      )}
      <div
        style={{
          position: 'relative',
          border: '2px solid #16a34a',
          borderRadius: 4,
          padding: '8px 14px',
          background: '#f0fdf4',
          display: 'inline-flex',
          flexDirection: 'column',
          alignItems: 'center',
          gap: 2,
          minWidth: 180,
        }}
      >
        <div
          style={{
            position: 'absolute',
            top: -1,
            left: 16,
            background: '#f0fdf4',
            padding: '0 4px',
          }}
        >
          <span style={{ fontSize: '0.6rem', color: '#16a34a', fontWeight: 700, letterSpacing: 1 }}>
            VERIFIED
          </span>
        </div>

        <div style={{ fontSize: '0.95rem', fontWeight: 800, color: '#15803d', marginTop: 8 }}>
          {role}
        </div>

        <div
          style={{
            width: 120,
            height: 1,
            background: '#16a34a',
            margin: '2px 0',
          }}
        />

        <div
          style={{ fontSize: '0.65rem', color: '#166534', textAlign: 'center', lineHeight: 1.4 }}
        >
          Dokumen ini dibuat secara otomatis oleh sistem
        </div>
        <div style={{ fontSize: '0.7rem', fontWeight: 700, color: '#14532d', textAlign: 'center' }}>
          MoedahPOS
        </div>

        <div
          style={{
            width: 100,
            height: 1,
            background: '#16a34a',
            margin: '2px 0',
          }}
        />

        <div style={{ fontSize: '0.65rem', color: '#166534' }}>Tanggal</div>
        <div style={{ fontSize: '0.7rem', fontWeight: 600, color: '#14532d' }}>{ts}</div>

        {storeName && (
          <>
            <div style={{ fontSize: '0.65rem', color: '#166534' }}>Toko</div>
            <div style={{ fontSize: '0.7rem', fontWeight: 600, color: '#14532d' }}>{storeName}</div>
          </>
        )}

        {poNumber && (
          <>
            <div style={{ fontSize: '0.65rem', color: '#166534' }}>No. PO</div>
            <div style={{ fontSize: '0.7rem', fontWeight: 600, color: '#14532d' }}>{poNumber}</div>
          </>
        )}
      </div>
    </div>
  );
}
