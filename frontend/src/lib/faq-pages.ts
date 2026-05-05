import type { SeoPageKey } from '@/lib/seo-pages';

export const faqByPage: Record<SeoPageKey, Array<{ q: string; a: string }>> = {
  'aplikasi-pos': [
    {
      q: 'Apa itu aplikasi POS?',
      a: 'Aplikasi POS adalah sistem kasir digital untuk transaksi, stok, dan laporan penjualan.',
    },
    {
      q: 'Apakah Moedah POS cocok untuk UMKM?',
      a: 'Ya, dirancang untuk UMKM agar operasional lebih cepat dan mudah dikontrol.',
    },
    {
      q: 'Berapa lama implementasi awal?',
      a: 'Implementasi bergantung kebutuhan, umumnya dapat dimulai cepat setelah setup produk awal.',
    },
  ],
  'aplikasi-kasir-digital': [
    {
      q: 'Bedanya kasir digital vs kasir manual?',
      a: 'Kasir digital mempercepat transaksi dan merapikan pencatatan otomatis dibanding manual.',
    },
    {
      q: 'Apakah perlu training lama untuk staf?',
      a: 'Tidak, alur penggunaan dibuat sederhana untuk onboarding cepat.',
    },
    {
      q: 'Bagaimana cara mulai demo?',
      a: 'Anda bisa chat WhatsApp untuk jadwal demo sesuai kebutuhan bisnis.',
    },
  ],
  'aplikasi-pos-restoran': [
    {
      q: 'Apa fitur wajib aplikasi POS restoran?',
      a: 'Fitur penting mencakup alur order, manajemen menu, dan laporan penjualan.',
    },
    {
      q: 'Apakah cocok untuk restoran kecil?',
      a: 'Ya, cocok untuk warung makan hingga restoran skala UMKM.',
    },
    {
      q: 'Bisa atur variasi menu?',
      a: 'Bisa, termasuk varian dan catatan pesanan sesuai kebutuhan operasional.',
    },
  ],
  'aplikasi-pos-umkm': [
    {
      q: 'Apa aplikasi POS cocok untuk UMKM baru?',
      a: 'Cocok, terutama untuk merapikan operasional sejak awal.',
    },
    {
      q: 'Apakah perlu perangkat khusus?',
      a: 'Kebutuhan perangkat mengikuti alur bisnis; konsultasi membantu menentukan setup.',
    },
    {
      q: 'Bisa dipakai pemilik usaha non-teknis?',
      a: 'Bisa, karena alur dibuat sederhana dan mudah dipelajari.',
    },
  ],
  'moka-pos-alternatif': [
    {
      q: 'Apa yang harus dibandingkan saat memilih alternatif POS?',
      a: 'Bandingkan kecocokan fitur, onboarding tim, dan kualitas support operasional.',
    },
    {
      q: 'Kapan waktu tepat migrasi sistem kasir?',
      a: 'Saat kebutuhan operasional tidak lagi terjawab sistem saat ini.',
    },
    {
      q: 'Bagaimana meminimalkan gangguan saat transisi?',
      a: 'Lakukan trial alur utama dan libatkan kasir/manager sebelum go-live.',
    },
  ],
  'harga-aplikasi-pos': [
    {
      q: 'Berapa kisaran harga aplikasi POS untuk UMKM?',
      a: 'Kisaran bervariasi tergantung fitur, skala operasional, dan kebutuhan onboarding.',
    },
    {
      q: 'Faktor apa yang memengaruhi biaya?',
      a: 'Komponen utama biasanya software, perangkat, dan implementasi awal.',
    },
    {
      q: 'Bagaimana menilai ROI aplikasi POS?',
      a: 'Nilai dari efisiensi operasional, akurasi data, dan kecepatan pengambilan keputusan.',
    },
  ],
};
