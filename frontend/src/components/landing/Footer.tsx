'use client';

import Image from 'next/image';
import { useTheme } from '@/lib/theme/ThemeContext';

export default function Footer() {
  const { theme } = useTheme();

  return (
    <footer className="py-12 bg-white dark:bg-black border-t border-gray-200 dark:border-gray-800">
      <div className="max-w-7xl mx-auto px-6 flex flex-col md:flex-row items-center justify-between gap-4">
        <div className="flex items-center gap-2">
          <Image
            src={theme === 'dark' ? '/logo-dashboard-dark.svg' : '/logo-dashboard-light.svg'}
            alt="Moedah Logo"
            width={120}
            height={32}
            className="h-6 w-auto opacity-70 grayscale"
            priority
          />
        </div>
        <p className="text-sm text-gray-500">
          &copy; {new Date().getFullYear()} Moedah POS. Seluruh hak cipta dilindungi.
        </p>
      </div>
    </footer>
  );
}
