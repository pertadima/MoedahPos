'use client';

import { useState, useEffect } from 'react';
import Image from 'next/image';
import { LogOut, Sun, Moon, Menu } from 'lucide-react';
import { useAuth } from '@/lib/auth/AuthContext';
import { useTheme } from '@/lib/theme/ThemeContext';
import { useSidebar } from '@/lib/context/SidebarContext';
import SyncStatusWidget from '@/components/ui/SyncStatusWidget';

export default function Header() {
  const { user, selectedStore, logout } = useAuth();
  const { isDark, toggleTheme } = useTheme();
  const { toggleCollapsed } = useSidebar();
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setMounted(true);
  }, []);

  return (
    <header className="header">
      <div className="header-left">
        <button
          onClick={toggleCollapsed}
          className="btn btn-ghost btn-sm lg:hidden"
          style={{ padding: '6px' }}
        >
          <Menu size={20} />
        </button>
        <div className="lg:hidden flex items-center ml-2">
          <Image
            src={mounted && isDark ? '/logo-icon-dark.svg' : '/logo-icon-light.svg'}
            alt="Moedah"
            width={32}
            height={32}
            style={{ width: '32px', height: 'auto' }}
            priority
          />
        </div>
      </div>

      <div className="header-right flex items-center">
        {/* Offline Sync Widget */}
        <SyncStatusWidget />

        {/* Theme Toggle */}
        <button
          onClick={toggleTheme}
          className="btn btn-ghost btn-sm"
          title={mounted ? (isDark ? 'Mode Terang' : 'Mode Gelap') : 'Tema'}
          style={{ padding: '8px' }}
        >
          {mounted ? (isDark ? <Sun size={18} /> : <Moon size={18} />) : <Sun size={18} />}
        </button>

        <div className="header-divider" />

        {/* User Profile */}
        {user && (
          <div className="header-profile">
            <div className="header-avatar">{user.name.charAt(0).toUpperCase()}</div>
            <div className="header-user-info">
              <div className="header-user-name">{user.name}</div>
              <div className="header-user-role">{selectedStore?.role ?? user.email}</div>
            </div>
          </div>
        )}

        {/* Logout Button */}
        <button
          onClick={logout}
          className="btn btn-ghost btn-sm"
          title="Keluar"
          style={{ padding: '8px', color: 'var(--text-3)' }}
        >
          <LogOut size={18} />
        </button>
      </div>
    </header>
  );
}
