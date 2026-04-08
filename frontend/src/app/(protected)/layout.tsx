'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { Loader2 } from 'lucide-react';
import { useAuth } from '@/lib/auth/AuthContext';
import Sidebar from '@/components/layout/Sidebar';

export default function ProtectedLayout({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, isLoading } = useAuth();
  const router = useRouter();
  const [isCollapsed, setIsCollapsed] = useState(false);
  const [isMounted, setIsMounted] = useState(false);

  useEffect(() => {
    setIsMounted(true);
    const saved = localStorage.getItem('sidebar-collapsed');
    if (saved !== null) {
      setIsCollapsed(JSON.parse(saved));
    }
  }, []);

  useEffect(() => {
    const checkCollapsed = () => {
      const saved = localStorage.getItem('sidebar-collapsed');
      if (saved !== null) {
        setIsCollapsed(JSON.parse(saved));
      }
    };
    window.addEventListener('storage', checkCollapsed);
    const interval = setInterval(checkCollapsed, 100);
    return () => {
      window.removeEventListener('storage', checkCollapsed);
      clearInterval(interval);
    };
  }, []);

  useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      router.replace('/login');
    }
  }, [isAuthenticated, isLoading, router]);

  if (isLoading) {
    return (
      <div
        style={{
          minHeight: '100vh',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          background: 'var(--bg-base)',
        }}
      >
        <Loader2 size={32} className="loading-spin" style={{ color: 'var(--accent-em)' }} />
      </div>
    );
  }

  if (!isAuthenticated) return null;

  const sidebarWidth = isMounted ? (isCollapsed ? 64 : 240) : 240;

  return (
    <div style={{ display: 'flex' }}>
      <Sidebar />
      <div style={{ flex: 1, marginLeft: sidebarWidth, transition: 'margin-left 0.3s ease-in-out' }}>
        {children}
      </div>
    </div>
  );
}
