import type { Metadata } from 'next';
import FaqSection from '@/components/seo/FaqSection';
import RelatedPages from '@/components/seo/RelatedPages';
import SeoPageShell from '@/components/seo/SeoPageShell';
import SeoJsonLd from '@/components/seo/SeoJsonLd';
import WaCta from '@/components/seo/WaCta';
import { buildMetadata } from '@/lib/seo';
import { seoPages } from '@/lib/seo-pages';

export const metadata: Metadata = buildMetadata(seoPages['aplikasi-kasir-digital']);

export default function AplikasiKasirDigitalPage() {
  return (
    <SeoPageShell>
      <h1 className="text-3xl font-bold">
        Aplikasi Kasir Digital untuk Bisnis yang Ingin Serba Praktis
      </h1>
      <p className="mt-4 text-base text-[#2f5c84]">
        Tinggalkan pencatatan manual dengan alur kasir digital yang cepat dan mudah.
      </p>
      <section className="mt-8 space-y-4 text-[#2f5c84]">
        <h2 className="text-2xl font-semibold text-[#0b3f6f]">Kendala kasir manual di lapangan</h2>
        <p>
          Antrian menumpuk saat jam sibuk, rekap tutup kas sering molor, dan data penjualan
          tersebar.
        </p>
        <h3 className="text-xl font-semibold text-[#0b3f6f]">
          Solusi kasir digital untuk operasional harian
        </h3>
        <p>
          Input transaksi lebih cepat, histori lebih mudah ditelusuri, dan owner bisa cek performa
          kapan saja.
        </p>
      </section>
      <RelatedPages current="aplikasi-kasir-digital" />
      <FaqSection pageKey="aplikasi-kasir-digital" />
      <WaCta
        title="Ingin kasir lebih cepat tanpa ribet?"
        description="Konsultasikan kebutuhan outlet Anda dan dapatkan rekomendasi alur kasir digital via WhatsApp."
        buttonText="Chat WhatsApp Sekarang"
        waMessage="Halo Moedah POS, saya ingin konsultasi aplikasi kasir digital."
      />
      <SeoJsonLd pageKey="aplikasi-kasir-digital" />
    </SeoPageShell>
  );
}
