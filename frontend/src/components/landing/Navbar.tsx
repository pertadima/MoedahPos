'use client';

import { useState, useEffect } from 'react';
import Link from 'next/link';
import Image from 'next/image';
import { motion } from 'framer-motion';
import { Sun, Moon } from 'lucide-react';
import { useTheme } from '@/lib/theme/ThemeContext';

export default function Navbar() {
  const { toggleTheme, isDark } = useTheme();
  const [scrolled, setScrolled] = useState(false);

  useEffect(() => {
    const handleScroll = () => setScrolled(window.scrollY > 50);
    window.addEventListener('scroll', handleScroll, { passive: true });
    handleScroll();
    return () => window.removeEventListener('scroll', handleScroll);
  }, []);

  const bgClass = scrolled
    ? isDark
      ? 'bg-black border-white/10'
      : 'bg-white border-gray-200'
    : 'bg-[#0884F6] border-white/10';

  const textClass = scrolled
    ? isDark
      ? 'text-white hover:text-gray-200'
      : 'text-gray-900 hover:text-gray-600'
    : 'text-blue-100 hover:text-white';

  const btnClass = scrolled
    ? isDark
      ? 'bg-gray-800 text-white hover:bg-gray-700'
      : 'bg-[#0884F6] text-white hover:bg-[#0770d4]'
    : 'bg-white text-[#0884F6] hover:bg-blue-50';

  return (
    <motion.header
      initial={{ opacity: 0, y: -20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.5 }}
      className={`fixed top-0 left-0 right-0 z-50 h-16 border-b transition-colors duration-300 ${bgClass}`}
    >
      <div className="max-w-7xl mx-auto px-6 h-full flex items-center justify-between">
        <Link href="/" className="flex items-center gap-3">
          <Image
            src={scrolled && !isDark ? '/logo-dashboard-light.svg' : '/logo-dashboard-dark.svg'}
            alt="Moedah Logo"
            width={120}
            height={32}
            className="h-8 w-auto transition-opacity"
            priority
          />
        </Link>

        <nav className="flex items-center gap-4">
          <button
            type="button"
            onClick={toggleTheme}
            className={`w-10 h-10 flex items-center justify-center rounded-lg hover:bg-white/10 transition-all duration-200 ${textClass}`}
            aria-label="Toggle theme"
          >
            {isDark ? <Sun size={18} /> : <Moon size={18} />}
          </button>

          <Link
            href="/login"
            className={`text-sm font-semibold px-3 transition-colors ${textClass}`}
          >
            Masuk
          </Link>

          <Link
            href="/login"
            className={`h-10 px-5 flex items-center text-sm font-bold rounded-lg transition-all duration-200 shadow-lg shadow-black/5 ${btnClass}`}
          >
            Mulai Sekarang
          </Link>
        </nav>
      </div>
    </motion.header>
  );
}
