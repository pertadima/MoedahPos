'use client';

import { motion } from 'framer-motion';
import { Receipt, Percent, FileText, CheckCircle2 } from 'lucide-react';

const trustItems = [
  {
    icon: Receipt,
    title: 'Terima Semua Pembayaran',
    desc: 'Catat pembayaran Tunai, QRIS, GoPay, OVO, dan Transfer Bank dengan mudah.',
  },
  {
    icon: Percent,
    title: 'Patuh Pajak Otomatis',
    desc: 'Kalkulasi PPN/PB1 otomatis di setiap struk sesuai dengan persentase default toko Anda.',
  },
  {
    icon: FileText,
    title: 'Struk Digital & Cetak',
    desc: 'Kirim struk via WhatsApp atau cetak dengan printer kasir bluetooth.',
  },
];

export default function TrustSection() {
  return (
    <section className="py-24 bg-gray-50 dark:bg-gray-900/50 border-y border-gray-200 dark:border-gray-800">
      <div className="max-w-7xl mx-auto px-6">
        <div className="grid lg:grid-cols-2 gap-16 items-center">
          <motion.div
            initial={{ opacity: 0, x: -20 }}
            whileInView={{ opacity: 1, x: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.6 }}
          >
            <h2 className="text-3xl sm:text-4xl font-bold text-gray-900 dark:text-white tracking-tight mb-6">
              Sesuai dengan Standar Bisnis Indonesia
            </h2>
            <p className="text-lg text-gray-600 dark:text-gray-400 mb-8 leading-relaxed">
              Dari pengelolaan pajak daerah hingga integrasi metode pembayaran lokal yang familiar,
              Moedah POS dirancang untuk mempermudah operasional harian Anda tanpa perlu setup yang
              rumit.
            </p>

            <ul className="space-y-6">
              {trustItems.map((item, idx) => (
                <li key={idx} className="flex items-start gap-4">
                  <div className="w-10 h-10 rounded-lg bg-blue-100 dark:bg-blue-500/20 flex items-center justify-center shrink-0 text-[#0070F3]">
                    <item.icon size={20} />
                  </div>
                  <div>
                    <h4 className="text-lg font-bold text-gray-900 dark:text-white mb-1">
                      {item.title}
                    </h4>
                    <p className="text-gray-600 dark:text-gray-400 leading-relaxed">{item.desc}</p>
                  </div>
                </li>
              ))}
            </ul>
          </motion.div>

          <motion.div
            initial={{ opacity: 0, x: 20 }}
            whileInView={{ opacity: 1, x: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.6, delay: 0.2 }}
            className="relative"
          >
            <div className="aspect-square max-w-md mx-auto rounded-3xl bg-gradient-to-br from-gray-100 to-white dark:from-gray-800 dark:to-gray-900 border border-gray-200 dark:border-gray-700 shadow-2xl p-8 relative overflow-hidden flex flex-col">
              <div className="flex-1 flex flex-col justify-center">
                <div className="text-center mb-8">
                  <h3 className="text-2xl font-bold text-gray-900 dark:text-white mb-2">
                    Total Tagihan
                  </h3>
                  <div className="text-4xl font-extrabold text-[#0070F3]">Rp 150.000</div>
                  <div className="text-sm text-gray-500 mt-2">Termasuk PPN 11%</div>
                </div>

                <div className="space-y-3">
                  <div className="p-4 rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-black flex items-center justify-between">
                    <span className="font-medium text-gray-900 dark:text-white">QRIS</span>
                    <CheckCircle2 size={20} className="text-gray-300" />
                  </div>
                  <div className="p-4 rounded-xl border-2 border-[#0070F3] bg-blue-50 dark:bg-blue-500/10 flex items-center justify-between">
                    <span className="font-bold text-[#0070F3]">Tunai</span>
                    <CheckCircle2 size={20} className="text-[#0070F3]" />
                  </div>
                  <div className="p-4 rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-black flex items-center justify-between">
                    <span className="font-medium text-gray-900 dark:text-white">Transfer Bank</span>
                    <CheckCircle2 size={20} className="text-gray-300" />
                  </div>
                </div>
              </div>
            </div>
          </motion.div>
        </div>
      </div>
    </section>
  );
}
