import { api } from './client';
import type {
  Category,
  Store,
  RestaurantTable,
  MenuItem,
  StockLevel,
  StockMovement,
  PaginatedData,
} from '@/types';

export const storesApi = {
  list: (params?: { page?: number; per_page?: number; search?: string }) => {
    const q = new URLSearchParams();
    if (params?.page) q.set('page', String(params.page));
    if (params?.per_page) q.set('per_page', String(params.per_page));
    if (params?.search) q.set('search', params.search);
    return api.get<PaginatedData<Store>>(`/stores?${q}`);
  },
  get: (id: string) => api.get<Store>(`/stores/${id}`),
  create: (payload: {
    name: string;
    address?: string;
    phone?: string;
    tax_number?: string;
    currency?: string;
    store_type: string;
    default_tax_percentage?: number;
  }) => api.post<Store>('/stores', payload),
  update: (
    id: string,
    payload: {
      name: string;
      address?: string;
      phone?: string;
      tax_number?: string;
      currency?: string;
      store_type: string;
      default_tax_percentage?: number;
      is_active?: boolean;
    }
  ) => api.put<Store>(`/stores/${id}`, payload),
  softDelete: (id: string) => api.delete(`/stores/${id}`),
  listMembers: (storeId: string) => api.get<PaginatedData<any>>(`/stores/${storeId}/members`),
};

export const categoriesApi = {
  list: (storeId: string) => api.get<Category[]>(`/stores/${storeId}/categories`),
  create: (storeId: string, payload: { name: string; parent_id?: string }) =>
    api.post<Category>(`/stores/${storeId}/categories`, payload),
  update: (storeId: string, categoryId: string, payload: { name: string; parent_id?: string }) =>
    api.put<Category>(`/stores/${storeId}/categories/${categoryId}`, payload),
  softDelete: (storeId: string, categoryId: string) =>
    api.delete(`/stores/${storeId}/categories/${categoryId}`),
};

export const tablesApi = {
  list: (storeId: string) => api.get<RestaurantTable[]>(`/stores/${storeId}/tables`),
  create: (storeId: string, payload: { table_number: string; capacity: number; notes?: string }) =>
    api.post<RestaurantTable>(`/stores/${storeId}/tables`, payload),
  update: (
    storeId: string,
    tableId: string,
    payload: { table_number: string; capacity: number; notes?: string; is_active?: boolean }
  ) => api.put<RestaurantTable>(`/stores/${storeId}/tables/${tableId}`, payload),
  updateStatus: (storeId: string, tableId: string, status: string) =>
    api.put(`/stores/${storeId}/tables/${tableId}/status`, { status }),
  delete: (storeId: string, tableId: string) => api.delete(`/stores/${storeId}/tables/${tableId}`),
};

export const menuItemsApi = {
  list: (storeId: string) => api.get<MenuItem[]>(`/stores/${storeId}/menu-items`),
  create: (storeId: string, payload: object) =>
    api.post<MenuItem>(`/stores/${storeId}/menu-items`, payload),
  update: (storeId: string, menuItemId: string, payload: object) =>
    api.put<MenuItem>(`/stores/${storeId}/menu-items/${menuItemId}`, payload),
  delete: (storeId: string, menuItemId: string) =>
    api.delete(`/stores/${storeId}/menu-items/${menuItemId}`),
};

export const stockApi = {
  levels: (storeId: string, lowStockOnly = false) =>
    api.get<StockLevel[]>(`/stores/${storeId}/stock${lowStockOnly ? '/low' : ''}`),
  movements: (storeId: string, params?: { page?: number; per_page?: number }) => {
    const q = new URLSearchParams();
    if (params?.page) q.set('page', String(params.page));
    if (params?.per_page) q.set('per_page', String(params.per_page));
    return api.get<PaginatedData<StockMovement>>(`/stores/${storeId}/stock/movements?${q}`);
  },
  adjust: (
    storeId: string,
    payload: { product_id: string; quantity_delta: number; notes: string }
  ) => api.post(`/stores/${storeId}/stock/adjust`, payload),
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
  // Accounts Payable
  payableSummary: (storeId: string) => api.get<any>(`/stores/${storeId}/purchase-orders/payables`),
  listPayments: (storeId: string, poId: string) =>
    api.get<any>(`/stores/${storeId}/purchase-orders/${poId}/payments`),
  createPayment: (storeId: string, poId: string, body: { amount: number; note?: string }) =>
    api.post<any>(`/stores/${storeId}/purchase-orders/${poId}/payments`, body),
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
  stockValuation: (storeId: string) => api.get<any>(`/stores/${storeId}/reports/stock-valuation`),
  profit: (
    storeId: string,
    dateFrom?: string,
    dateTo?: string,
    groupBy: 'day' | 'week' | 'month' = 'day'
  ) => {
    const q = new URLSearchParams({ group_by: groupBy });
    if (dateFrom) q.set('date_from', dateFrom);
    if (dateTo) q.set('date_to', dateTo);
    return api.get<any>(`/stores/${storeId}/reports/profit?${q}`);
  },
  cashFlow: (storeId: string, dateFrom?: string, dateTo?: string) => {
    const q = new URLSearchParams();
    if (dateFrom) q.set('date_from', dateFrom);
    if (dateTo) q.set('date_to', dateTo);
    return api.get<any>(`/stores/${storeId}/reports/cash-flow?${q}`);
  },
  cashFlowDetail: (storeId: string, date?: string) => {
    const q = new URLSearchParams();
    if (date) q.set('date', date);
    return api.get<any>(`/stores/${storeId}/reports/cash-flow/detail?${q}`);
  },
};

export const priceHistoryApi = {
  listByStore: (
    storeId: string,
    params?: { product_id?: string; source?: string; page?: number; per_page?: number }
  ) => {
    const q = new URLSearchParams();
    if (params?.product_id) q.set('product_id', params.product_id);
    if (params?.source) q.set('source', params.source);
    if (params?.page) q.set('page', String(params.page));
    if (params?.per_page) q.set('per_page', String(params.per_page));
    return api.get<any>(`/stores/${storeId}/price-history?${q}`);
  },
  listByProduct: (
    storeId: string,
    productId: string,
    params?: { page?: number; per_page?: number }
  ) => {
    const q = new URLSearchParams();
    if (params?.page) q.set('page', String(params.page));
    if (params?.per_page) q.set('per_page', String(params.per_page));
    return api.get<any>(`/stores/${storeId}/products/${productId}/price-history?${q}`);
  },
};

export const customersApi = {
  list: (storeId: string, params?: { page?: number; per_page?: number; search?: string }) => {
    const q = new URLSearchParams();
    if (params?.page) q.set('page', String(params.page));
    if (params?.per_page) q.set('per_page', String(params.per_page));
    if (params?.search) q.set('search', params.search);
    return api.get<any>(`/stores/${storeId}/customers?${q}`);
  },
  search: (storeId: string, query: string) =>
    api.get<any>(`/stores/${storeId}/customers/search?q=${encodeURIComponent(query)}`),
  get: (storeId: string, id: string) => api.get<any>(`/stores/${storeId}/customers/${id}`),
  create: (storeId: string, body: object) => api.post<any>(`/stores/${storeId}/customers`, body),
  update: (storeId: string, id: string, body: object) =>
    api.put<any>(`/stores/${storeId}/customers/${id}`, body),
  delete: (storeId: string, id: string) => api.delete<any>(`/stores/${storeId}/customers/${id}`),
};

export const kdsApi = {
  getTickets: (storeId: string) => api.get<any>(`/stores/${storeId}/kds/tickets`),
  markItemAsDone: (storeId: string, itemId: string) =>
    api.put(`/stores/${storeId}/kds/items/${itemId}`, { status: 'completed' }),
  markItemAsPending: (storeId: string, itemId: string) =>
    api.put(`/stores/${storeId}/kds/items/${itemId}`, { status: 'pending' }),
};

export const usersAdminApi = {
  list: (params?: {
    search?: string;
    include_inactive?: boolean;
    page?: number;
    per_page?: number;
  }) => {
    const q = new URLSearchParams();
    if (params?.search) q.set('search', params.search);
    if (params?.include_inactive) q.set('include_inactive', 'true');
    if (params?.page) q.set('page', String(params.page));
    if (params?.per_page) q.set('per_page', String(params.per_page));
    return api.get<any>(`/admin/users?${q}`);
  },
  get: (id: string) => api.get<any>(`/admin/users/${id}`),
  create: (body: object) => api.post<any>('/admin/users', body),
  update: (id: string, body: object) => api.put<any>(`/admin/users/${id}`, body),
  deactivate: (id: string) => api.post<any>(`/admin/users/${id}/deactivate`, {}),
  resetPassword: (id: string, body: { password: string }) =>
    api.post<any>(`/admin/users/${id}/reset-password`, body),
  setStores: (id: string, stores: { store_id: string; role_id: string }[]) =>
    api.put<any>(`/admin/users/${id}/stores`, { stores }),
};

export const rolesApi = {
  list: () => api.get<any>('/admin/roles'),
};

export const expensesApi = {
  listCategories: (params?: { include_deleted?: boolean }) => {
    const q = new URLSearchParams();
    if (params?.include_deleted) q.set('include_deleted', 'true');
    return api.get<any>(`/expense-categories?${q}`);
  },
  createCategory: (body: object) => api.post<any>('/expense-categories', body),
  updateCategory: (id: string, body: object) => api.put<any>(`/expense-categories/${id}`, body),
  deleteCategory: (id: string) => api.delete<any>(`/expense-categories/${id}`),
  list: (
    storeId: string,
    params?: {
      category_id?: string;
      date_from?: string;
      date_to?: string;
      page?: number;
      per_page?: number;
    }
  ) => {
    const q = new URLSearchParams();
    if (params?.category_id) q.set('category_id', params.category_id);
    if (params?.date_from) q.set('date_from', params.date_from);
    if (params?.date_to) q.set('date_to', params.date_to);
    if (params?.page) q.set('page', String(params.page));
    if (params?.per_page) q.set('per_page', String(params.per_page));
    return api.get<any>(`/stores/${storeId}/expenses?${q}`);
  },
  create: (storeId: string, body: object) => api.post<any>(`/stores/${storeId}/expenses`, body),
  update: (storeId: string, id: string, body: object) =>
    api.put<any>(`/stores/${storeId}/expenses/${id}`, body),
  updateStatus: (
    storeId: string,
    id: string,
    body: { payment_status: 'paid' | 'unpaid' | 'cancelled' }
  ) => api.patch<any>(`/stores/${storeId}/expenses/${id}/status`, body),
  delete: (storeId: string, id: string) => api.delete<any>(`/stores/${storeId}/expenses/${id}`),
};

export const recurringExpensesApi = {
  list: (storeId: string, params?: { category_id?: string; page?: number; per_page?: number }) => {
    const q = new URLSearchParams();
    if (params?.category_id) q.set('category_id', params.category_id);
    if (params?.page) q.set('page', String(params.page));
    if (params?.per_page) q.set('per_page', String(params.per_page));
    return api.get<any>(`/stores/${storeId}/recurring-expenses?${q}`);
  },
  create: (storeId: string, body: object) =>
    api.post<any>(`/stores/${storeId}/recurring-expenses`, body),
  update: (storeId: string, id: string, body: object) =>
    api.put<any>(`/stores/${storeId}/recurring-expenses/${id}`, body),
  delete: (storeId: string, id: string) =>
    api.delete<any>(`/stores/${storeId}/recurring-expenses/${id}`),
};

export const incomesApi = {
  listCategories: (params?: { include_deleted?: boolean }) => {
    const q = new URLSearchParams();
    if (params?.include_deleted) q.set('include_deleted', 'true');
    return api.get<any>(`/income-categories?${q}`);
  },
  createCategory: (body: object) => api.post<any>('/income-categories', body),
  updateCategory: (id: string, body: object) => api.put<any>(`/income-categories/${id}`, body),
  deleteCategory: (id: string) => api.delete<any>(`/income-categories/${id}`),
  list: (
    storeId: string,
    params?: {
      category_id?: string;
      date_from?: string;
      date_to?: string;
      page?: number;
      per_page?: number;
    }
  ) => {
    const q = new URLSearchParams();
    if (params?.category_id) q.set('category_id', params.category_id);
    if (params?.date_from) q.set('date_from', params.date_from);
    if (params?.date_to) q.set('date_to', params.date_to);
    if (params?.page) q.set('page', String(params.page));
    if (params?.per_page) q.set('per_page', String(params.per_page));
    return api.get<any>(`/stores/${storeId}/incomes?${q}`);
  },
  create: (storeId: string, body: object) => api.post<any>(`/stores/${storeId}/incomes`, body),
  update: (storeId: string, id: string, body: object) =>
    api.put<any>(`/stores/${storeId}/incomes/${id}`, body),
  delete: (storeId: string, id: string) => api.delete<any>(`/stores/${storeId}/incomes/${id}`),
};
