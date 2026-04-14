'use client';

import { LogOut, Sun, Moon } from 'lucide-react';
import { useAuth } from '@/lib/auth/AuthContext';
import { useTheme } from '@/lib/theme/ThemeContext';

export default function Header() {
  const { user, selectedStore, logout } = useAuth();
  const { isDark, toggleTheme } = useTheme();

  return (
    <header className="header">
      <div className="header-left">{/* Can put page titles or breadcrumbs here if needed */}</div>

      <div className="header-right">
        {/* Theme Toggle */}
        <button
          onClick={toggleTheme}
          className="btn btn-ghost btn-sm"
          title={isDark ? 'Mode Terang' : 'Mode Gelap'}
          style={{ padding: '8px' }}
        >
          {isDark ? <Sun size={18} /> : <Moon size={18} />}
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
