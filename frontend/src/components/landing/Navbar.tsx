'use client';

import Link from 'next/link';
import { motion } from 'framer-motion';

export default function Navbar() {
  return (
    <motion.header
      initial={{ y: -20, opacity: 0 }}
      animate={{ y: 0, opacity: 1 }}
      transition={{ duration: 0.5 }}
      className="fixed top-0 left-0 right-0 z-50 bg-white/80 dark:bg-black/50 backdrop-blur-md border-b border-gray-200 dark:border-gray-800"
    >
      <div className="max-w-7xl mx-auto px-6 h-16 flex items-center justify-between">
        <Link href="/" className="flex items-center gap-2">
          <div className="w-8 h-8 rounded bg-[#0070F3] flex items-center justify-center text-white font-bold text-lg">
            M
          </div>
          <span className="font-bold text-xl tracking-tight text-gray-900 dark:text-white">
            Moedah POS
          </span>
        </Link>
        <nav className="flex items-center gap-4">
          <Link
            href="/login"
            className="text-sm font-medium text-gray-600 hover:text-gray-900 dark:text-gray-300 dark:hover:text-white transition-colors"
          >
            Masuk
          </Link>
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
