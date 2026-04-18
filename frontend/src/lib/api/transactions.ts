import { api } from './client';
import type { Transaction, PaginatedData, TxItemInput } from '@/types';

export const transactionsApi = {
  // ── Existing (retail + restaurant direct checkout) ─────────────────────────
  list: (
    storeId: string,
    params?: {
      page?: number;
      per_page?: number;
      status?: string;
      date_from?: string;
      date_to?: string;
    }
  ) => {
    const q = new URLSearchParams();
    if (params?.page) q.set('page', String(params.page));
    if (params?.per_page) q.set('per_page', String(params.per_page));
    if (params?.status) q.set('status', params.status);
    if (params?.date_from) q.set('date_from', params.date_from);
    if (params?.date_to) q.set('date_to', params.date_to);
    return api.get<PaginatedData<Transaction>>(`/stores/${storeId}/transactions?${q}`);
  },

  get: (storeId: string, txnId: string) =>
    api.get<Transaction>(`/stores/${storeId}/transactions/${txnId}`),

  checkout: (
    storeId: string,
    payload: {
      customer_name?: string;
      customer_phone?: string;
      payment_method: string;
      payment_amount: number;
      notes?: string;
      items: TxItemInput[];
      cart_discount_type?: 'PERCENTAGE' | 'FIXED';
      cart_discount_value?: number;
    }
  ) => api.post<Transaction>(`/stores/${storeId}/transactions`, payload),

  void: (storeId: string, txnId: string) =>
    api.post(`/stores/${storeId}/transactions/${txnId}/void`, {}),

  // ── Offline Sync ────────────────────────────────────────────────────────────
  syncOffline: (storeId: string, payload: any) =>
    api.post<Transaction>(`/stores/${storeId}/transactions`, payload),

  // ── Draft / Table Order (restaurant) ──────────────────────────────────────

  /** Get the open draft order for a given table (returns null if none). */
  getDraftByTable: (storeId: string, tableId: string) =>
    api.get<Transaction | null>(`/stores/${storeId}/transactions/draft?table_id=${tableId}`),

  /** Create a draft order for a table (hold without payment). */
  createDraft: (
    storeId: string,
    payload: {
      table_id: string;
      customer_name?: string;
      notes?: string;
      items: TxItemInput[];
      cart_discount_type?: 'PERCENTAGE' | 'FIXED';
      cart_discount_value?: number;
    }
  ) => api.post<Transaction>(`/stores/${storeId}/transactions/draft`, payload),

  /** Replace items on an existing draft. */
  updateDraft: (
    storeId: string,
    txnId: string,
    payload: {
      customer_name?: string;
      notes?: string;
      items: TxItemInput[];
      cart_discount_type?: 'PERCENTAGE' | 'FIXED';
      cart_discount_value?: number;
    }
  ) => api.put<Transaction>(`/stores/${storeId}/transactions/${txnId}/draft`, payload),

  /** Pay a held draft order. */
  payDraft: (
    storeId: string,
    txnId: string,
    payload: {
      payment_method: string;
      payment_amount: number;
      customer_name?: string;
      customer_phone?: string;
    }
  ) => api.post<Transaction>(`/stores/${storeId}/transactions/${txnId}/pay`, payload),
};
