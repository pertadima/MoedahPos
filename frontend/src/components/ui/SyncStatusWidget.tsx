'use client';

import { useState, useEffect, useCallback, useRef } from 'react';
import { useLiveQuery } from 'dexie-react-hooks';
import { Cloud, CloudOff, CloudUpload, Loader2 } from 'lucide-react';
import { db } from '@/lib/dexie';

export default function SyncStatusWidget() {
  const [isOnline, setIsOnline] = useState(true);
  const [isSyncing, setIsSyncing] = useState(false);
  const [syncProgress, setSyncProgress] = useState({ current: 0, total: 0 });
  const isSyncingRef = useRef(false);

  // Monitor the number of offline transactions dynamically
  const dirtyTransactions = useLiveQuery(
    () => db.transactions.filter(t => t.is_dirty === true).toArray(),
    []
  );

  useEffect(() => {
    // Initial load
    if (typeof window !== 'undefined') {
      setIsOnline(window.navigator.onLine);
    }

    const handleOnline = () => setIsOnline(true);
    const handleOffline = () => setIsOnline(false);

    window.addEventListener('online', handleOnline);
    window.addEventListener('offline', handleOffline);

    return () => {
      window.removeEventListener('online', handleOnline);
      window.removeEventListener('offline', handleOffline);
    };
  }, []);

  const processSync = useCallback(async () => {
    if (!dirtyTransactions || dirtyTransactions.length === 0 || !isOnline || isSyncingRef.current) {
      return;
    }

    isSyncingRef.current = true;
    setIsSyncing(true);
    const total = dirtyTransactions.length;
    setSyncProgress({ current: 0, total });

    for (let i = 0; i < total; i++) {
        const tx = dirtyTransactions[i];
        try {
            const response = await fetch(`/api/v1/stores/${tx.store_id}/transactions`, {
              method: 'POST',
              headers: { 'Content-Type': 'application/json' },
              body: JSON.stringify(tx),
            });

            if (response.ok) {
                await db.transactions.update(tx.id, { is_dirty: false });
            } else {
                console.warn('Backend rejected transaction sync', tx.id, response.statusText);
                // Depending on the logic, maybe we shouldn't break, but rather skip
            }
        } catch (err) {
            console.error('Network failure during background sync', err);
            // Break loop on first network error assuming internet dropped
            break;
        } finally {
            setSyncProgress(prev => ({ ...prev, current: prev.current + 1 }));
        }
    }

    setIsSyncing(false);
    isSyncingRef.current = false;
  }, [dirtyTransactions, isOnline]);

  useEffect(() => {
    // Auto-trigger sync when transitioning online and dirty transactions exist
    if (isOnline && dirtyTransactions && dirtyTransactions.length > 0 && !isSyncingRef.current) {
      processSync();
    }
  }, [isOnline, dirtyTransactions?.length, processSync]);

  if (!isOnline) {
    return (
      <div className="flex items-center gap-1.5 px-3 py-1.5 bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400 rounded-full text-xs font-semibold shadow-sm mr-2 transition-all">
        <CloudOff size={14} strokeWidth={2.5} />
        <span>Offline Mode</span>
      </div>
    );
  }

  if (isSyncing) {
    return (
      <div className="flex items-center gap-1.5 px-3 py-1.5 bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400 rounded-full text-xs font-semibold shadow-sm mr-2 transition-all">
        <Loader2 size={14} className="animate-spin" strokeWidth={2.5} />
        <span>Syncing {syncProgress.current}/{syncProgress.total}</span>
      </div>
    );
  }

  if (dirtyTransactions && dirtyTransactions.length > 0) {
    return (
      <button 
        onClick={processSync}
        className="flex items-center gap-1.5 px-3 py-1.5 bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400 rounded-full text-xs font-semibold shadow-sm hover:bg-yellow-200 dark:hover:bg-yellow-900/50 transition-all mr-2 cursor-pointer"
        title="Click to force sync now"
      >
        <CloudUpload size={14} className="animate-bounce" strokeWidth={2.5} />
        <span>{dirtyTransactions.length} Pending</span>
      </button>
    );
  }

  // All Synced fully Online
  return (
    <div className="flex items-center gap-1.5 px-3 py-1.5 bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400 rounded-full text-xs font-semibold shadow-sm mr-2 opacity-80 hover:opacity-100 transition-all">
      <Cloud size={14} strokeWidth={2.5} />
      <span className="hidden sm:inline">Online</span>
    </div>
  );
}
