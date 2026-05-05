import type { Metadata } from 'next';
import FaqSection from '@/components/seo/FaqSection';
import RelatedPages from '@/components/seo/RelatedPages';
import SeoJsonLd from '@/components/seo/SeoJsonLd';
import WaCta from '@/components/seo/WaCta';
import { buildMetadata } from '@/lib/seo';
import { seoPages } from '@/lib/seo-pages';

export const metadata: Metadata = buildMetadata(seoPages['aplikasi-pos']);

export default function AplikasiPosPage() {
  return (
    <main className="mx-auto max-w-4xl px-4 py-12">
      <h1 className="text-3xl font-bold">Aplikasi POS untuk Operasional Bisnis yang Lebih Cepat</h1>
      <p className="mt-4 text-base text-gray-700">
        Kelola transaksi, stok, dan laporan dalam satu sistem untuk UMKM dan restoran.
      </p>
      <section className="mt-8 space-y-4 text-gray-700">
        <h2 className="text-2xl font-semibold text-gray-900">
          Kenapa bisnis butuh aplikasi POS modern?
        </h2>
        <p>
          Proses kasir manual memperlambat antrean, membuat pencatatan stok tidak sinkron, dan
          menyulitkan owner memantau performa harian.
        </p>
        <h3 className="text-xl font-semibold text-gray-900">
          Masalah kasir manual yang paling sering terjadi
        </h3>
        <p>Human error input harga, nota hilang, dan rekap akhir hari yang memakan waktu.</p>
        <h3 className="text-xl font-semibold text-gray-900">
          Dampak ke omzet, stok, dan kontrol owner
        </h3>
        <p>Keputusan belanja stok terlambat, margin bocor, dan peluang repeat order menurun.</p>
      </section>
      <RelatedPages current="aplikasi-pos" />
      <FaqSection pageKey="aplikasi-pos" />
      <WaCta
        title="Mau lihat demo aplikasi POS sesuai bisnis Anda?"
        description="Cerita singkat tipe usaha Anda, tim kami bantu mapping alur operasional terbaik via WhatsApp."
        buttonText="Chat WhatsApp untuk Demo"
        waMessage="Halo Moedah POS, saya ingin demo aplikasi POS untuk bisnis saya."
      />
      <SeoJsonLd pageKey="aplikasi-pos" />
    </main>
  );
}
