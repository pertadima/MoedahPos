'use client';

import { motion } from 'framer-motion';
import Link from 'next/link';
import { ArrowRight } from 'lucide-react';

export default function Hero() {
  return (
    <section className="relative min-h-[85vh] flex flex-col justify-center bg-[#f8f9fb] dark:bg-[#09090b] overflow-hidden transition-colors duration-500">
      {/* Structural pattern - subtle and professional */}
      <div
        className="absolute inset-0 opacity-[0.03] dark:opacity-[0.05] pointer-events-none"
        style={{
          backgroundImage: `radial-gradient(#4f6ef7 1px, transparent 1px)`,
          backgroundSize: '40px 40px',
        }}
      />

      <div className="relative z-10 max-w-7xl mx-auto px-6 pt-20 pb-20 w-full">
        <div className="max-w-4xl">
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5 }}
          >
            {/* Professional Eyebrow */}
            <div className="inline-flex items-center gap-2 mb-6 px-3 py-1 rounded-md bg-[#4f6ef7]/10 border border-[#4f6ef7]/20">
              <span className="text-[11px] font-bold tracking-wider uppercase text-[#4f6ef7]">
                Reliable POS Solution
              </span>
            </div>

            {/* Robust Headline */}
            <h1
              className="
                text-5xl sm:text-6xl lg:text-[5.5rem]
                leading-[1.1] tracking-tight
                text-gray-900 dark:text-white
                font-extrabold mb-8
              "
            >
              Kelola Bisnis Tanpa Henti,
              <br />
              <span className="text-[#4f6ef7]">Akuntansi Tanpa Pusing.</span>
            </h1>

            <p className="text-lg sm:text-xl text-gray-600 dark:text-gray-400 leading-relaxed max-w-2xl mb-12">
              Satu-satunya sistem POS yang menggabungkan kemudahan kasir dengan ketajaman laporan
              akuntansi profesional. Tetap jualan meski tanpa koneksi internet.
            </p>

            {/* High-Contrast Action Area */}
            <div className="flex items-center gap-6 flex-wrap">
              <Link
                href="/login"
                className="
                  inline-flex items-center gap-2
                  h-14 px-8 rounded-xl text-base font-bold text-white
                  bg-[#4f6ef7] hover:bg-[#3d56d4]
                  transition-all duration-200
                  shadow-lg shadow-[#4f6ef7]/25
                  hover:-translate-y-0.5
                "
              >
                Mulai Gratis Sekarang <ArrowRight size={18} />
              </Link>

              <div className="flex flex-col">
                <span className="text-sm font-semibold text-gray-900 dark:text-white">
                  600+ Bisnis Terdaftar
                </span>
                <span className="text-xs text-gray-500">
                  Bergabung dengan komunitas Moedah hari ini.
                </span>
              </div>
            </div>
          </motion.div>
        </div>
      </div>
    </section>
  );
}
