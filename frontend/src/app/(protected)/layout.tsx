'use client';

import { useRouter } from 'next/navigation';
import { Loader2 } from 'lucide-react';
import { useAuth } from '@/lib/auth/AuthContext';
import { SidebarProvider, useSidebar } from '@/lib/context/SidebarContext';
import Sidebar from '@/components/layout/Sidebar';

function LayoutContent({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, isLoading } = useAuth();
  const { isCollapsed } = useSidebar();
  const router = useRouter();

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
    router.replace('/login');
    return null;
  }

  const sidebarWidth = isCollapsed ? 64 : 240;

  return (
    <div style={{ display: 'flex' }}>
      <Sidebar />
      <div
        style={{ flex: 1, marginLeft: sidebarWidth, transition: 'margin-left 0.3s ease-in-out' }}
      >
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
