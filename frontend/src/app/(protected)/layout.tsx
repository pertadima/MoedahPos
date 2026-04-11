'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { Loader2, Menu } from 'lucide-react';
import { useAuth } from '@/lib/auth/AuthContext';
import { SidebarProvider, useSidebar } from '@/lib/context/SidebarContext';
import Sidebar from '@/components/layout/Sidebar';

function LayoutContent({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, isLoading } = useAuth();
  const { isCollapsed, toggleCollapsed } = useSidebar();
  const router = useRouter();

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

  if (!isAuthenticated) {
    return null;
  }

  return (
    <div style={{ display: 'flex', width: '100%' }}>
      <Sidebar />
      <div
        className={`main-layout ${isCollapsed ? 'collapsed' : 'expanded'}`}
        style={{ flex: 1, width: '100%' }}
      >
        <div className="lg:hidden flex items-center p-3 border-b border-[var(--border-md)] bg-[var(--bg-card)] sticky top-0 z-20">
          <button
            onClick={toggleCollapsed}
            className="btn btn-ghost btn-sm"
            style={{ padding: '6px' }}
          >
            <Menu size={20} />
          </button>
          <span style={{ marginLeft: '12px', fontWeight: 600, fontSize: '0.9rem' }}>MoedahPOS</span>
        </div>
        {children}
      </div>
    </div>
  );
}

export default function ProtectedLayout({ children }: { children: React.ReactNode }) {
  return (
    <SidebarProvider>
      <LayoutContent>{children}</LayoutContent>
    </SidebarProvider>
  );
}
