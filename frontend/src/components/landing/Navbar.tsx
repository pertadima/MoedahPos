'use client';

import Link from 'next/link';
import Image from 'next/image';
import { motion } from 'framer-motion';
import { Sun, Moon } from 'lucide-react';
import { useTheme } from '@/lib/theme/ThemeContext';

export default function Navbar() {
  const { toggleTheme, isDark } = useTheme();

  return (
    <motion.header
      initial={{ opacity: 0, y: -20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.5 }}
      className="fixed top-0 left-0 right-0 z-50 h-16 bg-[#0884F6] border-b border-white/10"
    >
      <div className="max-w-7xl mx-auto px-6 h-full flex items-center justify-between">
        <Link href="/" className="flex items-center gap-3">
          <Image
            src="/logo-dashboard-dark.svg"
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
            className="
              w-10 h-10 flex items-center justify-center rounded-lg
              text-blue-100 hover:text-white
              hover:bg-white/10
              transition-all duration-200
            "
            aria-label="Toggle theme"
          >
            {isDark ? <Sun size={18} /> : <Moon size={18} />}
          </button>

          <Link href="/login" className="text-sm font-semibold text-blue-50 hover:text-white px-3">
            Masuk
          </Link>

          <Link
            href="/dashboard"
            className="
              h-10 px-5 flex items-center
              text-sm font-bold text-[#0884F6]
              bg-white hover:bg-blue-50
              rounded-lg transition-all duration-200
              shadow-lg shadow-black/5
            "
          >
            Mulai Sekarang
          </Link>
        </nav>
      </div>
    </motion.header>
  );
}
