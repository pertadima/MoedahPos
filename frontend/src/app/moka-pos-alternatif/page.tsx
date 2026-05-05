import type { Metadata } from 'next';
import FaqSection from '@/components/seo/FaqSection';
import RelatedPages from '@/components/seo/RelatedPages';
import SeoPageShell from '@/components/seo/SeoPageShell';
import SeoJsonLd from '@/components/seo/SeoJsonLd';
import WaCta from '@/components/seo/WaCta';
import { buildMetadata } from '@/lib/seo';
import { seoPages } from '@/lib/seo-pages';

export const metadata: Metadata = buildMetadata(seoPages['moka-pos-alternatif']);

export default function MokaPosAlternatifPage() {
  return (
    <SeoPageShell>
      <h1 className="text-3xl font-bold">
        Cari Alternatif Moka POS? Ini Hal yang Perlu Dibandingkan
      </h1>
      <p className="mt-4 text-base text-[#2f5c84]">
        Bandingkan fitur, onboarding, dan support operasional untuk pilih sistem yang paling cocok.
      </p>
      <section className="mt-8 space-y-4 text-[#2f5c84]">
        <h2 className="text-2xl font-semibold text-[#0b3f6f]">
          Kapan bisnis mulai mencari alternatif POS
        </h2>
        <p>
          Saat kebutuhan fitur berubah, tim butuh onboarding lebih cepat, atau support kurang
          responsif.
        </p>
        <h3 className="text-xl font-semibold text-[#0b3f6f]">Kerangka evaluasi sebelum migrasi</h3>
        <p>
          Uji alur transaksi utama, libatkan kasir/manager, dan pastikan transisi minim gangguan
          operasional.
        </p>
      </section>
      <RelatedPages current="moka-pos-alternatif" />
      <FaqSection pageKey="moka-pos-alternatif" />
      <WaCta
        title="Masih bingung pilih solusi POS?"
        description="Kirim kebutuhan bisnis Anda, kami bantu framework evaluasi cepat lewat WhatsApp."
        buttonText="Chat WhatsApp untuk Compare"
        waMessage="Halo Moedah POS, saya ingin membandingkan alternatif POS."
      />
      <SeoJsonLd pageKey="moka-pos-alternatif" />
    </SeoPageShell>
  );
}
