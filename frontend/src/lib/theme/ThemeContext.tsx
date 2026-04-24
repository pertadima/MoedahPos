'use client';

import { createContext, useContext, useEffect, useState, type ReactNode } from 'react';

type Theme = 'dark' | 'light';

interface ThemeContextValue {
  theme: Theme;
  toggleTheme: () => void;
  isDark: boolean;
}

const ThemeContext = createContext<ThemeContextValue>({
  theme: 'dark',
  toggleTheme: () => {},
  isDark: true,
});

export function ThemeProvider({ children }: { children: ReactNode }) {
  // Start with 'dark' to match server-side rendering and avoid hydration mismatch.
  // The blocking script in layout.tsx ensures the correct classes are on <html>
  // before the first paint, preventing flashes.
  const [theme, setTheme] = useState<Theme>('dark');

  useEffect(() => {
    const saved = localStorage.getItem('moedah-theme') as Theme;
    if (saved) {
      // eslint-disable-next-line react-hooks/set-state-in-effect
      setTheme(saved);
    }
  }, []);

  // Sync the `dark` class on <html> whenever theme changes.
  // Tailwind v4's @custom-variant dark targets `.dark` ancestors,
  // so we toggle the class instead of the data-theme attribute.
  useEffect(() => {
    document.documentElement.classList.toggle('dark', theme === 'dark');
    // Keep data-theme for the legacy CSS vars still used in globals.css
    document.documentElement.setAttribute('data-theme', theme);
  }, [theme]);

  const toggleTheme = () => {
    setTheme(prev => {
      const next: Theme = prev === 'dark' ? 'light' : 'dark';
      localStorage.setItem('moedah-theme', next);
      return next;
    });
  };

  return (
    <ThemeContext.Provider value={{ theme, toggleTheme, isDark: theme === 'dark' }}>
      {children}
    </ThemeContext.Provider>
  );
}

export function useTheme() {
  return useContext(ThemeContext);
}
