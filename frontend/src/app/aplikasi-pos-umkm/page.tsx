import type { Metadata } from 'next';
import FaqSection from '@/components/seo/FaqSection';
import RelatedPages from '@/components/seo/RelatedPages';
import SeoJsonLd from '@/components/seo/SeoJsonLd';
import WaCta from '@/components/seo/WaCta';
import { buildMetadata } from '@/lib/seo';
import { seoPages } from '@/lib/seo-pages';

export const metadata: Metadata = buildMetadata(seoPages['aplikasi-pos-umkm']);

export default function AplikasiPosUmkmPage() {
  return (
    <main className="mx-auto max-w-4xl px-4 py-12">
      <h1 className="text-3xl font-bold">Aplikasi POS UMKM untuk Bantu Bisnis Naik Kelas</h1>
      <p className="mt-4 text-base text-gray-700">
        Permudah transaksi harian, kontrol stok, dan pantau laporan dengan lebih praktis.
      </p>
      <section className="mt-8 space-y-4 text-gray-700">
        <h2 className="text-2xl font-semibold text-gray-900">Kenapa UMKM perlu beralih ke POS</h2>
        <p>
          Catatan manual membuat owner sulit membaca tren produk laris dan arus kas operasional.
        </p>
        <h3 className="text-xl font-semibold text-gray-900">
          Manfaat utama untuk bisnis kecil-menengah
        </h3>
        <p>
          Operasional lebih sederhana, pengambilan keputusan lebih cepat, dan kontrol usaha lebih
          presisi.
        </p>
      </section>
      <RelatedPages current="aplikasi-pos-umkm" />
      <FaqSection pageKey="aplikasi-pos-umkm" />
      <WaCta
        title="Siap digitalisasi kasir UMKM?"
        description="Tim Moedah bantu pilih setup yang sesuai skala usaha dan ritme operasional Anda."
        buttonText="Chat WhatsApp untuk Mulai"
        waMessage="Halo Moedah POS, saya ingin konsultasi POS untuk UMKM."
      />
      <SeoJsonLd pageKey="aplikasi-pos-umkm" />
    </main>
  );
}
