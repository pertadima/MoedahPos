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
  LayoutGrid,
  ChefHat,
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
    light: '/image-table-light.webp',
    icon: LayoutGrid,
    tab: 'Meja',
    title: 'Manajemen Meja',
    desc: 'Kelola meja restoran, reservasi, dan pesanan makan di tempat dengan mudah.',
  },
  {
    id: 4,
    light: '/image-kitchen-light.webp',
    icon: ChefHat,
    tab: 'Dapur',
    title: 'Kitchen Display',
    desc: 'Tampilkan pesanan di dapur secara real-time untuk layanan lebih cepat.',
  },
  {
    id: 5,
    light: '/image-rbac-light.webp',
    icon: ShieldCheck,
    tab: 'RBAC',
    title: 'Kontrol Akses',
    desc: 'Peran dan izin per modul untuk setiap anggota tim. Cegah kecurangan sejak awal.',
  },
  {
    id: 6,
    light: '/order-purchase-light.webp',
    icon: Package,
    tab: 'Purchase',
    title: 'Purchase Order',
    desc: 'Kelola pengadaan stok dan supplier. Termin pembayaran, FIFO batch, dan lebih banyak lagi.',
  },
  {
    id: 7,
    light: '/order-report-light.webp',
    icon: FileText,
    tab: 'Laporan',
    title: 'Laporan Bisnis',
    desc: 'P&L, arus kas, dan valuasi stok real-time. Export PDF & CSV dalam satu klik.',
  },
  {
    id: 8,
    light: '/order-cashflow-light.webp',
    icon: Banknote,
    tab: 'Arus Kas',
    title: 'Arus Kas',
    desc: 'Visibilitas penuh atas setiap pergerakan keuangan bisnis Anda per hari.',
  },
];

const DURATION = 5000;

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
      className="py-24 bg-white"
      onMouseEnter={() => setPaused(true)}
      onMouseLeave={() => setPaused(false)}
    >
      <div className="max-w-7xl mx-auto px-6">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6 }}
          className="mb-12 text-center"
        >
          <p className="text-xs font-bold tracking-[0.2em] uppercase text-[#0884F6] mb-4">
            Powerful Modules
          </p>
          <h2 className="text-3xl sm:text-5xl font-extrabold text-gray-900 tracking-tight">
            Satu platform,
            <br />
            <span className="text-gray-500 font-normal">semua yang Anda butuhkan.</span>
          </h2>
        </motion.div>

        <motion.div
          initial={{ opacity: 0, y: 30 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6, delay: 0.15 }}
          className="rounded-2xl overflow-hidden bg-white shadow-2xl ring-1 ring-gray-200"
        >
          <div className="flex items-center gap-3 px-5 py-4 border-b border-gray-100">
            <div className="flex gap-1.5">
              <div className="w-3 h-3 rounded-full bg-[#ff5f56]" />
              <div className="w-3 h-3 rounded-full bg-[#ffbd2e]" />
              <div className="w-3 h-3 rounded-full bg-[#27c93f]" />
            </div>
            <div className="flex-1 flex justify-center">
              <div className="h-7 w-64 flex items-center justify-center bg-gray-50 border border-gray-200 rounded-md px-3">
                <AnimatePresence mode="wait">
                  <motion.span
                    key={active}
                    initial={{ opacity: 0, y: 5 }}
                    animate={{ opacity: 1, y: 0 }}
                    exit={{ opacity: 0, y: -5 }}
                    className="text-[11px] text-gray-400 font-medium"
                  >
                    moedah.com/app/{SLIDES[active].tab.toLowerCase()}
                  </motion.span>
                </AnimatePresence>
              </div>
            </div>
          </div>

          <div className="relative aspect-[16/9] bg-gray-50 p-4 sm:p-8 text-center">
            <AnimatePresence mode="wait">
              <motion.div
                key={active}
                initial={{ opacity: 0, scale: 0.98 }}
                animate={{ opacity: 1, scale: 1 }}
                exit={{ opacity: 0, scale: 1.02 }}
                transition={{ duration: 0.35 }}
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
        </motion.div>

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.5, delay: 0.25 }}
          className="mt-8 grid grid-cols-2 md:grid-cols-4 lg:grid-cols-8 gap-3"
        >
          {SLIDES.map((s, i) => {
            const Icon = s.icon;
            const isActive = active === i;
            return (
              <motion.button
                key={s.id}
                onClick={() => setActive(i)}
                initial={{ opacity: 0, y: 10 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ duration: 0.4, delay: 0.3 + i * 0.05 }}
                whileHover={{ scale: 1.02 }}
                whileTap={{ scale: 0.98 }}
                className={`relative flex flex-col items-start p-4 rounded-xl border transition-all duration-300 ${
                  isActive
                    ? 'bg-white border-[#0884F6]/50 shadow-lg shadow-[#0884F6]/20 scale-[1.02] z-10'
                    : 'bg-white border-gray-200 hover:border-gray-300 hover:shadow-md'
                }`}
              >
                <div
                  className={`w-10 h-10 rounded-lg flex items-center justify-center mb-4 transition-all ${
                    isActive
                      ? 'bg-[#0884F6] text-white shadow-md shadow-[#0884F6]/30'
                      : 'bg-gray-100 text-gray-400'
                  }`}
                >
                  <Icon size={18} />
                </div>
                <p
                  className={`text-xs font-bold uppercase tracking-wider leading-none mb-1 ${isActive ? 'text-[#0884F6]' : 'text-gray-400'}`}
                >
                  {s.tab}
                </p>
                <p
                  className={`text-[13px] font-semibold leading-tight ${isActive ? 'text-gray-900' : 'text-gray-500'}`}
                >
                  {s.title}
                </p>
                {isActive && !paused && (
                  <motion.div
                    className="absolute bottom-0 left-0 right-0 h-1 bg-[#FFA724]"
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 1 }}
                  >
                    <motion.div
                      className="h-full bg-[#FFA724]"
                      initial={{ width: '0%' }}
                      animate={{ width: '100%' }}
                      transition={{ duration: DURATION / 1000, ease: 'linear' }}
                    />
                  </motion.div>
                )}
              </motion.button>
            );
          })}
        </motion.div>

        <motion.div
          initial={{ opacity: 0, y: 15 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.5, delay: 0.35 }}
          className="mt-8 p-6 rounded-2xl bg-gradient-to-r from-[#0884F6]/5 to-[#FFA724]/5 border border-[#0884F6]/20 flex items-center justify-between gap-8"
        >
          <div className="flex-1">
            <AnimatePresence mode="wait">
              <motion.div
                key={active}
                initial={{ opacity: 0, x: -10 }}
                animate={{ opacity: 1, x: 0 }}
                exit={{ opacity: 0, x: 10 }}
                transition={{ duration: 0.25 }}
                className="flex items-center gap-4"
              >
                <div className="w-12 h-12 rounded-full bg-gradient-to-br from-[#0884F6] to-[#FFA724] text-white flex items-center justify-center shadow-lg">
                  <ActiveIcon size={22} />
                </div>
                <div>
                  <h4 className="font-bold text-gray-900 text-lg">{SLIDES[active].title}</h4>
                  <p className="text-gray-600 text-sm">{SLIDES[active].desc}</p>
                </div>
              </motion.div>
            </AnimatePresence>
          </div>
          <Link
            href="/login"
            className="hidden sm:flex items-center gap-2 text-sm font-bold text-[#0884F6] hover:text-[#FFA724] transition-colors whitespace-nowrap"
          >
            Coba Modul {SLIDES[active].tab} <ArrowRight size={16} />
          </Link>
        </motion.div>
      </div>
    </section>
  );
}
