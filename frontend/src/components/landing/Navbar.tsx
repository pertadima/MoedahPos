'use client';

import { useState, useEffect, useRef } from 'react';
import Link from 'next/link';
import Image from 'next/image';
import { motion } from 'framer-motion';

export default function Navbar() {
  const [scrollY, setScrollY] = useState(0);
  const [section, setSection] = useState<'hero' | 'other'>('hero');
  const heroRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleScroll = () => {
      setScrollY(window.scrollY);
      if (!heroRef.current) return;
      const rect = heroRef.current.getBoundingClientRect();
      setSection(rect.bottom > 100 ? 'hero' : 'other');
    };

    window.addEventListener('scroll', handleScroll, { passive: true });
    handleScroll();
    return () => window.removeEventListener('scroll', handleScroll);
  }, []);

  const isHero = section === 'hero';
  const isAtTop = scrollY < 50;
  const textClass = isHero && !isAtTop ? 'text-white' : isHero ? 'text-white' : 'text-gray-900';
  const hoverClass =
    isHero && !isAtTop ? 'text-white/80' : isHero ? 'hover:text-white/80' : 'hover:text-gray-600';
  const bgClass =
    isAtTop && isHero
      ? 'bg-transparent backdrop-blur-none border-transparent'
      : isHero
        ? 'bg-[#0884F6]/95 backdrop-blur-md border-white/10'
        : 'bg-white/90 backdrop-blur-md border-gray-100 shadow-sm';

  return (
    <>
      <div ref={heroRef} className="absolute top-0 left-0 w-full h-[100vh] pointer-events-none" />
      <motion.header
        initial={{ opacity: 0, y: -20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5 }}
        className={`fixed top-0 left-0 right-0 z-50 h-16 border-b transition-all duration-300 ${bgClass}`}
      >
        <div className="max-w-7xl mx-auto px-6 h-full flex items-center justify-between">
          <Link href="/" className="flex items-center gap-3">
            <Image
              src={isHero ? '/logo-dashboard-dark.svg' : '/logo-dashboard-light.svg'}
              alt="Moedah Logo"
              width={120}
              height={32}
              className="h-8 w-auto transition-opacity"
              priority
            />
          </Link>

          <nav className="flex items-center gap-4">
            <Link
              href="/login"
              className={`text-sm font-semibold px-3 transition-colors ${textClass} ${hoverClass}`}
            >
              Masuk
            </Link>

            <Link
              href="/login"
              className={`h-10 px-5 flex items-center text-sm font-bold rounded-lg transition-all duration-200 shadow-lg shadow-black/5 ${
                isAtTop && isHero
                  ? 'bg-white text-[#0884F6] hover:bg-blue-50'
                  : 'bg-white text-[#0884F6] hover:bg-blue-50'
              }`}
            >
              Mulai Sekarang
            </Link>
          </nav>
        </div>
      </motion.header>
    </>
  );
}
