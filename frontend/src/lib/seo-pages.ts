export type SeoPageKey =
  | 'aplikasi-pos'
  | 'aplikasi-kasir-digital'
  | 'aplikasi-pos-restoran'
  | 'aplikasi-pos-umkm'
  | 'moka-pos-alternatif'
  | 'harga-aplikasi-pos';

export const seoPages: Record<
  SeoPageKey,
  {
    path: `/${string}/`;
    title: string;
    description: string;
    image?: string;
  }
> = {
  'aplikasi-pos': {
    path: '/aplikasi-pos/',
    title: 'Aplikasi POS untuk Bisnis Ritel & F&B | Moedah POS',
    description:
      'Kelola transaksi, stok, dan laporan dalam satu aplikasi POS. Cocok untuk UMKM dan restoran. Chat WhatsApp untuk demo cepat.',
  },
  'aplikasi-kasir-digital': {
    path: '/aplikasi-kasir-digital/',
    title: 'Aplikasi Kasir Digital untuk Transaksi Cepat | Moedah POS',
    description:
      'Tinggalkan pencatatan manual. Gunakan aplikasi kasir digital untuk transaksi, stok, dan laporan otomatis. Konsultasi via WhatsApp.',
  },
  'aplikasi-pos-restoran': {
    path: '/aplikasi-pos-restoran/',
    title: 'Aplikasi POS Restoran untuk Order & Kasir Lebih Rapi | Moedah POS',
    description:
      'Kelola menu, order, dan pembayaran restoran dalam satu sistem POS. Cocok untuk warung makan hingga restoran UMKM. Chat WA untuk demo.',
  },
  'aplikasi-pos-umkm': {
    path: '/aplikasi-pos-umkm/',
    title: 'Aplikasi POS UMKM yang Mudah Dipakai | Moedah POS',
    description:
      'Aplikasi POS untuk UMKM yang ingin transaksi lebih cepat, stok terkontrol, dan laporan jelas. Konsultasi gratis via WhatsApp.',
  },
  'moka-pos-alternatif': {
    path: '/moka-pos-alternatif/',
    title: 'Alternatif Moka POS untuk UMKM & Restoran | Moedah POS',
    description:
      'Bandingkan opsi POS untuk UMKM dan restoran. Lihat fitur, kemudahan pakai, dan kecocokan bisnis sebelum memilih. Chat WA untuk konsultasi.',
  },
  'harga-aplikasi-pos': {
    path: '/harga-aplikasi-pos/',
    title: 'Harga Aplikasi POS: Panduan Pilih Paket | Moedah POS',
    description:
      'Pelajari komponen harga aplikasi POS dan cara memilih paket sesuai skala bisnis Anda. Konsultasi kebutuhan via WhatsApp.',
  },
};
