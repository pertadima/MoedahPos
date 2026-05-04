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

const enterpriseFeatures = [
  'Deploy & Setup Instan',
  'Hosting & Cloud Sync 24/7',
  'Update & Maintenance Rutin',
  'Training Staf Kasir & Admin',
  'Priority WhatsApp Support',
  'Integrasi WhatsApp untuk Laporan',
  'Konektivitas Printer Bluetooth Thermal',
  'Fitur Kustom Sesuai Kebutuhan',
  'Keamanan Data Terjamin',
];

export default function PricingSection() {
  const waNumber = process.env.NEXT_PUBLIC_WA_NUMBER;
  const waMessage = process.env.NEXT_PUBLIC_WA_MESSAGE;
  const waLink = waNumber
    ? `https://wa.me/${waNumber}?text=${encodeURIComponent(waMessage || '')}`
    : '#';

  return (
    <section className="py-32 bg-gray-50">
      <div className="max-w-7xl mx-auto px-6 text-center">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6 }}
          className="mb-20"
        >
          <p className="text-xs font-bold tracking-[0.2em] uppercase text-[#FFA724] mb-4">
            Transparent Pricing
          </p>
          <h2 className="text-3xl sm:text-5xl font-extrabold text-gray-900 tracking-tight">
            Transparan.
            <br />
            <span className="text-gray-500 font-normal">
              Pilih yang sesuai untuk skala bisnis Anda.
            </span>
          </h2>
        </motion.div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6 max-w-4xl mx-auto text-left">
          <motion.div
            initial={{ opacity: 0, x: -20 }}
            whileInView={{ opacity: 1, x: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.5, delay: 0.1 }}
            className="relative flex flex-col p-8 rounded-2xl bg-white shadow-xl ring-1 ring-gray-200"
          >
            <div className="mb-6">
              <p className="text-xs font-bold tracking-widest uppercase text-[#FFA724] mb-4">
                Open Source
              </p>
              <div className="flex items-baseline gap-2 mb-3">
                <span className="text-4xl font-extrabold text-gray-900">Free</span>
                <span className="text-gray-500 font-medium">/ selamanya</span>
              </div>
              <p className="text-gray-500 leading-relaxed text-sm">
                Bagi tim tech-savvy. Download, deploy, dan kelola sendiri di infrastruktur Anda.
              </p>
            </div>

            <div className="space-y-3 mb-8 flex-1">
              {freeFeatures.map((f, i) => (
                <motion.div
                  key={f}
                  initial={{ opacity: 0, x: -10 }}
                  whileInView={{ opacity: 1, x: 0 }}
                  viewport={{ once: true }}
                  transition={{ duration: 0.3, delay: 0.15 + i * 0.05 }}
                  className="flex items-center gap-3 text-sm text-gray-600"
                >
                  <CheckCircle2 size={16} className="text-[#FFA724] shrink-0" />
                  {f}
                </motion.div>
              ))}
            </div>

            <a
              href="https://github.com/pertadima/MoedahPos"
              target="_blank"
              rel="noopener noreferrer"
              className="w-full flex items-center justify-center gap-2 h-11 rounded-xl bg-gray-900 text-white text-sm font-bold transition-all duration-200 hover:bg-gray-800 hover:shadow-lg"
            >
              <Code size={18} />
              Lihat di GitHub
            </a>
          </motion.div>

          <motion.div
            initial={{ opacity: 0, x: 20 }}
            whileInView={{ opacity: 1, x: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.5, delay: 0.2 }}
            className="relative flex flex-col p-8 rounded-2xl bg-gradient-to-br from-[#0884F6] to-[#0770d4] shadow-2xl"
          >
            <div className="absolute -top-3 left-1/2 -translate-x-1/2">
              <span className="px-4 py-1.5 bg-[#FFA724] text-white text-[10px] font-bold rounded-full tracking-wider shadow-lg">
                PRO RECOMMENDED
              </span>
            </div>

            <div className="mb-6 mt-2">
              <p className="text-xs font-bold tracking-widest uppercase text-white/70 mb-4">
                Managed Service
              </p>
              <div className="flex items-baseline gap-2 mb-3">
                <span className="text-4xl font-extrabold text-white">Custom</span>
              </div>
              <p className="text-white/70 leading-relaxed text-sm">
                Fokus jualan, kami yang urus server. Deploy & support penuh dari tim ahli kami.
              </p>
            </div>

            <div className="space-y-3 mb-8 flex-1 border-t border-white/20 pt-6">
              {enterpriseFeatures.map((f, i) => (
                <motion.div
                  key={f}
                  initial={{ opacity: 0, x: -10 }}
                  whileInView={{ opacity: 1, x: 0 }}
                  viewport={{ once: true }}
                  transition={{ duration: 0.3, delay: 0.25 + i * 0.05 }}
                  className="flex items-center gap-3 text-sm text-white/80"
                >
                  <CheckCircle2 size={16} className="text-[#FFA724] shrink-0" />
                  {f}
                </motion.div>
              ))}
            </div>

            <Link
              href={waLink}
              target="_blank"
              rel="noopener noreferrer"
              className="w-full flex items-center justify-center gap-2 h-11 rounded-xl bg-[#FFA724] text-white text-sm font-bold transition-all duration-200 hover:bg-[#e69520] hover:shadow-lg"
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
