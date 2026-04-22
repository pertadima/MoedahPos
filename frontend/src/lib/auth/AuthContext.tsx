'use client';

import React, { createContext, useContext, useEffect, useState, useCallback } from 'react';
import { authApi } from '@/lib/api/auth';
import { setTokens, clearTokens, getAccessToken, getRefreshToken } from '@/lib/api/client';
import type { User, UserStore } from '@/types';

interface AuthState {
  user: User | null;
  isLoading: boolean;
  isAuthenticated: boolean;
  selectedStore: UserStore | null;
  stores: UserStore[];
  login: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  selectStore: (storeId: string) => void;
  refreshSession: () => Promise<void>;
}

const AuthContext = createContext<AuthState | null>(null);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [selectedStore, setSelectedStore] = useState<UserStore | null>(null);

  const stores = user?.stores ?? [];

  // On mount: try to restore session from stored refresh token
  useEffect(() => {
    const restore = async () => {
      const token = getAccessToken();
      if (!token) {
        setIsLoading(false);
        return;
      }
      try {
        const res = await authApi.me();
        const u = res.data;
        setUser(u);
        // Restore selected store from localStorage
        const savedStoreId = localStorage.getItem('selected_store_id');
        const storeList = u.stores ?? [];
        const preferred = savedStoreId ? storeList.find(s => s.store_id === savedStoreId) : null;
        const autoSelected = preferred ?? storeList[0] ?? null;
        if (autoSelected) {
          setSelectedStore(autoSelected);
          localStorage.setItem('selected_store_id', autoSelected.store_id);
        }
      } catch {
        clearTokens();
      } finally {
        setIsLoading(false);
      }
    };
    restore();
  }, []);

  const login = useCallback(async (email: string, password: string) => {
    const res = await authApi.login(email, password);
    const { access_token, refresh_token } = res.data;
    setTokens(access_token, refresh_token);
    // Fetch full user profile (includes stores) via /me
    const meRes = await authApi.me();
    const u = meRes.data as User;
    setUser(u);
    const storeList = u.stores ?? [];
    const savedStoreId = localStorage.getItem('selected_store_id');
    const preferred = savedStoreId ? storeList.find(s => s.store_id === savedStoreId) : null;
    const autoSelected = preferred ?? storeList[0] ?? null;
    if (autoSelected) {
      setSelectedStore(autoSelected);
      localStorage.setItem('selected_store_id', autoSelected.store_id);
    }
  }, []);

  const logout = useCallback(async () => {
    try {
      const rt = getRefreshToken();
      if (rt) await authApi.logout(rt);
    } catch {
      /* best effort */
    }
    clearTokens();
    localStorage.removeItem('selected_store_id');
    setUser(null);
    setSelectedStore(null);
    window.location.href = '/login';
  }, []);

  const selectStore = useCallback(
    (storeId: string) => {
      const found = stores.find(s => s.store_id === storeId);
      if (found) {
        setSelectedStore(found);
        localStorage.setItem('selected_store_id', storeId);
      }
    },
    [stores]
  );

  const refreshSession = useCallback(async () => {
    try {
      const res = await authApi.me();
      const u = res.data;
      setUser(u);
      
      // Update selected store if it still exists
      const savedStoreId = localStorage.getItem('selected_store_id');
      const storeList = u.stores ?? [];
      const preferred = savedStoreId ? storeList.find(s => s.store_id === savedStoreId) : null;
      const autoSelected = preferred ?? storeList[0] ?? null;
      if (autoSelected) {
        setSelectedStore(autoSelected);
      }
    } catch {
      // ignore
    }
  }, []);

  return (
    <AuthContext.Provider
      value={{
        user,
        isLoading,
        isAuthenticated: !!user,
        selectedStore,
        stores,
        login,
        logout,
        selectStore,
        refreshSession,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth(): AuthState {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within <AuthProvider>');
  return ctx;
}
