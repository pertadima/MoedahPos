import Link from 'next/link';
import type { SeoPageKey } from '@/lib/seo-pages';

type RelatedPagesProps = {
  current: SeoPageKey;
  title?: string;
};

const relatedMap: Record<SeoPageKey, Array<{ href: string; anchor: string }>> = {
  'aplikasi-pos': [
    { href: '/aplikasi-kasir-digital/', anchor: 'Aplikasi kasir digital' },
    { href: '/aplikasi-pos-restoran/', anchor: 'POS untuk restoran' },
    { href: '/aplikasi-pos-umkm/', anchor: 'POS untuk UMKM' },
    { href: '/harga-aplikasi-pos/', anchor: 'Panduan harga aplikasi POS' },
  ],
  'aplikasi-kasir-digital': [
    { href: '/aplikasi-pos/', anchor: 'Aplikasi POS' },
    { href: '/aplikasi-pos-umkm/', anchor: 'Kasir digital untuk UMKM' },
    { href: '/harga-aplikasi-pos/', anchor: 'Biaya aplikasi kasir' },
    { href: '/moka-pos-alternatif/', anchor: 'Alternatif Moka POS' },
  ],
  'aplikasi-pos-restoran': [
    { href: '/aplikasi-pos/', anchor: 'Aplikasi POS utama' },
    { href: '/aplikasi-kasir-digital/', anchor: 'Kasir digital cepat' },
    { href: '/harga-aplikasi-pos/', anchor: 'Harga POS restoran' },
    { href: '/moka-pos-alternatif/', anchor: 'Perbandingan solusi POS' },
  ],
  'aplikasi-pos-umkm': [
    { href: '/aplikasi-pos/', anchor: 'Aplikasi POS untuk bisnis' },
    { href: '/aplikasi-kasir-digital/', anchor: 'Kasir digital praktis' },
    { href: '/harga-aplikasi-pos/', anchor: 'Pilihan paket POS' },
    { href: '/moka-pos-alternatif/', anchor: 'Alternatif POS untuk UMKM' },
  ],
  'moka-pos-alternatif': [
    { href: '/aplikasi-pos/', anchor: 'Lihat aplikasi POS Moedah' },
    { href: '/aplikasi-pos-restoran/', anchor: 'POS untuk restoran' },
    { href: '/aplikasi-pos-umkm/', anchor: 'POS untuk UMKM' },
    { href: '/harga-aplikasi-pos/', anchor: 'Cek panduan harga POS' },
  ],
  'harga-aplikasi-pos': [
    { href: '/aplikasi-pos/', anchor: 'Aplikasi POS lengkap' },
    { href: '/aplikasi-kasir-digital/', anchor: 'Kasir digital untuk operasional' },
    { href: '/aplikasi-pos-umkm/', anchor: 'POS sesuai kebutuhan UMKM' },
    { href: '/moka-pos-alternatif/', anchor: 'Bandingkan opsi POS' },
  ],
};

export default function RelatedPages({ current, title = 'Halaman terkait' }: RelatedPagesProps) {
  const links = relatedMap[current];

  return (
    <section
      aria-labelledby="related-pages-title"
      className="mt-10 rounded-2xl border border-[#0884F6]/15 bg-[#f8fbff] p-6"
    >
      <h2 id="related-pages-title" className="text-2xl font-semibold text-[#0b3f6f]">
        {title}
      </h2>
      <ul className="mt-4 grid gap-3 md:grid-cols-2">
        {links.map(link => (
          <li key={link.href}>
            <Link
              href={link.href}
              className="block rounded-lg border border-[#0884F6]/10 bg-white px-3 py-2 font-medium text-[#0884F6] hover:border-[#0884F6]/30"
            >
              {link.anchor}
            </Link>
          </li>
        ))}
      </ul>
    </section>
  );
}
