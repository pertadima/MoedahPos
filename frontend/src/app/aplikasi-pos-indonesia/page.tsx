import type { Metadata } from 'next';
import Link from 'next/link';
import WaCta from '@/components/seo/WaCta';
import { buildMetadata } from '@/lib/seo';

export const metadata: Metadata = buildMetadata({
  path: '/aplikasi-pos-indonesia/',
  title: 'Aplikasi POS Indonesia untuk UMKM & Restoran | Moedah POS',
  description:
    'Panduan memilih aplikasi POS Indonesia untuk UMKM dan restoran. Bandingkan kebutuhan, fitur, dan alur implementasi sebelum memilih.',
});

const links = [
  {
    href: '/aplikasi-pos/',
    title: 'Aplikasi POS',
    desc: 'Gambaran umum solusi POS untuk operasional bisnis harian.',
  },
  {
    href: '/aplikasi-kasir-digital/',
    title: 'Aplikasi Kasir Digital',
    desc: 'Fokus pada kecepatan transaksi dan kemudahan operasional kasir.',
  },
  {
    href: '/aplikasi-pos-restoran/',
    title: 'Aplikasi POS Restoran',
    desc: 'Alur khusus untuk menu, order, pembayaran, dan tim restoran.',
  },
  {
    href: '/aplikasi-pos-umkm/',
    title: 'Aplikasi POS UMKM',
    desc: 'Pendekatan praktis untuk bisnis kecil menengah yang ingin naik kelas.',
  },
  {
    href: '/moka-pos-alternatif/',
    title: 'Alternatif Moka POS',
    desc: 'Framework evaluasi sebelum memutuskan migrasi atau adopsi sistem baru.',
  },
  {
    href: '/harga-aplikasi-pos/',
    title: 'Harga Aplikasi POS',
    desc: 'Panduan memahami komponen biaya dan nilai investasi POS.',
  },
];

export default function AplikasiPosIndonesiaPage() {
  const schema = {
    '@context': 'https://schema.org',
    '@graph': [
      {
        '@type': 'WebPage',
        name: 'Aplikasi POS Indonesia untuk UMKM & Restoran',
        inLanguage: 'id-ID',
        url: 'https://moedah.com/aplikasi-pos-indonesia/',
        description:
          'Panduan memilih aplikasi POS Indonesia untuk UMKM dan restoran berdasarkan kebutuhan bisnis.',
      },
      {
        '@type': 'BreadcrumbList',
        itemListElement: [
          { '@type': 'ListItem', position: 1, name: 'Beranda', item: 'https://moedah.com/' },
          {
            '@type': 'ListItem',
            position: 2,
            name: 'Aplikasi POS Indonesia',
            item: 'https://moedah.com/aplikasi-pos-indonesia/',
          },
        ],
      },
    ],
  };

  return (
    <main className="mx-auto max-w-5xl px-4 py-12">
      <h1 className="text-3xl font-bold">Aplikasi POS Indonesia untuk UMKM dan Restoran</h1>
      <p className="mt-4 text-gray-700">
        Halaman hub ini merangkum pilihan konten utama Moedah POS agar Anda bisa memilih materi
        sesuai kebutuhan bisnis, tahap pertumbuhan, dan prioritas operasional.
      </p>

      <section className="mt-8 space-y-4 text-gray-700">
        <h2 className="text-2xl font-semibold text-gray-900">Cara memakai halaman ini</h2>
        <p>
          Jika Anda butuh overview, mulai dari halaman aplikasi POS utama. Jika kebutuhan Anda lebih
          spesifik, lanjut ke halaman restoran, UMKM, perbandingan alternatif, atau panduan harga.
        </p>
      </section>

      <section className="mt-10 grid gap-4 md:grid-cols-2">
        {links.map(link => (
          <article key={link.href} className="rounded-lg border border-gray-200 p-5">
            <h3 className="text-lg font-semibold text-gray-900">
              <Link href={link.href} className="hover:underline">
                {link.title}
              </Link>
            </h3>
            <p className="mt-2 text-gray-700">{link.desc}</p>
          </article>
        ))}
      </section>

      <WaCta
        title="Butuh rekomendasi halaman yang paling relevan?"
        description="Ceritakan tipe usaha dan prioritas Anda. Tim kami bantu arahkan jalur evaluasi tercepat via WhatsApp."
        buttonText="Chat WhatsApp Sekarang"
        waMessage="Halo Moedah POS, saya ingin rekomendasi solusi POS sesuai tipe bisnis saya."
      />

      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(schema) }}
      />
    </main>
  );
}
