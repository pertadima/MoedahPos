'use client';

import { motion, AnimatePresence } from 'framer-motion';
import Image from 'next/image';
import {
  ShoppingCart,
  BarChart3,
  ShieldCheck,
  Package,
  FileText,
  Banknote,
  ArrowRight,
} from 'lucide-react';
import { useState, useEffect } from 'react';
import Link from 'next/link';

const SLIDES = [
  {
    id: 1,
    light: '/order-pos-light.webp',
    icon: ShoppingCart,
    tab: 'Kasir',
    title: 'Point of Sale',
    desc: 'Transaksi cepat dan akurat. Tetap berjalan sempurna saat koneksi internet terputus.',
  },
  {
    id: 2,
    light: '/order-hpp-light.webp',
    icon: BarChart3,
    tab: 'HPP',
    title: 'Kalkulator HPP',
    desc: 'Harga Pokok Penjualan dikalkulasi otomatis setiap transaksi. Margin selalu akurat.',
  },
  {
    id: 3,
    light: '/image-rbac-light.webp',
    icon: ShieldCheck,
    tab: 'RBAC',
    title: 'Kontrol Akses',
    desc: 'Peran dan izin per modul untuk setiap anggota tim. Cegah kecurangan sejak awal.',
  },
  {
    id: 4,
    light: '/order-purchase-light.webp',
    icon: Package,
    tab: 'Purchase',
    title: 'Purchase Order',
    desc: 'Kelola pengadaan stok dan supplier. Termin pembayaran, FIFO batch, dan lebih banyak lagi.',
  },
  {
    id: 5,
    light: '/order-report-light.webp',
    icon: FileText,
    tab: 'Laporan',
    title: 'Laporan Bisnis',
    desc: 'P&L, arus kas, dan valuasi stok real-time. Export PDF & CSV dalam satu klik.',
  },
  {
    id: 6,
    light: '/order-cashflow-light.webp',
    icon: Banknote,
    tab: 'Arus Kas',
    title: 'Arus Kas',
    desc: 'Visibilitas penuh atas setiap pergerakan keuangan bisnis Anda per hari.',
  },
];

const DURATION = 6000;

export default function FeatureSection() {
  const [active, setActive] = useState(0);
  const [paused, setPaused] = useState(false);

  useEffect(() => {
    if (paused) return;
    const t = setInterval(() => setActive(p => (p + 1) % SLIDES.length), DURATION);
    return () => clearInterval(t);
  }, [paused]);

  const ActiveIcon = SLIDES[active].icon;

  return (
    <section
      className="py-24 bg-[#f1f3f7] border-t border-gray-200"
      onMouseEnter={() => setPaused(true)}
      onMouseLeave={() => setPaused(false)}
    >
      <div className="max-w-7xl mx-auto px-6">
        {/* Section header */}
        <div className="mb-16 text-center">
          <p className="text-xs font-bold tracking-[0.2em] uppercase text-[#0884F6] mb-4">
            Powerful Modules
          </p>
          <h2 className="text-3xl sm:text-5xl font-extrabold text-gray-900 tracking-tight">
            Satu platform,
            <br />
            <span className="text-gray-500 font-normal underline decoration-[#0884F6]/30">
              semua yang Anda butuhkan.
            </span>
          </h2>
        </div>

        {/* Browser mockup */}
        <div className="rounded-2xl overflow-hidden border border-gray-200 bg-white shadow-2xl">
          <div className="flex items-center gap-3 px-5 py-4 border-b border-gray-100 bg-gray-50">
            <div className="flex gap-1.5">
              <div className="w-3 h-3 rounded-full bg-[#ff5f56]" />
              <div className="w-3 h-3 rounded-full bg-[#ffbd2e]" />
              <div className="w-3 h-3 rounded-full bg-[#27c93f]" />
            </div>
            <div className="flex-1 flex justify-center">
              <div className="h-7 w-64 flex items-center justify-center bg-white border border-gray-200 rounded-md px-3">
                <AnimatePresence mode="wait">
                  <motion.span
                    key={active}
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 1 }}
                    exit={{ opacity: 0 }}
                    className="text-[11px] text-gray-400 font-medium tracking-tight"
                  >
                    moedah.com/app/{SLIDES[active].tab.toLowerCase()}
                  </motion.span>
                </AnimatePresence>
              </div>
            </div>
          </div>

          <div className="relative aspect-[16/9] bg-gray-100 p-4 sm:p-8 text-center">
            <AnimatePresence mode="wait">
              <motion.div
                key={active}
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                exit={{ opacity: 0 }}
                transition={{ duration: 0.4 }}
                className="absolute inset-0 p-4 sm:p-8"
              >
                <Image
                  src={SLIDES[active].light}
                  alt={SLIDES[active].title}
                  fill
                  className="object-contain object-center"
                  priority
                />
              </motion.div>
            </AnimatePresence>
          </div>
        </div>

        {/* Tabs */}
        <div className="mt-8 grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-3">
          {SLIDES.map((s, i) => {
            const Icon = s.icon;
            const isActive = active === i;
            return (
              <button
                key={s.id}
                onClick={() => setActive(i)}
                className={`
                  relative flex flex-col items-start p-4 rounded-xl
                  border transition-all duration-300 group
                  ${
                    isActive
                      ? 'bg-white border-[#0884F6]/50 shadow-lg scale-[1.02] z-10'
                      : 'bg-white/40 border-gray-200 hover:border-gray-300'
                  }
                `}
              >
                <div
                  className={`w-10 h-10 rounded-lg flex items-center justify-center mb-4 transition-colors
                    ${
                      isActive
                        ? 'bg-[#0884F6] text-white shadow-md shadow-[#0884F6]/30'
                        : 'bg-gray-100 text-gray-400 group-hover:text-gray-600'
                    }`}
                >
                  <Icon size={18} />
                </div>
                <p
                  className={`text-xs font-bold uppercase tracking-wider leading-none mb-1
                  ${isActive ? 'text-[#0884F6]' : 'text-gray-400'}`}
                >
                  {s.tab}
                </p>
                <p
                  className={`text-[13px] font-semibold leading-tight
                  ${isActive ? 'text-gray-900' : 'text-gray-500'}`}
                >
                  {s.title}
                </p>
                {isActive && !paused && (
                  <motion.div
                    className="absolute bottom-0 left-0 h-1 bg-[#0884F6]"
                    initial={{ width: 0 }}
                    animate={{ width: '100%' }}
                    transition={{ duration: DURATION / 1000, ease: 'linear' }}
                    key={`bar-${i}`}
                  />
                )}
              </button>
            );
          })}
        </div>

        {/* Feature detail */}
        <div className="mt-8 p-6 rounded-2xl bg-[#0884F6]/[0.03] border border-[#0884F6]/10 flex items-center justify-between gap-8">
          <div className="flex-1">
            <AnimatePresence mode="wait">
              <motion.div
                key={active}
                initial={{ opacity: 0, x: -10 }}
                animate={{ opacity: 1, x: 0 }}
                exit={{ opacity: 0, x: 10 }}
                className="flex items-center gap-4"
              >
                <div className="w-10 h-10 rounded-full bg-[#0884F6] text-white flex items-center justify-center shadow-lg shadow-[#0884F6]/20">
                  <ActiveIcon size={20} />
                </div>
                <div>
                  <h4 className="font-bold text-gray-900">{SLIDES[active].title}</h4>
                  <p className="text-sm text-gray-600">{SLIDES[active].desc}</p>
                </div>
              </motion.div>
            </AnimatePresence>
          </div>
          <Link
            href="/login"
            className="hidden sm:flex items-center gap-2 text-sm font-bold text-[#0884F6] hover:underline whitespace-nowrap"
          >
            Coba Modul {SLIDES[active].tab} <ArrowRight size={16} />
          </Link>
        </div>
      </div>
    </section>
  );
}
