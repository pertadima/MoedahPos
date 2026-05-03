'use client';

import { motion } from 'framer-motion';
import Link from 'next/link';
import { Play, ArrowRight, CheckCircle2 } from 'lucide-react';

export default function Hero() {
  return (
    <section className="relative pt-32 pb-20 lg:pt-48 lg:pb-32 overflow-hidden">
      {/* Background decoration */}
      <div className="absolute inset-0 -z-10 bg-[radial-gradient(ellipse_at_top_right,_var(--tw-gradient-stops))] from-blue-100 via-white to-white dark:from-blue-900/20 dark:via-black dark:to-black" />

      <div className="max-w-7xl mx-auto px-6">
        <div className="grid lg:grid-cols-2 gap-12 lg:gap-8 items-center">
          {/* Text Content */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6 }}
            className="max-w-2xl"
          >
            <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-blue-50 dark:bg-blue-500/10 text-blue-600 dark:text-blue-400 text-sm font-medium mb-6">
              <span className="relative flex h-2 w-2">
                <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-blue-400 opacity-75" />
                <span className="relative inline-flex rounded-full h-2 w-2 bg-blue-500" />
              </span>
              Sistem POS & Akuntansi #1 di Indonesia
            </div>

            <h1 className="text-4xl sm:text-5xl lg:text-6xl font-extrabold text-gray-900 dark:text-white tracking-tight mb-6 leading-[1.1]">
              Kelola Bisnis Tanpa Henti,{' '}
              <span className="text-transparent bg-clip-text bg-gradient-to-r from-[#0070F3] to-cyan-500">
                Akuntansi Tanpa Pusing.
              </span>
            </h1>

            <p className="text-lg sm:text-xl text-gray-600 dark:text-gray-400 mb-8 leading-relaxed max-w-xl">
              POS modern dengan mode offline otomatis dan sistem akuntansi terpadu. Jualan tetap
              jalan meski internet mati, laporan keuangan beres seketika.
            </p>

            <div className="flex flex-col sm:flex-row items-start sm:items-center gap-4">
              <Link
                href="/login"
                className="inline-flex items-center justify-center gap-2 bg-[#0070F3] text-white px-8 py-3.5 rounded-full text-base font-semibold hover:bg-blue-600 transition-all shadow-lg hover:shadow-xl hover:shadow-blue-500/20 w-full sm:w-auto"
              >
                Mulai Gratis Sekarang
                <ArrowRight size={18} />
              </Link>
              <button
                type="button"
                className="inline-flex items-center justify-center gap-2 bg-white dark:bg-white/5 border border-gray-200 dark:border-white/10 text-gray-900 dark:text-white px-8 py-3.5 rounded-full text-base font-semibold hover:bg-gray-50 dark:hover:bg-white/10 transition-all w-full sm:w-auto"
              >
                <Play size={18} className="text-[#0070F3]" fill="currentColor" />
                Lihat Demo
              </button>
            </div>

            <div className="mt-8 flex items-center gap-6 text-sm text-gray-500 dark:text-gray-400">
              <span className="flex items-center gap-1.5">
                <CheckCircle2 size={16} className="text-green-500" /> Tanpa Kartu Kredit
              </span>
              <span className="flex items-center gap-1.5">
                <CheckCircle2 size={16} className="text-green-500" /> Setup 5 Menit
              </span>
            </div>
          </motion.div>

          {/* Hero Image / Glass UI */}
          <motion.div
            initial={{ opacity: 0, x: 20, rotateY: 5 }}
            animate={{ opacity: 1, x: 0, rotateY: 0 }}
            transition={{ duration: 0.8, delay: 0.2 }}
            className="relative lg:h-[600px] flex items-center justify-center perspective-[1000px]"
          >
            {/* Abstract blobs behind the UI */}
            <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[120%] h-[120%] bg-gradient-to-br from-blue-400/20 to-purple-400/20 blur-3xl rounded-full -z-10" />

            {/* Glassmorphism POS Mockup */}
            <div className="w-full max-w-lg aspect-[4/3] bg-white/40 dark:bg-black/40 backdrop-blur-xl border border-white/40 dark:border-white/10 rounded-2xl shadow-2xl p-4 sm:p-6 transform rotate-y-[-5deg] rotate-x-[5deg]">
              {/* Header mockup */}
              <div className="flex items-center justify-between border-b border-gray-200/50 dark:border-gray-700/50 pb-4 mb-4">
                <div className="flex gap-2">
                  <div className="w-3 h-3 rounded-full bg-red-400" />
                  <div className="w-3 h-3 rounded-full bg-amber-400" />
                  <div className="w-3 h-3 rounded-full bg-green-400" />
                </div>
                <div className="text-xs font-medium text-gray-500">Kasir - Offline Ready</div>
              </div>

              {/* Grid mockup */}
              <div className="grid grid-cols-3 gap-4 h-[calc(100%-3rem)]">
                <div className="col-span-2 flex flex-col gap-3">
                  <div className="h-24 rounded-xl bg-gradient-to-r from-blue-500 to-[#0070F3] p-4 text-white flex flex-col justify-end shadow-inner">
                    <div className="text-xs opacity-80">Total Penjualan</div>
                    <div className="text-2xl font-bold">Rp 12.450.000</div>
                  </div>
                  <div className="flex-1 rounded-xl bg-white/60 dark:bg-white/5 border border-white/20 p-4">
                    <div className="h-4 w-1/3 bg-gray-200 dark:bg-gray-700 rounded mb-4" />
                    <div className="space-y-2">
                      {[1, 2, 3].map(i => (
                        <div key={i} className="flex justify-between items-center">
                          <div className="h-3 w-1/2 bg-gray-200 dark:bg-gray-700 rounded" />
                          <div className="h-3 w-1/4 bg-gray-200 dark:bg-gray-700 rounded" />
                        </div>
                      ))}
                    </div>
                  </div>
                </div>
                <div className="col-span-1 flex flex-col gap-3">
                  <div className="flex-1 rounded-xl bg-white/60 dark:bg-white/5 border border-white/20 p-4 flex flex-col justify-between">
                    <div>
                      <div className="h-3 w-2/3 bg-gray-200 dark:bg-gray-700 rounded mb-2" />
                      <div className="h-8 w-full bg-gray-200 dark:bg-gray-700 rounded" />
                    </div>
                    <div className="h-10 w-full bg-[#0070F3] rounded-lg mt-4 opacity-90" />
                  </div>
                </div>
              </div>
            </div>

            {/* Floating badge */}
            <motion.div
              initial={{ y: 20, opacity: 0 }}
              animate={{ y: 0, opacity: 1 }}
              transition={{ delay: 0.8, duration: 0.5 }}
              className="absolute -bottom-6 -left-6 sm:bottom-10 sm:-left-10 bg-white dark:bg-gray-800 p-4 rounded-xl shadow-xl border border-gray-100 dark:border-gray-700 flex items-center gap-3"
            >
              <div className="w-10 h-10 rounded-full bg-green-100 dark:bg-green-900/30 flex items-center justify-center text-green-600 dark:text-green-400">
                <CheckCircle2 size={20} />
              </div>
              <div>
                <p className="text-xs font-medium text-gray-500 dark:text-gray-400">
                  Status Sinkronisasi
                </p>
                <p className="text-sm font-bold text-gray-900 dark:text-white">
                  Online & Tersinkron
                </p>
              </div>
            </motion.div>
          </motion.div>
        </div>
      </div>
    </section>
  );
}
