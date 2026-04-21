'use client';

import { useState, useEffect, useCallback, useRef } from 'react';
import { useLiveQuery } from 'dexie-react-hooks';
import {
  Cloud,
  CloudOff,
  CloudUpload,
  Loader2,
  AlertTriangle,
  ArrowRight,
  X,
  Trash2,
  RotateCcw,
} from 'lucide-react';
import { db } from '@/lib/dexie';
import { transactionsApi } from '@/lib/api/transactions';
import Portal from '@/components/ui/Portal';
import { formatRp } from '@/lib/utils';

export default function SyncStatusWidget() {
  const [isOnline, setIsOnline] = useState(true);
  const [isSyncing, setIsSyncing] = useState(false);
  const [syncProgress, setSyncProgress] = useState({ current: 0, total: 0 });
  const [showModal, setShowModal] = useState(false);

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

    // Only process pending transactions that haven't explicitly failed
    // AND allow a 10s grace period for the foreground hook to finish its own sync, preventing double API calls.
    const nowMs = new Date().getTime();
    const toSync = dirtyTransactions.filter(t => {
      if (t.status === 'failed') return false;
      if (!t.created_at) return true;
      const ageMs = nowMs - new Date(t.created_at).getTime();
      return ageMs > 10000; // 10 seconds grace period
    });
    if (toSync.length === 0) return;

    isSyncingRef.current = true;
    setIsSyncing(true);
    const total = toSync.length;
    setSyncProgress({ current: 0, total });

    for (let i = 0; i < total; i++) {
      const tx = toSync[i];
      try {
        await transactionsApi.syncOffline(tx.store_id, tx);
        await db.transactions.update(tx.id, { is_dirty: false });

        // Artificial delay to allow visual feedback of the sync progress
        await new Promise(resolve => setTimeout(resolve, 800));
      } catch (err: any) {
        console.error('Network or API failure during background sync', err);
        if (err.name === 'ApiError') {
          let errorMsg = err.message;
          if (errorMsg.toLowerCase().includes('insufficient stock')) {
            const match = errorMsg.match(
              /items?:\s*(.+?)\s*\(have\s*([\d.]+),\s*need\s*([\d.]+)\)/i
            );
            if (match) {
              errorMsg = `Stok tidak mencukupi: ${match[1]} (tersedia ${match[2]}, butuh ${match[3]})`;
            } else {
              errorMsg = 'Stok tidak mencukupi untuk item pesanan.';
            }
          }

          // Backend rejected it permanently. Set error and status to failed, keeping it dirty so user can see it.
          await db.transactions.update(tx.id, { sync_error: errorMsg, status: 'failed' });
        } else {
          // True network error, assume internet dropped and break loop
          break;
        }
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
      const hasPending = dirtyTransactions.some(t => t.status !== 'failed');
      if (hasPending) {
        processSync();
      }
    }
  }, [isOnline, dirtyTransactions, processSync]);

  const handleDiscard = async (id: string) => {
    await db.transactions.update(id, { is_dirty: false, status: 'cancelled_offline' });
  };

  const handleRetrySingle = async (id: string) => {
    await db.transactions.update(id, { status: 'pending', sync_error: '' });
    if (isOnline) processSync();
  };

  // Combined render
  return (
    <div className="flex items-center">
      {/* Network Status Badge */}
      <div
        className={`flex items-center gap-1.5 px-3 py-1.5 rounded-full text-xs font-semibold shadow-sm mr-2 transition-all ${
          isOnline
            ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400 opacity-80 hover:opacity-100'
            : 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'
        }`}
      >
        {isOnline ? (
          <Cloud size={14} strokeWidth={2.5} />
        ) : (
          <CloudOff size={14} strokeWidth={2.5} />
        )}
        <span className="hidden sm:inline">{isOnline ? 'Online' : 'Offline Mode'}</span>
      </div>

      {/* Sync Queue Badge (Only if there's data to sync/failed) */}
      {dirtyTransactions && dirtyTransactions.length > 0 && (
        <button
          onClick={() => setShowModal(true)}
          className={`flex items-center gap-1.5 px-3 py-1.5 rounded-full text-xs font-semibold shadow-sm mr-2 transition-all cursor-pointer hover:opacity-80 ${
            isSyncing
              ? 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400'
              : dirtyTransactions.some(t => t.status === 'failed')
                ? 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'
                : 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400'
          }`}
        >
          {isSyncing ? (
            <Loader2 size={14} className="animate-spin" strokeWidth={2.5} />
          ) : dirtyTransactions.some(t => t.status === 'failed') ? (
            <AlertTriangle size={14} strokeWidth={2.5} />
          ) : (
            <CloudUpload size={14} className="animate-bounce" strokeWidth={2.5} />
          )}

          <span>
            {(() => {
              if (isSyncing) return `${syncProgress.total} item sync in progress`;
              const failed = dirtyTransactions.filter(t => t.status === 'failed').length;
              const pending = dirtyTransactions.length - failed;
              const text = [];
              if (failed > 0) text.push(`${failed} item gagal sync`);
              if (pending > 0) text.push(`${pending} item menunggu`);
              return text.join(', ');
            })()}
          </span>
        </button>
      )}

      {/* Sync Queue Modal Portal */}
      {showModal && (
        <Portal>
          <div
            style={{
              position: 'fixed',
              inset: 0,
              zIndex: 5000,
              display: 'flex',
              justifyContent: 'flex-end',
            }}
            onClick={() => setShowModal(false)}
          >
            <style>{`
            @keyframes slideInRight {
              from { transform: translateX(100%); }
              to { transform: translateX(0); }
            }
          `}</style>
            {/* Backdrop */}
            <div
              style={{
                position: 'absolute',
                inset: 0,
                background: 'rgba(0,0,0,0.4)',
                backdropFilter: 'blur(3px)',
              }}
            />
            {/* Sidebar drawer content */}
            <div
              className="card"
              style={{
                position: 'relative',
                width: '100%',
                maxWidth: 480,
                height: '100%',
                borderRadius: 0,
                padding: 0,
                display: 'flex',
                flexDirection: 'column',
                boxShadow: '-8px 0 32px rgba(0,0,0,0.15)',
                animation: 'slideInRight 0.3s cubic-bezier(0.16, 1, 0.3, 1)',
              }}
              onClick={e => e.stopPropagation()}
            >
              <div
                style={{
                  padding: '20px 24px',
                  borderBottom: '1px solid var(--border-md)',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'space-between',
                }}
              >
                <div>
                  <h2
                    style={{
                      fontWeight: 700,
                      fontSize: '1.1rem',
                      display: 'flex',
                      alignItems: 'center',
                      gap: 8,
                    }}
                  >
                    <CloudUpload size={20} className="text-blue-500" />
                    Antrean Sinkronisasi
                  </h2>
                  <p style={{ fontSize: '0.8rem', color: 'var(--text-3)', marginTop: 4 }}>
                    {dirtyTransactions?.length || 0} transaksi tertunda
                  </p>
                </div>
                <button className="btn btn-ghost btn-sm" onClick={() => setShowModal(false)}>
                  <X size={18} />
                </button>
              </div>

              <div style={{ padding: '0', overflowY: 'auto', flex: 1 }}>
                {dirtyTransactions?.length === 0 ? (
                  <div style={{ padding: 40, textAlign: 'center', color: 'var(--text-3)' }}>
                    Tidak ada antrean sinkronisasi.
                  </div>
                ) : (
                  <div style={{ display: 'flex', flexDirection: 'column' }}>
                    {dirtyTransactions?.map(tx => {
                      const isFailed = tx.status === 'failed';
                      return (
                        <div
                          key={tx.id}
                          style={{
                            padding: '16px 24px',
                            borderBottom: '1px solid var(--border-sm)',
                            background: isFailed ? 'rgba(239,68,68,0.03)' : 'transparent',
                          }}
                        >
                          <div
                            style={{
                              display: 'flex',
                              justifyContent: 'space-between',
                              alignItems: 'flex-start',
                              marginBottom: 8,
                            }}
                          >
                            <div>
                              <div
                                style={{
                                  fontWeight: 600,
                                  fontSize: '0.9rem',
                                  display: 'flex',
                                  alignItems: 'center',
                                  gap: 6,
                                }}
                              >
                                INV-{tx.id.slice(0, 8).toUpperCase()}
                                <span
                                  style={{
                                    fontSize: '0.65rem',
                                    padding: '2px 8px',
                                    borderRadius: 12,
                                    background: isFailed
                                      ? 'rgba(239,68,68,0.1)'
                                      : 'rgba(245,158,11,0.1)',
                                    color: isFailed ? '#ef4444' : '#d97706',
                                    fontWeight: 700,
                                  }}
                                >
                                  {isFailed ? 'Gagal' : 'Menunggu'}
                                </span>
                              </div>
                              <div
                                style={{
                                  fontSize: '0.75rem',
                                  color: 'var(--text-3)',
                                  marginTop: 4,
                                }}
                              >
                                {new Date(tx.created_at).toLocaleString('id-ID')}
                              </div>
                            </div>
                            <div style={{ fontWeight: 700, color: 'var(--text-1)' }}>
                              {formatRp(tx.total)}
                            </div>
                          </div>

                          <div
                            style={{
                              padding: '8px 12px',
                              background: 'var(--bg-base)',
                              borderRadius: 6,
                              marginBottom: 12,
                            }}
                          >
                            <div
                              style={{
                                fontSize: '0.7rem',
                                fontWeight: 600,
                                color: 'var(--text-2)',
                                marginBottom: 4,
                              }}
                            >
                              {tx.items.length} ITEM
                            </div>
                            {tx.items.map((item, idx) => (
                              <div
                                key={idx}
                                style={{
                                  display: 'flex',
                                  justifyContent: 'space-between',
                                  fontSize: '0.75rem',
                                  color: 'var(--text-2)',
                                  marginBottom: 2,
                                }}
                              >
                                <span>
                                  {item.quantity}x {item.product_name}
                                </span>
                                <span>{formatRp(item.subtotal)}</span>
                              </div>
                            ))}
                          </div>

                          {isFailed && tx.sync_error && (
                            <div
                              style={{
                                padding: '8px 12px',
                                background: 'rgba(239,68,68,0.1)',
                                borderLeft: '3px solid #ef4444',
                                borderRadius: '0 6px 6px 0',
                                fontSize: '0.8rem',
                                color: '#b91c1c',
                                marginBottom: 12,
                              }}
                            >
                              {tx.sync_error}
                            </div>
                          )}

                          <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end' }}>
                            <button
                              onClick={() => handleDiscard(tx.id)}
                              className="btn btn-ghost btn-sm"
                              style={{
                                color: '#ef4444',
                                padding: '6px 12px',
                                fontSize: '0.75rem',
                                height: 'auto',
                                minHeight: 0,
                              }}
                            >
                              <Trash2 size={12} style={{ marginRight: 4 }} /> Hapus
                            </button>
                            {isFailed && (
                              <button
                                onClick={() => handleRetrySingle(tx.id)}
                                className="btn btn-primary btn-sm"
                                style={{
                                  padding: '6px 12px',
                                  fontSize: '0.75rem',
                                  height: 'auto',
                                  minHeight: 0,
                                }}
                              >
                                <RotateCcw size={12} style={{ marginRight: 4 }} /> Coba Lagi
                              </button>
                            )}
                          </div>
                        </div>
                      );
                    })}
                  </div>
                )}
              </div>

              {dirtyTransactions && dirtyTransactions.length > 0 && (
                <div
                  style={{
                    padding: '16px 24px',
                    borderTop: '1px solid var(--border-md)',
                    background: 'var(--bg-elevated)',
                    display: 'flex',
                    justifyContent: 'flex-end',
                  }}
                >
                  <button
                    className="btn btn-primary"
                    onClick={() => {
                      dirtyTransactions.forEach(t => {
                        if (t.status === 'failed') handleRetrySingle(t.id);
                      });
                      processSync();
                    }}
                    disabled={isSyncing}
                  >
                    {isSyncing ? (
                      <Loader2 size={16} className="animate-spin" />
                    ) : (
                      <ArrowRight size={16} />
                    )}
                    Sinkronkan Semua
                  </button>
                </div>
              )}
            </div>
          </div>
        </Portal>
      )}
    </div>
  );
}
