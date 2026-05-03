'use client';

import { motion } from 'framer-motion';
import { WifiOff, Calculator, HeartHandshake, ShieldCheck } from 'lucide-react';

const features = [
  {
    title: 'Internet Mati? Tidak Masalah.',
    description:
      'Lanjutkan transaksi, cetak struk, dan layani pelanggan tanpa hambatan. Data akan otomatis tersinkronisasi dengan aman saat koneksi internet kembali.',
    icon: WifiOff,
    colSpan: 'md:col-span-2 lg:col-span-2',
    bgClass: 'bg-gradient-to-br from-blue-500 to-[#0070F3] text-white',
    iconClass: 'text-white/80',
    titleClass: 'text-white',
    descClass: 'text-white/90',
  },
  {
    title: 'Laporan Keuangan Otomatis',
    description:
      'Setiap transaksi langsung tercatat di P&L dan Balance Sheet. Export ke PDF & CSV dalam satu klik.',
    icon: Calculator,
    colSpan: 'md:col-span-1 lg:col-span-1',
    bgClass: 'bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800',
    iconClass: 'text-[#0070F3]',
    titleClass: 'text-gray-900 dark:text-white',
    descClass: 'text-gray-600 dark:text-gray-400',
  },
  {
    title: 'Kenali Pelanggan Anda',
    description:
      'Program loyalitas, riwayat poin, dan tingkatan member (Silver, Gold) untuk menjaga pelanggan tetap kembali.',
    icon: HeartHandshake,
    colSpan: 'md:col-span-1 lg:col-span-1',
    bgClass: 'bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800',
    iconClass: 'text-rose-500',
    titleClass: 'text-gray-900 dark:text-white',
    descClass: 'text-gray-600 dark:text-gray-400',
  },
  {
    title: 'Akses Terkontrol (RBAC)',
    description:
      'Atur hak akses kasir, manager, dan admin. Cegah kecurangan dengan izin spesifik untuk void, diskon, dan retur.',
    icon: ShieldCheck,
    colSpan: 'md:col-span-2 lg:col-span-2',
    bgClass: 'bg-gray-50 dark:bg-gray-800/50 border border-gray-200 dark:border-gray-800',
    iconClass: 'text-emerald-500',
    titleClass: 'text-gray-900 dark:text-white',
    descClass: 'text-gray-600 dark:text-gray-400',
  },
];

export default function BentoGrid() {
  return (
    <section className="py-24 bg-white dark:bg-black">
      <div className="max-w-7xl mx-auto px-6">
        <div className="text-center max-w-2xl mx-auto mb-16">
          <h2 className="text-3xl sm:text-4xl font-bold text-gray-900 dark:text-white tracking-tight mb-4">
            Fondasi Bisnis yang Tangguh
          </h2>
          <p className="text-lg text-gray-600 dark:text-gray-400">
            Sistem POS yang dirancang khusus untuk menghadapi tantangan unik bisnis retail dan F&B
            di Indonesia.
          </p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          {features.map((feature, index) => (
            <motion.div
              key={index}
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true, margin: '-50px' }}
              transition={{ duration: 0.5, delay: index * 0.1 }}
              className={`rounded-3xl p-8 flex flex-col justify-between overflow-hidden relative group ${feature.colSpan} ${feature.bgClass}`}
            >
              <div className="relative z-10">
                <div
                  className={`w-12 h-12 rounded-xl flex items-center justify-center mb-6 bg-white/10 ${feature.bgClass.includes('text-white') ? '' : 'bg-gray-100 dark:bg-gray-800'}`}
                >
                  <feature.icon size={24} className={feature.iconClass} />
                </div>
                <h3 className={`text-2xl font-bold mb-3 ${feature.titleClass}`}>{feature.title}</h3>
                <p className={`text-base leading-relaxed ${feature.descClass}`}>
                  {feature.description}
                </p>
              </div>

              {/* Subtle hover effect overlay */}
              <div className="absolute inset-0 bg-black/0 group-hover:bg-black/[0.02] dark:group-hover:bg-white/[0.02] transition-colors duration-300 pointer-events-none" />
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
}
