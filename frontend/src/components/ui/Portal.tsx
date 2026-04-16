'use client';

import { useEffect, useState } from 'react';
import { createPortal } from 'react-dom';

/**
 * A Portal component that ensures modals and drawers are rendered
 * at the body level to avoid "stacking context traps" from
 * transformed parent elements (e.g. reveal-animate).
 */
export default function Portal({ children }: { children: React.ReactNode }) {
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    // We defer the mounting state to the next tick to avoid
    // "cascading render" lint errors and ensure clean hydration.
    const timer = setTimeout(() => setMounted(true), 0);
    return () => clearTimeout(timer);
  }, []);

  if (!mounted) return null;

  return createPortal(children, document.body);
}
