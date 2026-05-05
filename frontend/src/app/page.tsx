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
  alternates: {
    canonical: 'https://moedah.com/',
  },
  openGraph: {
    title: 'Moedah POS',
    description: 'Kelola Bisnis Tanpa Henti, Akuntansi Tanpa Pusing.',
    type: 'website',
    url: 'https://moedah.com/',
    locale: 'id_ID',
    siteName: 'Moedah POS',
    images: [
      {
        url: 'https://moedah.com/og-default.jpg',
        width: 1200,
        height: 630,
        alt: 'Moedah POS',
      },
    ],
  },
  twitter: {
    card: 'summary_large_image',
    title: 'Moedah POS - Aplikasi Kasir & Akuntansi No. 1 di Indonesia',
    description:
      'Sistem POS modern dengan mode offline otomatis, sistem akuntansi terpadu, dan manajemen loyalitas pelanggan untuk bisnis retail dan F&B di Indonesia.',
    images: ['https://moedah.com/og-default.jpg'],
  },
};

export default function LandingPage() {
  const schema = {
    '@context': 'https://schema.org',
    '@graph': [
      {
        '@type': 'Organization',
        name: 'Moedah POS',
        url: 'https://moedah.com/',
        logo: 'https://moedah.com/logo.png',
      },
      {
        '@type': 'WebSite',
        name: 'Moedah POS',
        url: 'https://moedah.com/',
        inLanguage: 'id-ID',
      },
    ],
  };

  return (
    <div className="bg-[#f8f9fb] selection:bg-[#4f6ef7]/30 selection:text-white">
      <Navbar />
      <main>
        <Hero />
        <FeatureSection />
        <PricingSection />
      </main>
      <Footer />
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(schema) }}
      />
    </div>
  );
}
