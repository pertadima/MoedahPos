import type { Metadata } from 'next';
import FaqSection from '@/components/seo/FaqSection';
import RelatedPages from '@/components/seo/RelatedPages';
import SeoJsonLd from '@/components/seo/SeoJsonLd';
import WaCta from '@/components/seo/WaCta';
import { buildMetadata } from '@/lib/seo';
import { seoPages } from '@/lib/seo-pages';

export const metadata: Metadata = buildMetadata(seoPages['aplikasi-pos-restoran']);

export default function AplikasiPosRestoranPage() {
  return (
    <main className="mx-auto max-w-4xl px-4 py-12">
      <h1 className="text-3xl font-bold">
        Aplikasi POS Restoran untuk Layanan Cepat dan Minim Error
      </h1>
      <p className="mt-4 text-base text-gray-700">
        Kelola menu, order, dan pembayaran restoran dengan alur operasional yang lebih rapi.
      </p>
      <section className="mt-8 space-y-4 text-gray-700">
        <h2 className="text-2xl font-semibold text-gray-900">
          Tantangan operasional restoran sehari-hari
        </h2>
        <p>
          Pesanan rawan salah saat ramai, sinkronisasi tim kasir-dapur kurang mulus, dan kontrol
          menu sulit.
        </p>
        <h3 className="text-xl font-semibold text-gray-900">Alur restoran yang lebih stabil</h3>
        <p>
          Pengelolaan menu, catatan pesanan, dan ringkasan penjualan harian jadi lebih terstruktur.
        </p>
      </section>
      <RelatedPages current="aplikasi-pos-restoran" />
      <FaqSection pageKey="aplikasi-pos-restoran" />
      <WaCta
        title="Butuh setup POS khusus restoran?"
        description="Diskusikan alur menu, order, dan pembayaran restoran Anda langsung via WhatsApp."
        buttonText="Chat WhatsApp untuk Konsultasi"
        waMessage="Halo Moedah POS, saya ingin demo aplikasi POS restoran."
      />
      <SeoJsonLd pageKey="aplikasi-pos-restoran" />
    </main>
  );
}
