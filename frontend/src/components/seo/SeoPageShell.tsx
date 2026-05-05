import type { ReactNode } from 'react';

type SeoPageShellProps = {
  children: ReactNode;
  maxWidth?: '4xl' | '5xl';
};

const maxWidthClass: Record<NonNullable<SeoPageShellProps['maxWidth']>, string> = {
  '4xl': 'max-w-4xl',
  '5xl': 'max-w-5xl',
};

export default function SeoPageShell({ children, maxWidth = '4xl' }: SeoPageShellProps) {
  return (
    <main className="relative overflow-hidden bg-[#f8f9fb]">
      <div className="absolute inset-x-0 top-0 h-56 bg-gradient-to-b from-[#0884F6]/10 to-transparent" />
      <div className={`relative mx-auto ${maxWidthClass[maxWidth]} px-4 py-12`}>{children}</div>
    </main>
  );
}
