import { api } from './client';
import type { StockLevel, StockMovement, PaginatedData } from '@/types';

export const stockApi = {
  levels: (storeId: string, lowStockOnly = false) =>
    api.get<StockLevel[]>(`/stores/${storeId}/stock${lowStockOnly ? '?low_stock=true' : ''}`),
  movements: (storeId: string, params?: { page?: number; per_page?: number }) => {
    const q = new URLSearchParams();
    if (params?.page) q.set('page', String(params.page));
    if (params?.per_page) q.set('per_page', String(params.per_page));
    return api.get<PaginatedData<StockMovement>>(`/stores/${storeId}/stock/movements?${q}`);
  },
  adjust: (storeId: string, payload: { product_id: string; quantity_delta: number; notes: string }) =>
    api.post(`/stores/${storeId}/stock/adjust`, payload),
  setMin: (storeId: string, payload: { product_id: string; min_quantity: number }) =>
    api.put(`/stores/${storeId}/stock/min`, payload),
};

export const purchaseOrdersApi = {
  list: (storeId: string, params?: { page?: number; per_page?: number; status?: string }) => {
    const q = new URLSearchParams();
    if (params?.page) q.set('page', String(params.page));
    if (params?.per_page) q.set('per_page', String(params.per_page));
    if (params?.status) q.set('status', params.status);
    return api.get<any>(`/stores/${storeId}/purchase-orders?${q}`);
  },
  get: (storeId: string, poId: string) =>
    api.get<any>(`/stores/${storeId}/purchase-orders/${poId}`),
  create: (storeId: string, payload: object) =>
    api.post<any>(`/stores/${storeId}/purchase-orders`, payload),
  update: (storeId: string, poId: string, payload: object) =>
    api.put<any>(`/stores/${storeId}/purchase-orders/${poId}`, payload),
  submit: (storeId: string, poId: string) =>
    api.post(`/stores/${storeId}/purchase-orders/${poId}/submit`, {}),
  receive: (storeId: string, poId: string) =>
    api.post(`/stores/${storeId}/purchase-orders/${poId}/receive`, {}),
  cancel: (storeId: string, poId: string) =>
    api.delete(`/stores/${storeId}/purchase-orders/${poId}`),
};

export const suppliersApi = {
  list: (params?: { page?: number; per_page?: number; search?: string }) => {
    const q = new URLSearchParams();
    if (params?.page) q.set('page', String(params.page));
    if (params?.per_page) q.set('per_page', String(params.per_page));
    if (params?.search) q.set('search', params.search);
    return api.get<any>(`/suppliers?${q}`);
  },
  get: (id: string) => api.get<any>(`/suppliers/${id}`),
  create: (payload: object) => api.post<any>('/suppliers', payload),
  update: (id: string, payload: object) => api.put<any>(`/suppliers/${id}`, payload),
  delete: (id: string) => api.delete(`/suppliers/${id}`),
};

export const reportsApi = {
  salesSummary: (storeId: string, dateFrom?: string, dateTo?: string) => {
    const q = new URLSearchParams();
    if (dateFrom) q.set('date_from', dateFrom);
    if (dateTo) q.set('date_to', dateTo);
    return api.get<any>(`/stores/${storeId}/reports/sales?${q}`);
  },
  byProduct: (storeId: string, dateFrom?: string, dateTo?: string) => {
    const q = new URLSearchParams();
    if (dateFrom) q.set('date_from', dateFrom);
    if (dateTo) q.set('date_to', dateTo);
    return api.get<any>(`/stores/${storeId}/reports/sales/by-product?${q}`);
  },
  byCashier: (storeId: string, dateFrom?: string, dateTo?: string) => {
    const q = new URLSearchParams();
    if (dateFrom) q.set('date_from', dateFrom);
    if (dateTo) q.set('date_to', dateTo);
    return api.get<any>(`/stores/${storeId}/reports/sales/by-cashier?${q}`);
  },
  stockValuation: (storeId: string) =>
    api.get<any>(`/stores/${storeId}/reports/stock-valuation`),
};
