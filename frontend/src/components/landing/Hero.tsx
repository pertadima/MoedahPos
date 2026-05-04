'use client';

import { motion } from 'framer-motion';
import Link from 'next/link';
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

      {/* 1. Animated Gradient Mesh Orbs */}
      <div className="absolute inset-0 z-0 pointer-events-none overflow-hidden">
        <motion.div
          animate={{
            x: [0, 40, 0],
            y: [0, -30, 0],
            scale: [1, 1.1, 1],
          }}
          transition={{ duration: 15, repeat: Infinity, ease: 'easeInOut' }}
          className="absolute -top-[10%] -right-[5%] w-[60%] h-[60%] bg-[#4f6ef7]/[0.08] dark:bg-[#4f6ef7]/[0.15] blur-[120px] rounded-full"
        />
        <motion.div
          animate={{
            x: [0, -50, 0],
            y: [0, 40, 0],
            scale: [1, 1.2, 1],
          }}
          transition={{ duration: 18, repeat: Infinity, ease: 'easeInOut', delay: 2 }}
          className="absolute -bottom-[20%] -left-[10%] w-[70%] h-[70%] bg-cyan-400/[0.05] dark:bg-cyan-500/[0.08] blur-[140px] rounded-full"
        />
      </div>

      {/* 2. Geometric Grid with Masking */}
      <div
        className="absolute inset-0 z-0 opacity-[0.4] dark:opacity-[0.6] pointer-events-none [mask-image:radial-gradient(ellipse_at_center,black,transparent_75%)]"
        style={{
          backgroundImage: `
            linear-gradient(to right, #4f6ef710 1px, transparent 1px),
            linear-gradient(to bottom, #4f6ef710 1px, transparent 1px)
          `,
          backgroundSize: '64px 64px',
        }}
      />

      {/* 3. Floating "Dust" particles */}
      <div className="absolute inset-0 pointer-events-none">
        {particles.map((p, i) => (
          <motion.div
            key={i}
            className="absolute w-1 h-1 bg-[#4f6ef7]/40 rounded-full"
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

      <div className="relative z-10 max-w-7xl mx-auto px-6 pt-20 pb-20 w-full">
        <div className="max-w-5xl">
          <motion.div
            initial={{ opacity: 0, x: -30 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ duration: 0.8, ease: [0.16, 1, 0.3, 1] }}
          >
            {/* Professional Eyebrow with Sparkle */}
            <motion.div
              initial={{ opacity: 0, scale: 0.9 }}
              animate={{ opacity: 1, scale: 1 }}
              transition={{ delay: 0.2 }}
              className="inline-flex items-center gap-2 mb-8 px-4 py-1.5 rounded-full bg-white dark:bg-white/[0.05] border border-gray-200 dark:border-white/[0.1] shadow-sm backdrop-blur-md"
            >
              <Sparkles size={14} className="text-[#4f6ef7] animate-pulse" />
              <span className="text-[11px] font-bold tracking-[0.15em] uppercase text-gray-600 dark:text-gray-300">
                Next-Gen POS for Modern Business
              </span>
            </motion.div>

            {/* Robust Kinetic Headline */}
            <h1
              className="
                text-6xl sm:text-7xl lg:text-[6.5rem]
                leading-[0.95] tracking-tight
                text-gray-900 dark:text-white
                font-black mb-10
              "
            >
              Kelola Bisnis
              <br />
              <span className="relative inline-block text-transparent bg-clip-text bg-gradient-to-r from-[#4f6ef7] via-[#7c8fff] to-[#4f6ef7] bg-[length:200%_auto] animate-gradient-flow">
                Tanpa Henti.
              </span>
            </h1>

            <motion.p
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.4 }}
              className="text-lg sm:text-xl text-gray-500 dark:text-slate-400 leading-relaxed max-w-xl mb-12 font-medium"
            >
              Tingkatkan efisiensi dengan sistem POS offline-first yang secara otomatis
              menyinkronkan akuntansi, inventaris, dan loyalitas pelanggan Anda ke cloud.
            </motion.p>

            {/* Vibrant Action Area */}
            <div className="flex items-center gap-8 flex-wrap">
              <Link
                href="/login"
                className="
                  group relative inline-flex items-center gap-3
                  h-16 px-10 rounded-2xl text-lg font-bold text-white
                  bg-[#4f6ef7] hover:bg-[#3d56d4]
                  transition-all duration-300
                  shadow-[0_20px_40px_-10px_rgba(79,110,247,0.4)]
                  hover:shadow-[0_25px_50px_-12px_rgba(79,110,247,0.5)]
                  hover:-translate-y-1
                  overflow-hidden
                "
              >
                {/* Gloss effect */}
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
                  <div className="w-10 h-10 rounded-full border-2 border-white dark:border-[#09090b] bg-[#4f6ef7] flex items-center justify-center text-[10px] font-bold text-white">
                    +600
                  </div>
                </div>
                <div className="flex flex-col">
                  <span className="text-sm font-bold text-gray-900 dark:text-white leading-none mb-1">
                    Terpercaya
                  </span>
                  <span className="text-[11px] text-gray-500 uppercase tracking-widest font-semibold opacity-80">
                    Bisnis Indonesia
                  </span>
                </div>
              </div>
            </div>
          </motion.div>
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
