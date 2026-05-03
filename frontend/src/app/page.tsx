import type { Metadata } from 'next';
import Navbar from '@/components/landing/Navbar';
import Hero from '@/components/landing/Hero';
import BentoGrid from '@/components/landing/BentoGrid';
import TrustSection from '@/components/landing/TrustSection';

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

      {/* Simple Footer */}
      <footer className="py-12 bg-white dark:bg-black border-t border-gray-200 dark:border-gray-800">
        <div className="max-w-7xl mx-auto px-6 flex flex-col md:flex-row items-center justify-between gap-4">
          <div className="flex items-center gap-2">
            <div className="w-6 h-6 rounded bg-[#0070F3] flex items-center justify-center text-white font-bold text-xs">
              M
            </div>
            <span className="font-bold text-gray-900 dark:text-white tracking-tight">
              Moedah POS
            </span>
          </div>
          <p className="text-sm text-gray-500">
            &copy; {new Date().getFullYear()} Moedah POS. Seluruh hak cipta dilindungi.
          </p>
        </div>
      </footer>
    </div>
  );
}
