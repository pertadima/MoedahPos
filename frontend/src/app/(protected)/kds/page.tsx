'use client';

import { useEffect, useState, useCallback } from 'react';
import { useAuth } from '@/lib/auth/AuthContext';
import { kdsApi } from '@/lib/api/store-apis';
import type { Transaction } from '@/types';
import {
  Clock,
  CheckSquare,
  Square,
  CheckCircle2,
  ChefHat,
  RefreshCw,
  ShoppingBag,
  UtensilsCrossed,
  AlertCircle,
} from 'lucide-react';
import { ApiError } from '@/lib/api/client';

export default function KDSPage() {
  const { selectedStore } = useAuth();
  const [tickets, setTickets] = useState<Transaction[]>([]);
  const [loading, setLoading] = useState(true);
  const [errorHeader, setErrorHeader] = useState('');
  const [now, setNow] = useState(Date.now());

  const storeId = selectedStore?.store_id;

  const fetchTickets = useCallback(async () => {
    if (!storeId) return;
    try {
      const res = await kdsApi.getTickets(storeId);
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      setTickets(((res.data as any).data as Transaction[]) ?? (res.data as Transaction[]) ?? []);
      setErrorHeader('');
    } catch (err: unknown) {
      if (err instanceof ApiError) {
        setErrorHeader(err.message);
      } else {
        setErrorHeader('Gagal memuat tiket KDS');
      }
    } finally {
      setLoading(false);
    }
  }, [storeId]);

  useEffect(() => {
    fetchTickets();
    // Auto-refresh every 10 seconds
    const interval = setInterval(fetchTickets, 10000);
    return () => clearInterval(interval);
  }, [fetchTickets]);

  // Update elapsed time every second
  useEffect(() => {
    const t = setInterval(() => setNow(Date.now()), 1000);
    return () => clearInterval(t);
  }, []);

  const handleToggleItem = async (itemId: string, currentStatus: string) => {
    if (!storeId) return;
    try {
      if (currentStatus === 'completed') {
        await kdsApi.markItemAsPending(storeId, itemId);
      } else {
        await kdsApi.markItemAsDone(storeId, itemId);
      }
      fetchTickets();
    } catch (err: unknown) {
      console.error(err);
      alert('Gagal mengupdate status item');
    }
  };

  const handleCompleteTicket = async (ticket: Transaction) => {
    if (!storeId) return;
    try {
      const pendingItems = ticket.items.filter(i => i.status !== 'completed');
      await Promise.all(pendingItems.map(i => kdsApi.markItemAsDone(storeId, i.id)));
      fetchTickets();
    } catch (err: unknown) {
      console.error(err);
      alert('Gagal menyelesaikan tiket pencetakan');
    }
  };

  if (!selectedStore || selectedStore.store_type !== 'restaurant') {
    return (
      <div
        style={{
          marginLeft: 0,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          height: '100vh',
        }}
      >
        <div className="empty-state">
          <AlertCircle size={40} />
          <p>Sistem KDS hanya tersedia untuk tipe toko Restoran.</p>
        </div>
      </div>
    );
  }

  return (
    <div
      style={{
        marginLeft: 0,
        padding: '24px 28px',
        minHeight: '100vh',
        background: 'var(--bg-base)',
      }}
    >
      {/* Header */}
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          marginBottom: 24,
          paddingBottom: 16,
          borderBottom: '1px solid var(--border)',
        }}
      >
        <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
          <div
            style={{
              width: 44,
              height: 44,
              borderRadius: 12,
              background: 'linear-gradient(135deg, #ef4444, #b91c1c)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              color: '#fff',
              boxShadow: '0 4px 12px rgba(239,68,68,0.3)',
            }}
          >
            <ChefHat size={22} />
          </div>
          <div>
            <h1 style={{ fontSize: '1.4rem', fontWeight: 800, margin: 0, color: 'var(--text-1)' }}>
              Kitchen Display
            </h1>
            <p style={{ color: 'var(--text-3)', fontSize: '0.85rem', margin: 0, marginTop: 2 }}>
              {selectedStore.store_name}
            </p>
          </div>
        </div>

        <button onClick={fetchTickets} className="btn btn-secondary btn-sm" disabled={loading}>
          <RefreshCw size={15} className={loading ? 'loading-spin' : ''} />
          Refresh
        </button>
      </div>

      {loading && tickets.length === 0 ? (
        <div className="empty-state">Memuat antrean...</div>
      ) : tickets.length === 0 ? (
        <div className="empty-state" style={{ marginTop: 60 }}>
          <CheckCircle2 size={48} style={{ color: 'var(--brand)', marginBottom: 16 }} />
          <h3 style={{ fontSize: '1.2rem', color: 'var(--text-1)', marginBottom: 8 }}>
            Dapur Kosong!
          </h3>
          <p>Semua pesanan telah selesai dikerjakan.</p>
        </div>
      ) : (
        <div
          style={{
            display: 'flex',
            gap: 16,
            overflowX: 'auto',
            paddingBottom: 20,
            alignItems: 'flex-start',
          }}
        >
          {tickets.map(ticket => {
            const isTakeAway = !ticket.table_id;
            const completedCount = ticket.items.filter(i => i.status === 'completed').length;
            const totalCount = ticket.items.length;
            const isAllDone = completedCount === totalCount;

            // Calculate elapsed time
            const created = new Date(ticket.created_at).getTime();
            const elapsedMins = Math.max(0, Math.floor((now - created) / 60000));
            const isLate = elapsedMins > 15;

            return (
              <div
                key={ticket.id}
                style={{
                  flex: '0 0 280px',
                  background: 'var(--bg-elevated)',
                  borderRadius: 8,
                  border: `1px solid ${isAllDone ? '#10b981' : isLate ? '#ef4444' : 'var(--border)'}`,
                  boxShadow: isLate
                    ? '0 2px 8px rgba(239,68,68,0.1)'
                    : '0 1px 4px rgba(0,0,0,0.03)',
                  overflow: 'hidden',
                  display: 'flex',
                  flexDirection: 'column',
                }}
              >
                {/* Ticket Header */}
                <div
                  style={{
                    background: isAllDone ? 'rgba(16,185,129,0.1)' : 'var(--bg-elevated)',
                    color: 'var(--text-1)',
                    padding: '12px 14px',
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'flex-start',
                    borderBottom: '1px solid var(--border)',
                  }}
                >
                  <div>
                    <div
                      style={{
                        fontSize: '0.9rem',
                        fontWeight: 600,
                        display: 'flex',
                        alignItems: 'center',
                        gap: 6,
                        color: isAllDone ? '#10b981' : 'var(--text-1)',
                      }}
                    >
                      {isTakeAway ? (
                        <ShoppingBag size={14} style={{ color: '#fb923c' }} />
                      ) : (
                        <UtensilsCrossed size={14} style={{ color: 'var(--brand)' }} />
                      )}
                      {isTakeAway
                        ? 'Take Away'
                        : `Meja ${ticket.table_number || ticket.table_id?.substring(0, 4)}`}
                    </div>
                    <div style={{ fontSize: '0.7rem', color: 'var(--text-3)', marginTop: 4 }}>
                      #{ticket.id.substring(0, 6).toUpperCase()} • {ticket.cashier_name}
                    </div>
                  </div>
                  <div
                    style={{
                      textAlign: 'right',
                      background: 'var(--bg-base)',
                      padding: '2px 8px',
                      borderRadius: 4,
                      border: '1px solid var(--border-md)',
                    }}
                  >
                    <div
                      style={{
                        fontSize: '0.85rem',
                        fontWeight: 700,
                        color: isAllDone ? '#10b981' : 'var(--text-1)',
                      }}
                    >
                      {completedCount}/{totalCount}
                    </div>
                  </div>
                </div>

                {/* Timing bar */}
                <div
                  style={{
                    padding: '6px 16px',
                    background: isLate ? '#ef4444' : 'var(--bg-base)',
                    color: isLate ? '#fff' : 'var(--text-3)',
                    fontSize: '0.7rem',
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center',
                    fontWeight: 500,
                    borderBottom: '1px solid var(--border)',
                  }}
                >
                  <span style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
                    <Clock size={12} />{' '}
                    {new Date(ticket.created_at).toLocaleTimeString('id-ID', {
                      hour: '2-digit',
                      minute: '2-digit',
                    })}
                  </span>
                  <span style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
                    {elapsedMins} mnt lalu
                  </span>
                </div>

                {/* Items */}
                <div
                  style={{
                    padding: 12,
                    display: 'flex',
                    flexDirection: 'column',
                    gap: 8,
                    maxHeight: 'calc(100vh - 280px)',
                    overflowY: 'auto',
                  }}
                >
                  {ticket.items.map(item => {
                    const isDone = item.status === 'completed';
                    return (
                      <div
                        key={item.id}
                        onClick={() => handleToggleItem(item.id, item.status)}
                        style={{
                          display: 'flex',
                          alignItems: 'flex-start',
                          gap: 12,
                          padding: '12px 14px',
                          background: isDone ? 'rgba(16,185,129,0.06)' : 'var(--bg-base)',
                          border: `1px solid ${isDone ? '#10b981' : 'transparent'}`,
                          borderBottom: isDone ? '1px solid #10b981' : '1px solid var(--border-md)',
                          borderRadius: 6,
                          cursor: 'pointer',
                          transition: 'all 0.2s',
                        }}
                      >
                        <div style={{ color: isDone ? '#10b981' : 'var(--text-3)' }}>
                          {isDone ? <CheckSquare size={16} /> : <Square size={16} />}
                        </div>
                        <div style={{ flex: 1 }}>
                          <div
                            style={{
                              fontSize: '0.85rem',
                              fontWeight: 600,
                              color: isDone ? 'var(--text-3)' : 'var(--text-1)',
                              textDecoration: isDone ? 'line-through' : 'none',
                            }}
                          >
                            {item.quantity} x {item.product_name}
                          </div>
                          {isDone && item.completed_at && (
                            <div
                              style={{ fontSize: '0.7rem', color: 'var(--text-3)', marginTop: 4 }}
                            >
                              Selesai: {new Date(item.completed_at).toLocaleTimeString('id-ID')}
                            </div>
                          )}
                        </div>
                      </div>
                    );
                  })}
                </div>

                {/* Footer Action */}
                <div
                  style={{
                    padding: 16,
                    borderTop: '1px solid var(--border-md)',
                    background: 'var(--bg-base)',
                    marginTop: 'auto',
                  }}
                >
                  <button
                    onClick={() => handleCompleteTicket(ticket)}
                    disabled={isAllDone}
                    style={{
                      width: '100%',
                      background: isAllDone ? '#10b981' : 'var(--accent-em)',
                      color: '#fff',
                      border: 'none',
                      padding: '10px 12px',
                      borderRadius: 6,
                      fontSize: '0.8rem',
                      fontWeight: 600,
                      cursor: isAllDone ? 'not-allowed' : 'pointer',
                      opacity: isAllDone ? 0.7 : 1,
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      gap: 8,
                      transition: 'all 0.2s',
                    }}
                  >
                    <CheckCircle2 size={18} />
                    {isAllDone ? 'TIKET SELESAI' : 'TANDAI SEMUA DONE'}
                  </button>
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
