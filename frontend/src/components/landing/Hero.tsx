'use client';

import { motion } from 'framer-motion';
import Link from 'next/link';
import Image from 'next/image';
import { ArrowRight, Sparkles } from 'lucide-react';
import { useMemo } from 'react';

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
    <section className="relative min-h-[92vh] flex flex-col justify-center bg-[#0884F6] text-white overflow-hidden transition-colors duration-700">
      {/* ── BACKGROUND ARTILLERY ─────────────────────────────────────────── */}

      <div className="absolute inset-0 z-0 pointer-events-none overflow-hidden">
        <div className="absolute -top-[10%] -right-[5%] w-[60%] h-[60%] bg-white/[0.08] blur-[120px] rounded-full" />
        <div className="absolute -bottom-[20%] -left-[10%] w-[70%] h-[70%] bg-[#FFA724]/[0.08] blur-[140px] rounded-full" />
      </div>

      <div
        className="absolute inset-0 z-0 opacity-[0.2] pointer-events-none [mask-image:radial-gradient(ellipse_at_center,black,transparent_75%)]"
        style={{
          backgroundImage: `
            linear-gradient(to right, #ffffff10 1px, transparent 1px),
            linear-gradient(to bottom, #ffffff10 1px, transparent 1px)
          `,
          backgroundSize: '64px 64px',
        }}
      />

      <div className="absolute inset-0 pointer-events-none">
        {particles.map((p, i) => (
          <motion.div
            key={i}
            className="absolute w-1 h-1 bg-white/40 rounded-full"
            style={{ left: p.x, top: p.y }}
            animate={{
              y: [0, -40, 0],
              opacity: [0, 1, 0],
            }}
            transition={{ duration: p.d, repeat: Infinity, delay: p.delay }}
          />
        ))}
      </div>

      {/* ── CONTENT ─────────────────────────────────────────────────────── */}

      <div className="relative z-10 max-w-7xl mx-auto px-6 pt-20 pb-20 w-full flex flex-col lg:flex-row items-center lg:items-stretch gap-12">
        <div className="flex-1 max-w-2xl text-left py-8 flex flex-col justify-center shrink-0">
          <motion.div
            initial={{ opacity: 0, x: -30 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ duration: 0.8, ease: [0.16, 1, 0.3, 1] }}
          >
            <div className="inline-flex items-center gap-2 mb-8 px-4 py-1.5 rounded-full bg-white/10 border border-white/20 shadow-sm backdrop-blur-md">
              <Sparkles size={14} className="text-[#FFA724] animate-pulse" />
              <span className="text-[11px] font-bold tracking-[0.15em] uppercase text-blue-50">
                Next-Gen POS for Modern Business
              </span>
            </div>

            <h1 className="text-5xl sm:text-6xl lg:text-[4.5rem] leading-[0.95] tracking-tight text-white font-black mb-8">
              Kelola Bisnis
              <br />
              <span className="relative inline-block text-transparent bg-clip-text bg-gradient-to-r from-white via-[#FFA724] to-white bg-[length:200%_auto] animate-gradient-flow">
                Tanpa Henti.
              </span>
            </h1>

            <motion.p
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.4 }}
              className="text-lg text-blue-50 leading-relaxed mb-10 font-medium max-w-lg"
            >
              Tingkatkan efisiensi dengan sistem POS offline-first yang secara otomatis
              menyinkronkan akuntansi, inventaris, dan loyalitas pelanggan Anda ke cloud.
            </motion.p>

            <div className="flex items-center gap-6 flex-wrap">
              <Link
                href="/login"
                className="
                  group relative inline-flex items-center gap-3
                  h-14 px-10 rounded-2xl text-base font-bold text-[#0884F6]
                  bg-white hover:bg-blue-50
                  transition-all duration-300
                  shadow-xl shadow-black/10
                  hover:-translate-y-1
                  overflow-hidden
                "
              >
                <div className="absolute inset-0 bg-gradient-to-r from-transparent via-[#0884F6]/5 to-transparent -translate-x-full group-hover:translate-x-full transition-transform duration-1000" />
                Mulai Sekarang
                <ArrowRight size={20} className="group-hover:translate-x-1 transition-transform" />
              </Link>

              <div className="flex items-center gap-4">
                <div className="flex -space-x-3">
                  {[1, 2, 3].map(i => (
                    <div
                      key={i}
                      className="w-10 h-10 rounded-full border-2 border-[#0884F6] bg-white/20"
                    />
                  ))}
                </div>
                <div className="flex flex-col">
                  <span className="text-xs font-bold text-white leading-none mb-1">Terpercaya</span>
                  <span className="text-[10px] text-blue-100 uppercase tracking-widest font-semibold opacity-80">
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
            {[1, 2, 3].map(i => (
              <motion.div
                key={i}
                initial={{ opacity: 0, y: 20 }}
                animate={{
                  opacity: 1,
                  y: [0, -15, 0],
                }}
                transition={{
                  opacity: { duration: 0.6, delay: 0.3 + i * 0.1 },
                  y: {
                    duration: 4 + i * 0.5,
                    repeat: Infinity,
                    ease: 'easeInOut',
                    delay: 0.3 + i * 0.1,
                  },
                }}
              >
                <div className="aspect-[1/2] relative rounded-lg transition-transform duration-500 hover:scale-[1.05] hover:z-30">
                  <Image
                    src={`/mockuplight-${i}.png`}
                    alt={`POS ${i}`}
                    fill
                    className="object-contain dark:hidden"
                    sizes="(max-width: 1024px) 33vw, 300px"
                  />
                  <Image
                    src={`/mockupdark-${i}.png`}
                    alt={`POS ${i}`}
                    fill
                    className="object-contain hidden dark:block"
                    sizes="(max-width: 1024px) 33vw, 300px"
                  />
                </div>
              </motion.div>
            ))}
          </div>
          <div className="absolute right-0 top-1/2 -translate-y-1/2 w-[400px] h-[400px] bg-white/10 blur-[120px] -z-10" />
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
