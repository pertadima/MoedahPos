'use client';

import Link from 'next/link';
import Image from 'next/image';
import { motion } from 'framer-motion';
import { useTheme } from '@/lib/theme/ThemeContext';
import { Sun, Moon } from 'lucide-react';

export default function Navbar() {
  const { theme, toggleTheme } = useTheme();

  return (
    <motion.header
      initial={{ y: -20, opacity: 0 }}
      animate={{ y: 0, opacity: 1 }}
      transition={{ duration: 0.5 }}
      className="fixed top-0 left-0 right-0 z-50 bg-white/80 dark:bg-black/50 backdrop-blur-md border-b border-gray-200 dark:border-gray-800"
    >
      <div className="max-w-7xl mx-auto px-6 h-16 flex items-center justify-between">
        <Link href="/" className="flex items-center gap-3">
          <Image
            src={theme === 'dark' ? '/logo-dashboard-dark.svg' : '/logo-dashboard-light.svg'}
            alt="Moedah Logo"
            width={120}
            height={32}
            className="h-8 w-auto"
            priority
          />
        </Link>
        <nav className="flex items-center gap-4">
          <button
            type="button"
            onClick={toggleTheme}
            className="p-2 text-gray-500 hover:text-gray-900 dark:text-gray-400 dark:hover:text-white transition-colors rounded-full hover:bg-gray-100 dark:hover:bg-gray-800"
            aria-label="Toggle dark mode"
          >
            {theme === 'dark' ? <Sun size={18} /> : <Moon size={18} />}
          </button>
          <Link
            href="/dashboard"
            className="text-sm font-medium bg-[#0070F3] text-white px-4 py-2 rounded-full hover:bg-blue-600 transition-colors shadow-sm"
          >
            Dashboard
          </Link>
        </nav>
      </div>
    </motion.header>
  );
}
