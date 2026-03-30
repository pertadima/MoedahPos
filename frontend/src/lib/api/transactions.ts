import { api } from './client';
import type { Transaction, PaginatedData } from '@/types';

export const transactionsApi = {
  list: (storeId: string, params?: { page?: number; per_page?: number; status?: string; date_from?: string; date_to?: string }) => {
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
  checkout: (storeId: string, payload: {
    customer_name?: string;
    customer_phone?: string;
    payment_method: string;
    payment_amount: number;
    notes?: string;
    items: { product_id: string; quantity: number; discount_pct: number }[];
  }) => api.post<Transaction>(`/stores/${storeId}/transactions`, payload),
  void: (storeId: string, txnId: string) =>
    api.post(`/stores/${storeId}/transactions/${txnId}/void`, {}),
};
