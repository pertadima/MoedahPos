import type { Metadata } from 'next';
import FaqSection from '@/components/seo/FaqSection';
import RelatedPages from '@/components/seo/RelatedPages';
import SeoJsonLd from '@/components/seo/SeoJsonLd';
import WaCta from '@/components/seo/WaCta';
import { buildMetadata } from '@/lib/seo';
import { seoPages } from '@/lib/seo-pages';

export const metadata: Metadata = buildMetadata(seoPages['harga-aplikasi-pos']);

export default function HargaAplikasiPosPage() {
  return (
    <main className="mx-auto max-w-4xl px-4 py-12">
      <h1 className="text-3xl font-bold">
        Harga Aplikasi POS: Cara Hitung yang Tepat untuk Bisnis Anda
      </h1>
      <p className="mt-4 text-base text-gray-700">
        Pahami komponen biaya dan pilih paket yang sesuai tahap pertumbuhan bisnis.
      </p>
      <section className="mt-8 space-y-4 text-gray-700">
        <h2 className="text-2xl font-semibold text-gray-900">Komponen biaya yang perlu dipahami</h2>
        <p>
          Evaluasi biaya software, perangkat, dan onboarding agar investasi sesuai kapasitas bisnis.
        </p>
        <h3 className="text-xl font-semibold text-gray-900">
          Fokus ke nilai, bukan hanya harga termurah
        </h3>
        <p>
          Nilai terbaik datang dari efisiensi operasional, akurasi data, dan kecepatan keputusan
          owner.
        </p>
      </section>
      <RelatedPages current="harga-aplikasi-pos" />
      <FaqSection pageKey="harga-aplikasi-pos" />
      <WaCta
        title="Mau hitung kebutuhan POS yang pas?"
        description="Konsultasikan budget dan target operasional Anda, tim kami bantu rekomendasi paket via WhatsApp."
        buttonText="Chat WhatsApp untuk Hitung"
        waMessage="Halo Moedah POS, saya ingin konsultasi harga aplikasi POS."
      />
      <SeoJsonLd pageKey="harga-aplikasi-pos" />
    </main>
  );
}
