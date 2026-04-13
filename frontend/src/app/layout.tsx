import type { Metadata } from 'next';
import { Inter } from 'next/font/google';
import './globals.css';
import { AuthProvider } from '@/lib/auth/AuthContext';
import { ThemeProvider } from '@/lib/theme/ThemeContext';

/**
 * Inter — variable font loaded via next/font/google.
 * Subsets: latin (covers all Indonesian characters).
 * display: swap prevents invisible text during load.
 * Variable font gives us the full weight axis (100–900)
 * without extra requests.
 */
const inter = Inter({
  subsets: ['latin'],
  display: 'swap',
  variable: '--font-inter',
  // Pre-load only the weights actually used:
  // 400 (body), 500 (medium/nav), 600 (semibold/headings), 700 (bold/values)
  weight: ['400', '500', '600', '700'],
});

export const metadata: Metadata = {
  title: 'Moedah POS — Point of Sale System',
  description: 'Multi-store point of sale management system',
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="id" suppressHydrationWarning className={`dark ${inter.variable}`} data-theme="dark">
      <body className={inter.className}>
        <ThemeProvider>
          <AuthProvider>{children}</AuthProvider>
        </ThemeProvider>
      </body>
    </html>
  );
}
