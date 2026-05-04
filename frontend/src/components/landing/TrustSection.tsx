'use client';

import { motion, type Variants } from 'framer-motion';
import { WifiOff, Receipt, ShieldCheck, Star } from 'lucide-react';

const pillars = [
  {
    icon: WifiOff,
    title: 'Offline First',
    desc: 'Bertransaksi tanpa internet. Semua data tersinkron otomatis saat koneksi kembali.',
  },
  {
    icon: Receipt,
    title: 'Patuh Pajak',
    desc: 'PPN/PB1 dikalkulasi dan dicetak di setiap struk. Siap audit kapan saja.',
  },
  {
    icon: ShieldCheck,
    title: 'Akses Aman',
    desc: 'RBAC per-modul. Kasir, Manager, Finance — masing-masing hanya melihat yang perlu.',
  },
  {
    icon: Star,
    title: 'Loyalitas',
    desc: 'Program poin dan tier keanggotaan terintegrasi langsung ke setiap transaksi.',
  },
];

const methods = ['Tunai', 'QRIS', 'GoPay', 'OVO', 'Dana', 'Transfer Bank'];

const containerVariants: Variants = {
  hidden: {},
  visible: { transition: { staggerChildren: 0.08 } },
};

const itemVariants: Variants = {
  hidden: { opacity: 0, y: 16 },
  visible: {
    opacity: 1,
    y: 0,
    transition: { duration: 0.5, ease: 'easeOut' },
  },
};

export default function TrustSection() {
  return (
    <section
      className="
        py-32 bg-[#0d0d12]
        font-[family-name:var(--font-instrument)]
        border-t border-white/[0.06]
        relative overflow-hidden
      "
    >
      {/* Bottom section glow */}
      <div className="absolute bottom-0 right-0 w-[500px] h-[400px] bg-[#4f6ef7]/[0.05] blur-[100px] rounded-full pointer-events-none" />

      <div className="relative z-10 max-w-7xl mx-auto px-6">
        {/* Section heading */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6, ease: 'easeOut' }}
          className="mb-16"
        >
          <p className="text-xs font-semibold tracking-[0.2em] uppercase text-[#4f6ef7] mb-4">
            Dirancang untuk Indonesia
          </p>
          <h2 className="font-[family-name:var(--font-syne)] text-3xl sm:text-4xl font-bold text-white tracking-tight leading-tight">
            Semua yang dibutuhkan
            <br />
            <span className="text-slate-400 font-normal">bisnis modern.</span>
          </h2>
        </motion.div>

        <div className="grid grid-cols-1 lg:grid-cols-5 gap-12 items-start">
          {/* Feature pillar cards — 3 cols */}
          <motion.div
            variants={containerVariants}
            initial="hidden"
            whileInView="visible"
            viewport={{ once: true }}
            className="lg:col-span-3 grid grid-cols-1 sm:grid-cols-2 gap-4"
          >
            {pillars.map(p => (
              <motion.div
                key={p.title}
                variants={itemVariants}
                className="
                  group p-5 rounded-2xl
                  bg-[#13131a] border border-white/[0.07]
                  hover:border-[#4f6ef7]/30 hover:bg-[#14141e]
                  transition-all duration-300
                "
              >
                <div className="w-9 h-9 rounded-xl bg-[#4f6ef7]/[0.12] border border-[#4f6ef7]/20 flex items-center justify-center mb-4">
                  <p.icon size={15} className="text-[#4f6ef7]" />
                </div>
                <h3 className="font-[family-name:var(--font-syne)] text-sm font-semibold text-white mb-2">
                  {p.title}
                </h3>
                <p className="text-xs text-slate-500 leading-relaxed">{p.desc}</p>
              </motion.div>
            ))}
          </motion.div>

          {/* Right column — payments + tax — 2 cols */}
          <motion.div
            initial={{ opacity: 0, x: 20 }}
            whileInView={{ opacity: 1, x: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.6, delay: 0.2, ease: 'easeOut' }}
            className="lg:col-span-2 space-y-6"
          >
            {/* Payment methods */}
            <div className="p-6 rounded-2xl bg-[#13131a] border border-white/[0.07]">
              <p className="text-[10px] font-semibold tracking-[0.18em] uppercase text-slate-600 mb-5">
                Metode Pembayaran
              </p>
              <div className="flex flex-wrap gap-2">
                {methods.map(m => (
                  <span
                    key={m}
                    className="h-8 px-3.5 flex items-center text-xs font-medium
                      rounded-full border border-white/[0.08] bg-white/[0.04]
                      text-slate-300"
                  >
                    {m}
                  </span>
                ))}
              </div>
            </div>

            {/* Tax compliance */}
            <div className="p-6 rounded-2xl bg-[#13131a] border border-white/[0.07]">
              <p className="text-[10px] font-semibold tracking-[0.18em] uppercase text-slate-600 mb-3">
                Pajak Otomatis
              </p>
              <p className="text-sm text-slate-400 leading-relaxed">
                PPN & PB1 dikalkulasi dan ditampilkan di setiap struk berdasarkan persentase yang
                Anda atur per toko. Siap untuk kebutuhan pelaporan pajak.
              </p>
            </div>

            {/* Export CTA highlight */}
            <div className="p-6 rounded-2xl bg-[#4f6ef7]/[0.08] border border-[#4f6ef7]/20">
              <p className="text-[10px] font-semibold tracking-[0.18em] uppercase text-[#4f6ef7] mb-3">
                Export Laporan
              </p>
              <p className="text-sm text-slate-400 leading-relaxed">
                Unduh laporan keuangan dalam format{' '}
                <span className="text-white font-medium">PDF</span> atau{' '}
                <span className="text-white font-medium">CSV</span> langsung dari dashboard.
              </p>
            </div>
          </motion.div>
        </div>
      </div>
    </section>
  );
}
