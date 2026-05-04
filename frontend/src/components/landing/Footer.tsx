'use client';

import Image from 'next/image';
import Link from 'next/link';
import { useTheme } from '@/lib/theme/ThemeContext';

export default function Footer() {
  const { isDark } = useTheme();

  return (
    <footer className="py-12 border-t border-gray-200 dark:border-white/[0.05] bg-white dark:bg-[#09090b] transition-colors duration-500">
      <div className="max-w-7xl mx-auto px-6 flex flex-col sm:flex-row items-center justify-between gap-8 text-center sm:text-left">
        <div className="flex flex-col gap-4">
          <Image
            src={isDark ? '/logo-dashboard-dark.svg' : '/logo-dashboard-light.svg'}
            alt="Moedah"
            width={110}
            height={28}
            className="h-6 w-auto opacity-80"
          />
          <p className="text-xs text-gray-500 dark:text-gray-500 max-w-[240px] leading-relaxed">
            Sistem Point of Sale modern untuk bisnis retail & F&B di Indonesia. Berdayakan bisnis
            Anda dengan data.
          </p>
        </div>

        <div className="flex flex-col sm:flex-row items-center gap-8">
          <div className="flex items-center gap-6">
            <Link
              href="/login"
              className="text-xs font-semibold text-gray-600 dark:text-gray-400 hover:text-[#4f6ef7] dark:hover:text-[#4f6ef7] transition-colors"
            >
              Masuk
            </Link>
            <Link
              href="/dashboard"
              className="text-xs font-semibold text-gray-600 dark:text-gray-400 hover:text-[#4f6ef7] dark:hover:text-[#4f6ef7] transition-colors"
            >
              Dashboard
            </Link>
          </div>
          <p className="text-xs text-gray-400 dark:text-gray-600">
            &copy; {new Date().getFullYear()} Moedah POS. All rights reserved.
          </p>
        </div>
      </div>
    </footer>
  );
}
