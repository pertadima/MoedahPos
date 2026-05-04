'use client';

import { useState, useEffect } from 'react';
import Link from 'next/link';
import Image from 'next/image';
import { motion } from 'framer-motion';

export default function Navbar() {
  const [scrolled, setScrolled] = useState(false);

  useEffect(() => {
    const handleScroll = () => setScrolled(window.scrollY > 50);
    window.addEventListener('scroll', handleScroll, { passive: true });
    handleScroll();
    return () => window.removeEventListener('scroll', handleScroll);
  }, []);

  return (
    <motion.header
      initial={{ opacity: 0, y: -20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.5 }}
      className={`fixed top-0 left-0 right-0 z-50 h-16 border-b transition-colors duration-300 ${
        scrolled ? 'bg-white border-gray-200' : 'bg-[#0884F6] border-white/10'
      }`}
    >
      <div className="max-w-7xl mx-auto px-6 h-full flex items-center justify-between">
        <Link href="/" className="flex items-center gap-3">
          <Image
            src={scrolled ? '/logo-dashboard-light.svg' : '/logo-dashboard-dark.svg'}
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
            className={`text-sm font-semibold px-3 transition-colors ${
              scrolled ? 'text-gray-900 hover:text-gray-600' : 'text-blue-100 hover:text-white'
            }`}
          >
            Masuk
          </Link>

          <Link
            href="/login"
            className={`h-10 px-5 flex items-center text-sm font-bold rounded-lg transition-all duration-200 shadow-lg shadow-black/5 ${
              scrolled
                ? 'bg-[#0884F6] text-white hover:bg-[#0770d4]'
                : 'bg-white text-[#0884F6] hover:bg-blue-50'
            }`}
          >
            Mulai Sekarang
          </Link>
        </nav>
      </div>
    </motion.header>
  );
}
