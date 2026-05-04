'use client';

import { motion, AnimatePresence } from 'framer-motion';
import Link from 'next/link';
import Image from 'next/image';
import { ArrowRight, CheckCircle2 } from 'lucide-react';
import { useState, useEffect } from 'react';

const SLIDES = [
  {
    id: 1,
    light: '/order-pos-light.png',
    dark: '/order-pos-dark.png',
    tab: 'Kasir Pintar',
    title: 'Point of Sale yang Mudah',
    subtitle: 'Kelola pesanan dan transaksi toko Anda dengan cepat (Offline Ready).',
  },
  {
    id: 2,
    light: '/order-hpp-light.png',
    dark: '/order-hpp-dark.png',
    tab: 'Auto HPP',
    title: 'Kalkulator HPP Otomatis',
    subtitle: 'Hitung Harga Pokok Penjualan secara instan untuk margin akurat.',
  },
  {
    id: 3,
    light: '/image-rbac-light.png',
    dark: '/image-rbac-dark.png',
    tab: 'Keamanan (RBAC)',
    title: 'Akses Terkontrol (RBAC)',
    subtitle: 'Atur peran spesifik kasir dan admin untuk cegah kecurangan.',
  },
  {
    id: 4,
    light: '/order-cashflow-light.png',
    dark: '/order-cashflow-dark.png',
    tab: 'Arus Kas',
    title: 'Catatan Arus Kas',
    subtitle: 'Lacak setiap uang masuk dan keluar operasional dengan transparan.',
  },
];

const AUTOPLAY_INTERVAL = 6000;

export default function Hero() {
  const [currentSlide, setCurrentSlide] = useState(0);
  const [isPaused, setIsPaused] = useState(false);

  useEffect(() => {
    if (isPaused) return;
    const timer = setInterval(() => {
      setCurrentSlide(prev => (prev + 1) % SLIDES.length);
    }, AUTOPLAY_INTERVAL);
    return () => clearInterval(timer);
  }, [isPaused]);

  return (
    <section className="relative pt-32 pb-20 lg:pt-48 lg:pb-32 bg-gray-50 dark:bg-[#050505] overflow-hidden text-gray-900 dark:text-white transition-colors duration-500">
      {/* Light Mode Grid Pattern */}
      <div
        className="absolute inset-0 z-0 opacity-[0.04] dark:hidden"
        style={{
          backgroundImage: `url("data:image/svg+xml,%3Csvg width='60' height='60' viewBox='0 0 60 60' xmlns='http://www.w3.org/2000/svg'%3E%3Cg fill='none' fill-rule='evenodd'%3E%3Cg fill='%23000000' fill-opacity='1'%3E%3Cpath d='M36 34v-4h-2v4h-4v2h4v4h2v-4h4v-2h-4zm0-30V0h-2v4h-4v2h4v4h2V6h4V4h-4zM6 34v-4H4v4H0v2h4v4h2v-4h4v-2H6zM6 4V0H4v4H0v2h4v4h2V6h4V4H6z'/%3E%3C/g%3E%3C/g%3E%3C/svg%3E")`,
        }}
      />
      {/* Dark Mode Grid Pattern */}
      <div
        className="absolute inset-0 z-0 opacity-[0.05] hidden dark:block"
        style={{
          backgroundImage: `url("data:image/svg+xml,%3Csvg width='60' height='60' viewBox='0 0 60 60' xmlns='http://www.w3.org/2000/svg'%3E%3Cg fill='none' fill-rule='evenodd'%3E%3Cg fill='%23ffffff' fill-opacity='1'%3E%3Cpath d='M36 34v-4h-2v4h-4v2h4v4h2v-4h4v-2h-4zm0-30V0h-2v4h-4v2h4v4h2V6h4V4h-4zM6 34v-4H4v4H0v2h4v4h2v-4h4v-2H6zM6 4V0H4v4H0v2h4v4h2V6h4V4H6z'/%3E%3C/g%3E%3C/g%3E%3C/svg%3E")`,
        }}
      />

      {/* Subtle Glows */}
      <div className="absolute top-0 left-1/2 -translate-x-1/2 w-[800px] h-[400px] bg-[#0070F3]/10 dark:bg-[#0070F3]/20 blur-[120px] rounded-full pointer-events-none" />

      <div className="max-w-7xl mx-auto px-6 relative z-10">
        <div className="flex flex-col items-center text-center">
          {/* Badge */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6 }}
          >
            <div className="inline-flex items-center gap-2 px-3 py-1.5 rounded-full bg-white dark:bg-white/5 border border-gray-200 dark:border-white/10 text-gray-600 dark:text-gray-300 text-sm font-medium mb-8 shadow-sm backdrop-blur-sm">
              <span className="relative flex h-2 w-2">
                <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-[#0070F3] opacity-75" />
                <span className="relative inline-flex rounded-full h-2 w-2 bg-[#0070F3]" />
              </span>
              Sistem POS & Akuntansi #1 di Indonesia
            </div>
          </motion.div>

          {/* Headline */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6, delay: 0.1 }}
            className="max-w-4xl flex flex-col items-center"
          >
            <h1 className="text-5xl sm:text-6xl lg:text-7xl font-extrabold tracking-tight mb-8 leading-[1.05] text-gray-900 dark:text-white">
              Kelola Bisnis Tanpa Henti,
              <br />
              <span className="text-[#0070F3]">Akuntansi Tanpa Pusing.</span>
            </h1>

            <p className="text-lg sm:text-xl text-gray-600 dark:text-gray-400 mb-10 leading-relaxed max-w-2xl">
              Tinggalkan cara lama. POS modern kami dirancang khusus agar jualan Anda tetap jalan
              meski internet terputus, dengan laporan keuangan yang beres seketika.
            </p>

            <div className="flex flex-col sm:flex-row items-center gap-4 w-full sm:w-auto">
              <Link
                href="/login"
                className="inline-flex items-center justify-center gap-2 bg-[#0070F3] text-white px-8 py-4 rounded-full text-base font-semibold hover:bg-blue-500 transition-all shadow-[0_0_20px_rgba(0,112,243,0.3)] hover:shadow-[0_0_40px_rgba(0,112,243,0.5)] hover:-translate-y-0.5 w-full sm:w-auto"
              >
                Mulai Gratis Sekarang
                <ArrowRight size={18} />
              </Link>
            </div>

            <div className="mt-10 flex items-center justify-center gap-8 text-sm text-gray-500 dark:text-gray-400 font-medium">
              <span className="flex items-center gap-2">
                <CheckCircle2 size={18} className="text-[#0070F3]" /> Tanpa Kartu Kredit
              </span>
              <span className="flex items-center gap-2">
                <CheckCircle2 size={18} className="text-[#0070F3]" /> Setup 5 Menit
              </span>
            </div>
          </motion.div>

          {/* Interactive Feature Slider */}
          <motion.div
            initial={{ opacity: 0, y: 40 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.8, delay: 0.3 }}
            className="w-full mt-24"
            onMouseEnter={() => setIsPaused(true)}
            onMouseLeave={() => setIsPaused(false)}
          >
            {/* Interactive Tabs */}
            <div className="flex flex-wrap justify-center gap-2 mb-8">
              {SLIDES.map((slide, idx) => (
                <button
                  key={slide.id}
                  onClick={() => setCurrentSlide(idx)}
                  className={`relative px-5 py-3 rounded-xl transition-all duration-300 flex flex-col text-left overflow-hidden min-w-[200px] ${
                    currentSlide === idx
                      ? 'bg-white dark:bg-white/10 border border-gray-200 dark:border-white/20 shadow-sm'
                      : 'bg-transparent border border-transparent hover:bg-gray-100 dark:hover:bg-white/5 text-gray-500 dark:text-gray-500'
                  }`}
                >
                  <span
                    className={`text-sm font-bold mb-1 ${currentSlide === idx ? 'text-gray-900 dark:text-white' : ''}`}
                  >
                    {slide.tab}
                  </span>
                  <span
                    className={`text-xs ${currentSlide === idx ? 'text-gray-500 dark:text-gray-300' : 'opacity-0 h-0'}`}
                  >
                    {slide.subtitle.substring(0, 45)}...
                  </span>

                  {/* Animated Progress Bar */}
                  {currentSlide === idx && !isPaused && (
                    <motion.div
                      className="absolute bottom-0 left-0 h-[2px] bg-[#0070F3]"
                      initial={{ width: 0 }}
                      animate={{ width: '100%' }}
                      transition={{ duration: AUTOPLAY_INTERVAL / 1000, ease: 'linear' }}
                      key={`progress-${idx}`}
                    />
                  )}
                  {/* Static bar if paused */}
                  {currentSlide === idx && isPaused && (
                    <div className="absolute bottom-0 left-0 h-[2px] bg-[#0070F3] w-full" />
                  )}
                </button>
              ))}
            </div>

            {/* Browser Window Mockup Container */}
            <div className="max-w-5xl mx-auto">
              <div className="relative rounded-2xl overflow-hidden border border-gray-200 dark:border-gray-800 bg-white dark:bg-[#111] shadow-2xl dark:shadow-[0_20px_50px_rgba(0,0,0,0.5)]">
                {/* macOS Browser Header */}
                <div className="flex items-center gap-2 px-4 py-3 border-b border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-[#161616]">
                  <div className="flex gap-1.5">
                    <div className="w-3 h-3 rounded-full bg-[#ff5f56]" />
                    <div className="w-3 h-3 rounded-full bg-[#ffbd2e]" />
                    <div className="w-3 h-3 rounded-full bg-[#27c93f]" />
                  </div>
                  <div className="mx-auto bg-white dark:bg-black border border-gray-200 dark:border-gray-800 rounded-md px-4 py-1 text-xs text-gray-500 font-mono w-64 text-center truncate">
                    app.moedah.com/dashboard
                  </div>
                  <div className="w-12" /> {/* Spacer to balance dots */}
                </div>

                {/* Slides Container */}
                <div className="relative aspect-[16/10] bg-gray-100 dark:bg-black">
                  <AnimatePresence mode="wait">
                    <motion.div
                      key={currentSlide}
                      initial={{ opacity: 0, scale: 0.98 }}
                      animate={{ opacity: 1, scale: 1 }}
                      exit={{ opacity: 0 }}
                      transition={{ duration: 0.5 }}
                      className="absolute inset-0"
                    >
                      {/* Using CSS to toggle dark/light images instantly */}
                      <Image
                        src={SLIDES[currentSlide].light}
                        alt={SLIDES[currentSlide].title}
                        fill
                        className="object-cover object-top dark:hidden"
                        priority={true}
                      />
                      <Image
                        src={SLIDES[currentSlide].dark}
                        alt={SLIDES[currentSlide].title}
                        fill
                        className="object-cover object-top hidden dark:block"
                        priority={true}
                      />
                    </motion.div>
                  </AnimatePresence>
                </div>
              </div>
            </div>
          </motion.div>
        </div>
      </div>
    </section>
  );
}
