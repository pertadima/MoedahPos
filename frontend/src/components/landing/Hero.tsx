'use client';

import { motion } from 'framer-motion';
import Link from 'next/link';
import Image from 'next/image';
import { ArrowRight, Sparkles } from 'lucide-react';
import { useMemo } from 'react';

// Static particle data to avoid Math.random() in render
const PARTICLES = [
  { x: '10%', y: '20%', d: 6, delay: 0 },
  { x: '80%', y: '15%', d: 8, delay: 1 },
  { x: '45%', y: '50%', d: 5, delay: 2 },
  { x: '15%', y: '75%', d: 7, delay: 0.5 },
  { x: '85%', y: '80%', d: 9, delay: 1.5 },
  { x: '55%', y: '10%', d: 6, delay: 3 },
];

export default function Hero() {
  const particles = useMemo(() => PARTICLES, []);

  return (
    <section className="relative min-h-[92vh] flex flex-col justify-center bg-[#f8f9fb] dark:bg-[#09090b] overflow-hidden transition-colors duration-700">
      {/* ── BACKGROUND ARTILLERY ─────────────────────────────────────────── */}

      {/* 1. Static Gradient Mesh Orbs (Removed animation for performance) */}
      <div className="absolute inset-0 z-0 pointer-events-none overflow-hidden">
        <div className="absolute -top-[10%] -right-[5%] w-[60%] h-[60%] bg-[#0884F6]/[0.05] dark:bg-[#0884F6]/[0.1] blur-[80px] rounded-full" />
        <div className="absolute -bottom-[20%] -left-[10%] w-[70%] h-[70%] bg-[#FFA724]/[0.03] dark:bg-[#FFA724]/[0.05] blur-[100px] rounded-full" />
      </div>

      {/* 2. Geometric Grid with Masking */}
      <div
        className="absolute inset-0 z-0 opacity-[0.4] dark:opacity-[0.6] pointer-events-none [mask-image:radial-gradient(ellipse_at_center,black,transparent_75%)]"
        style={{
          backgroundImage: `
            linear-gradient(to right, #0884F610 1px, transparent 1px),
            linear-gradient(to bottom, #0884F610 1px, transparent 1px)
          `,
          backgroundSize: '64px 64px',
        }}
      />

      {/* 3. Floating "Dust" particles */}
      <div className="absolute inset-0 pointer-events-none">
        {particles.map((p, i) => (
          <motion.div
            key={i}
            className="absolute w-1 h-1 bg-[#0884F6]/40 rounded-full"
            style={{ left: p.x, top: p.y }}
            animate={{
              y: [0, -40, 0],
              opacity: [0, 1, 0],
            }}
            transition={{
              duration: p.d,
              repeat: Infinity,
              delay: p.delay,
            }}
          />
        ))}
      </div>

      {/* ── CONTENT ─────────────────────────────────────────────────────── */}

      <div className="relative z-10 max-w-7xl mx-auto px-6 pt-20 pb-20 w-full flex flex-col lg:flex-row items-center lg:items-stretch gap-12">
        <div className="flex-1 max-w-2xl text-left lg:text-left py-8 flex flex-col justify-center shrink-0">
          <motion.div
            initial={{ opacity: 0, x: -30 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ duration: 0.8, ease: [0.16, 1, 0.3, 1] }}
          >
            {/* Professional Eyebrow with Sparkle */}
            <div className="inline-flex items-center gap-2 mb-8 px-4 py-1.5 rounded-full bg-white dark:bg-white/[0.05] border border-gray-200 dark:border-white/[0.1] shadow-sm backdrop-blur-md">
              <Sparkles size={14} className="text-[#FFA724] animate-pulse" />
              <span className="text-[11px] font-bold tracking-[0.15em] uppercase text-gray-600 dark:text-gray-300">
                Next-Gen POS for Modern Business
              </span>
            </div>

            {/* Robust Kinetic Headline */}
            <h1
              className="
                text-5xl sm:text-6xl lg:text-[4.5rem]
                leading-[0.95] tracking-tight
                text-gray-900 dark:text-white
                font-black mb-8
              "
            >
              Kelola Bisnis
              <br />
              <span className="relative inline-block text-transparent bg-clip-text bg-gradient-to-r from-[#0884F6] via-[#FFA724] to-[#0884F6] bg-[length:200%_auto] animate-gradient-flow">
                Tanpa Henti.
              </span>
            </h1>

            <motion.p
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.4 }}
              className="text-lg text-gray-500 dark:text-slate-400 leading-relaxed mb-10 font-medium max-w-lg"
            >
              Tingkatkan efisiensi dengan sistem POS offline-first yang secara otomatis
              menyinkronkan akuntansi, inventaris, dan loyalitas pelanggan Anda ke cloud.
            </motion.p>

            {/* Vibrant Action Area */}
            <div className="flex items-center gap-6 flex-wrap">
              <Link
                href="/login"
                className="
                  group relative inline-flex items-center gap-3
                  h-14 px-10 rounded-2xl text-base font-bold text-white
                  bg-[#0884F6] hover:bg-[#0672d6]
                  transition-all duration-300
                  shadow-[0_20px_40px_-10px_rgba(8, 132, 246, 0.4)]
                  hover:shadow-[0_25px_50px_-12px_rgba(8, 132, 246, 0.5)]
                  hover:-translate-y-1
                  overflow-hidden
                "
              >
                <div className="absolute inset-0 bg-gradient-to-r from-transparent via-white/10 to-transparent -translate-x-full group-hover:translate-x-full transition-transform duration-1000" />
                Mulai Sekarang
                <ArrowRight size={20} className="group-hover:translate-x-1 transition-transform" />
              </Link>

              <div className="flex items-center gap-4">
                <div className="flex -space-x-3">
                  {[1, 2, 3].map(i => (
                    <div
                      key={i}
                      className="w-10 h-10 rounded-full border-2 border-white dark:border-[#09090b] bg-gray-200 dark:bg-gray-800"
                    />
                  ))}
                </div>
                <div className="flex flex-col">
                  <span className="text-xs font-bold text-gray-900 dark:text-white leading-none mb-1">
                    Terpercaya
                  </span>
                  <span className="text-[10px] text-gray-500 uppercase tracking-widest font-semibold opacity-80">
                    600+ Bisnis
                  </span>
                </div>
              </div>
            </div>
          </motion.div>
        </div>

        {/* ── MOCKUP AREA ── */}
        <div className="flex-1 w-full flex items-center justify-center lg:justify-start pt-12 lg:pt-0">
          <div className="grid grid-cols-3 gap-4 w-full max-w-[800px]">
            {/* Mockup 1 */}
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.6, delay: 0.3 }}
            >
              <div className="aspect-[1/2] relative">
                <Image
                  src="/mockuplight-1.png"
                  alt="POS 1"
                  fill
                  className="object-contain dark:hidden"
                  sizes="(max-width: 1024px) 33vw, 300px"
                />
                <Image
                  src="/mockupdark-1.png"
                  alt="POS 1"
                  fill
                  className="object-contain hidden dark:block"
                  sizes="(max-width: 1024px) 33vw, 300px"
                />
              </div>
            </motion.div>

            {/* Mockup 2 */}
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.6, delay: 0.4 }}
            >
              <div className="aspect-[1/2] relative">
                <Image
                  src="/mockuplight-2.png"
                  alt="POS 2"
                  fill
                  className="object-contain dark:hidden"
                  sizes="(max-width: 1024px) 33vw, 300px"
                />
                <Image
                  src="/mockupdark-2.png"
                  alt="POS 2"
                  fill
                  className="object-contain hidden dark:block"
                  sizes="(max-width: 1024px) 33vw, 300px"
                />
              </div>
            </motion.div>

            {/* Mockup 3 */}
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.6, delay: 0.5 }}
            >
              <div className="aspect-[1/2] relative">
                <Image
                  src="/mockuplight-3.png"
                  alt="POS 3"
                  fill
                  className="object-contain dark:hidden"
                  sizes="(max-width: 1024px) 33vw, 300px"
                />
                <Image
                  src="/mockupdark-3.png"
                  alt="POS 3"
                  fill
                  className="object-contain hidden dark:block"
                  sizes="(max-width: 1024px) 33vw, 300px"
                />
              </div>
            </motion.div>
          </div>

          {/* Ambient Background Glow */}
          <div className="absolute right-0 top-1/2 -translate-y-1/2 w-[400px] h-[400px] bg-[#0884F6]/10 dark:bg-[#0884F6]/20 blur-[120px] -z-10" />
        </div>
      </div>

      <style
        dangerouslySetInnerHTML={{
          __html: `
        @keyframes gradient-flow {
          0% { background-position: 0% center; }
          100% { background-position: 200% center; }
        }
        .animate-gradient-flow {
          animation: gradient-flow 4s linear infinite;
        }
      `,
        }}
      />
    </section>
  );
}
