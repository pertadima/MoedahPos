'use client';

import Image from 'next/image';
import Link from 'next/link';

export default function Footer() {
  return (
    <footer className="py-12 border-t border-gray-200 bg-white">
      <div className="max-w-7xl mx-auto px-6">
        <div className="flex flex-col lg:flex-row gap-12 lg:justify-between lg:items-start">
          <div>
            <Image
              src="/logo-dashboard-dark.svg"
              alt="Moedah"
              width={120}
              height={32}
              className="h-8 w-auto mb-4"
            />
            <p className="text-xs text-gray-500 max-w-[240px] leading-relaxed">
              Sistem Point of Sale modern untuk bisnis retail & F&B di Indonesia. Berdayakan bisnis
              Anda dengan data.
            </p>
          </div>

          <div className="flex flex-col sm:flex-row items-center gap-8">
            <div className="flex items-center gap-6">
              <Link
                href="/login"
                className="text-xs font-semibold text-gray-600 hover:text-[#4f6ef7] transition-colors"
              >
                Masuk
              </Link>
              <Link
                href="/dashboard"
                className="text-xs font-semibold text-gray-600 hover:text-[#4f6ef7] transition-colors"
              >
                Dashboard
              </Link>
            </div>
            <p className="text-xs text-gray-400">
              &copy; {new Date().getFullYear()} Moedah POS. All rights reserved.
            </p>
          </div>
        </div>
      </div>
    </footer>
  );
}
