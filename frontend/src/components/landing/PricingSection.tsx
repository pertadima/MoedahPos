'use client';

import { motion } from 'framer-motion';
import Link from 'next/link';
import { Code, CheckCircle2, MessageCircle } from 'lucide-react';

const freeFeatures = [
  'Kasir (POS) dengan mode offline',
  'Kalkulator HPP otomatis',
  'Manajemen Purchase Order',
  'Laporan keuangan & arus kas',
  'Kontrol akses (RBAC)',
  'Program loyalitas pelanggan',
  'Export PDF & CSV',
  'Multi-outlet support',
];

export default function PricingSection() {
  return (
    <section className="py-32 bg-white dark:bg-[#09090b] border-t border-gray-200 dark:border-white/[0.05] transition-colors duration-500">
      <div className="max-w-7xl mx-auto px-6 text-center">
        {/* Heading */}
        <div className="mb-20">
          <p className="text-xs font-bold tracking-[0.2em] uppercase text-[#4f6ef7] mb-4">
            Transparent Pricing
          </p>
          <h2 className="text-3xl sm:text-5xl font-extrabold text-gray-900 dark:text-white tracking-tight">
            Transparan.
            <br />
            <span className="text-gray-400 dark:text-gray-500 font-normal">
              Pilih yang sesuai untuk skala bisnis Anda.
            </span>
          </h2>
        </div>

        {/* Cards */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-8 max-w-5xl mx-auto text-left">
          {/* Free plan */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            className="
              relative flex flex-col
              p-10 rounded-3xl
              bg-[#f8f9fb] dark:bg-[#13131a]
              border border-gray-200 dark:border-white/[0.07]
              shadow-sm
            "
          >
            <div className="mb-8">
              <p className="text-xs font-bold tracking-widest uppercase text-gray-400 mb-6">
                Open Source
              </p>
              <div className="flex items-baseline gap-2 mb-4">
                <span className="text-5xl font-extrabold text-gray-900 dark:text-white tracking-tight">
                  Free
                </span>
                <span className="text-gray-500 font-medium">/ selamanya</span>
              </div>
              <p className="text-gray-500 dark:text-gray-400 leading-relaxed text-sm">
                Bagi tim tech-savvy. Download, deploy, dan kelola sendiri di infrastruktur Anda.
              </p>
            </div>

            <div className="space-y-4 mb-10 flex-1">
              {freeFeatures.map(f => (
                <div
                  key={f}
                  className="flex items-center gap-3 text-sm text-gray-600 dark:text-gray-400"
                >
                  <CheckCircle2 size={16} className="text-[#4f6ef7] shrink-0" />
                  {f}
                </div>
              ))}
            </div>

            <a
              href="https://github.com/pertadima/MoedahPos"
              target="_blank"
              rel="noopener noreferrer"
              className="
                w-full flex items-center justify-center gap-2
                h-12 rounded-xl
                bg-gray-900 dark:bg-white
                text-white dark:text-gray-900
                text-sm font-bold
                transition-all duration-200
                hover:shadow-xl
              "
            >
              <Code size={18} />
              Lihat di GitHub
            </a>
          </motion.div>

          {/* Enterprise plan */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ delay: 0.1 }}
            className="
              relative flex flex-col
              p-10 rounded-3xl
              bg-[#1a1a24] dark:bg-[#1a1a24]
              border border-[#2e2e3e] dark:border-[#2e2e3e]
              shadow-2xl shadow-black/40
            "
          >
            <div className="absolute top-10 right-10">
              <span className="px-3 py-1 bg-[#4f6ef7]/20 text-[#4f6ef7] text-[10px] font-bold rounded-full tracking-wider border border-[#4f6ef7]/20">
                PRO RECOMMENDED
              </span>
            </div>

            <div className="mb-8">
              <p className="text-xs font-bold tracking-widest uppercase text-slate-500 mb-6">
                Managed Service
              </p>
              <div className="flex items-baseline gap-2 mb-4">
                <span className="text-5xl font-extrabold text-white tracking-tight">Custom</span>
              </div>
              <p className="text-slate-400 leading-relaxed text-sm">
                Fokus jualan, kami yang urus server. Deploy & support penuh dari tim ahli kami.
              </p>
            </div>

            <div className="space-y-4 mb-10 flex-1 border-t border-white/[0.05] pt-8">
              {[
                'Deploy & Setup Instan',
                'Hosting & Cloud Sync 24/7',
                'Update & Maintenance Rutin',
                'Training Staf Kasir & Admin',
                'Priority WhatsApp Support',
                'Keamanan Data Terjamin',
              ].map(f => (
                <div key={f} className="flex items-center gap-3 text-sm text-slate-300">
                  <CheckCircle2 size={16} className="text-[#4f6ef7] shrink-0" />
                  {f}
                </div>
              ))}
            </div>

            <Link
              href="/login"
              className="
                w-full flex items-center justify-center gap-2
                h-12 rounded-xl
                bg-[#4f6ef7] hover:bg-[#3d56d4]
                text-white text-sm font-bold
                transition-all duration-200
                shadow-lg shadow-[#4f6ef7]/30
              "
            >
              <MessageCircle size={18} />
              Hubungi Kontak Penjualan
            </Link>
          </motion.div>
        </div>
      </div>
    </section>
  );
}
