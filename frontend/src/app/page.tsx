import type { Metadata } from 'next';
import Navbar from '@/components/landing/Navbar';
import Hero from '@/components/landing/Hero';
import FeatureSection from '@/components/landing/FeatureSection';
import PricingSection from '@/components/landing/PricingSection';
import Footer from '@/components/landing/Footer';

export const metadata: Metadata = {
  title: 'Moedah POS - Aplikasi Kasir & Akuntansi No. 1 di Indonesia',
  description:
    'Sistem POS modern dengan mode offline otomatis, sistem akuntansi terpadu, dan manajemen loyalitas pelanggan untuk bisnis retail dan F&B di Indonesia.',
  keywords: ['POS', 'Point of Sale', 'Aplikasi Kasir', 'Akuntansi', 'Offline POS', 'Indonesia'],
  openGraph: {
    title: 'Moedah POS',
    description: 'Kelola Bisnis Tanpa Henti, Akuntansi Tanpa Pusing.',
    type: 'website',
    locale: 'id_ID',
  },
};

export default function LandingPage() {
  return (
    <div className="bg-[#f8f9fb] dark:bg-[#09090b] selection:bg-[#4f6ef7]/30 selection:text-white transition-colors duration-500">
      <Navbar />
      <main>
        <Hero />
        <FeatureSection />
        <PricingSection />
      </main>
      <Footer />
    </div>
  );
}
