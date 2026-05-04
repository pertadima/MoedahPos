'use client';

import Image from 'next/image';

export default function Footer() {
  return (
    <footer className="py-8 border-t border-gray-100 bg-white">
      <div className="max-w-7xl mx-auto px-6 flex flex-col items-center gap-4">
        <Image
          src="/logo-dashboard-light.svg"
          alt="Moedah"
          width={100}
          height={28}
          className="h-7 w-auto"
        />
        <p className="text-xs text-gray-400">
          &copy; {new Date().getFullYear()} Moedah POS. All rights reserved.
        </p>
      </div>
    </footer>
  );
}
