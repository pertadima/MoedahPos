import type { Metadata } from 'next';
import Navbar from '@/components/landing/Navbar';
import Hero from '@/components/landing/Hero';
import BentoGrid from '@/components/landing/BentoGrid';
import TrustSection from '@/components/landing/TrustSection';
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
    <div className="min-h-screen bg-white dark:bg-black selection:bg-[#0070F3] selection:text-white">
      <Navbar />
      <main>
        <Hero />
        <BentoGrid />
        <TrustSection />
      </main>
      <Footer />
    </div>
  );
}
