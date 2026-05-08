import { api, getAccessToken } from './client';
import type {
  Category,
  Store,
  RestaurantTable,
  MenuItem,
  StockLevel,
  StockMovement,
  PaginatedData,
  MembershipTier,
  LoyaltyBalance,
  LoyaltyLedgerEntry,
  LoyaltyHistoryPage,
  LoyaltySummary,
  PurchaseOrder,
  CreatePORequest,
  RecordPaymentRequest,
  POPayment,
  PayableSummary,
  POListParams,
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
    loyalty_points_per_rupiah?: number;
    loyalty_rupiah_per_point?: number;
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
      loyalty_points_per_rupiah?: number;
      loyalty_rupiah_per_point?: number;
      is_active?: boolean;
    }
  ) => api.put<Store>(`/stores/${id}`, payload),
  softDelete: (id: string) => api.delete(`/stores/${id}`),
  listMembers: (storeId: string) => api.get<PaginatedData<unknown>>(`/stores/${storeId}/members`),
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
  list: (storeId: string, params?: POListParams) => {
    const sp = new URLSearchParams();
    if (params) {
      if (params.status) sp.set('status', params.status);
      if (params.per_page) sp.set('per_page', String(params.per_page));
      if (params.page) sp.set('page', String(params.page));
      if (params.supplier_id) sp.set('supplier_id', params.supplier_id);
    }
    const qs = sp.toString();
    return api.get<PaginatedData<PurchaseOrder>>(
      `/stores/${storeId}/purchase-orders${qs ? `?${qs}` : ''}`
    );
  },
  create: (storeId: string, body: CreatePORequest) =>
    api.post<PurchaseOrder>(`/stores/${storeId}/purchase-orders`, body),
  get: (storeId: string, id: string) =>
    api.get<PurchaseOrder>(`/stores/${storeId}/purchase-orders/${id}`),
  submit: (storeId: string, id: string) =>
    api.post(`/stores/${storeId}/purchase-orders/${id}/submit`, {}),
  receive: (storeId: string, id: string) =>
    api.post(`/stores/${storeId}/purchase-orders/${id}/receive`, {}),
  cancel: (storeId: string, id: string) =>
    api.post(`/stores/${storeId}/purchase-orders/${id}/cancel`, {}),
  createPayment: (storeId: string, poId: string, body: RecordPaymentRequest) =>
    api.post(`/stores/${storeId}/purchase-orders/${poId}/payments`, body),
  listPayments: (storeId: string, poId: string) =>
    api.get<POPayment[]>(`/stores/${storeId}/purchase-orders/${poId}/payments`),
  payableSummary: (storeId: string) =>
    api.get<PayableSummary>(`/stores/${storeId}/purchase-orders/payables`),
  logDocumentGenerate: (storeId: string, poId: string, docType: string) =>
    api.post(`/stores/${storeId}/purchase-orders/${poId}/document-log`, { document_type: docType }),
};

export const suppliersApi = {
  list: (params?: { page?: number; per_page?: number; search?: string }) => {
    const q = new URLSearchParams();
    if (params?.page) q.set('page', String(params.page));
    if (params?.per_page) q.set('per_page', String(params.per_page));
    if (params?.search) q.set('search', params.search);
    return api.get<unknown>(`/suppliers?${q}`);
  },
  get: (id: string) => api.get<unknown>(`/suppliers/${id}`),
  create: (payload: object) => api.post<unknown>('/suppliers', payload),
  update: (id: string, payload: object) => api.put<unknown>(`/suppliers/${id}`, payload),
  delete: (id: string) => api.delete(`/suppliers/${id}`),
};

export const reportsApi = {
  salesSummary: (storeId: string, dateFrom?: string, dateTo?: string) => {
    const q = new URLSearchParams();
    if (dateFrom) q.set('date_from', dateFrom);
    if (dateTo) q.set('date_to', dateTo);
    return api.get<unknown>(`/stores/${storeId}/reports/sales?${q}`);
  },
  byProduct: (storeId: string, dateFrom?: string, dateTo?: string) => {
    const q = new URLSearchParams();
    if (dateFrom) q.set('date_from', dateFrom);
    if (dateTo) q.set('date_to', dateTo);
    return api.get<unknown>(`/stores/${storeId}/reports/sales/by-product?${q}`);
  },
  byCashier: (storeId: string, dateFrom?: string, dateTo?: string) => {
    const q = new URLSearchParams();
    if (dateFrom) q.set('date_from', dateFrom);
    if (dateTo) q.set('date_to', dateTo);
    return api.get<unknown>(`/stores/${storeId}/reports/sales/by-cashier?${q}`);
  },
  stockValuation: (storeId: string) =>
    api.get<unknown>(`/stores/${storeId}/reports/stock-valuation`),
  profit: (
    storeId: string,
    dateFrom?: string,
    dateTo?: string,
    groupBy: 'day' | 'week' | 'month' = 'day'
  ) => {
    const q = new URLSearchParams({ group_by: groupBy });
    if (dateFrom) q.set('date_from', dateFrom);
    if (dateTo) q.set('date_to', dateTo);
    return api.get<unknown>(`/stores/${storeId}/reports/profit?${q}`);
  },
  cashFlow: (storeId: string, dateFrom?: string, dateTo?: string) => {
    const q = new URLSearchParams();
    if (dateFrom) q.set('date_from', dateFrom);
    if (dateTo) q.set('date_to', dateTo);
    return api.get<unknown>(`/stores/${storeId}/reports/cash-flow?${q}`);
  },
  cashFlowDetail: (storeId: string, date?: string) => {
    const q = new URLSearchParams();
    if (date) q.set('date', date);
    return api.get<unknown>(`/stores/${storeId}/reports/cash-flow/detail?${q}`);
  },

  /**
   * Download a CSV or printable-HTML export for a given report type.
   * Triggers a browser file download.
   *
   * @param storeId  - the store UUID
   * @param type     - "csv" | "pdf"
   * @param report   - "sales" | "inventory" | "profit"
   * @param dateFrom - optional YYYY-MM-DD
   * @param dateTo   - optional YYYY-MM-DD
   */
  exportReport: async (
    storeId: string,
    type: 'csv' | 'pdf',
    report: 'sales' | 'inventory' | 'profit',
    dateFrom?: string,
    dateTo?: string
  ): Promise<void> => {
    const BASE_URL = process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost:8080/api/v1';
    const q = new URLSearchParams({ type, report });
    if (dateFrom) q.set('date_from', dateFrom);
    if (dateTo) q.set('date_to', dateTo);

    const token = getAccessToken();
    const res = await fetch(`${BASE_URL}/stores/${storeId}/reports/export?${q}`, {
      method: 'GET',
      headers: token ? { Authorization: `Bearer ${token}` } : {},
    });

    if (!res.ok) {
      const json = await res.json().catch(() => ({}));
      throw new Error(json?.error ?? `Export failed: HTTP ${res.status}`);
    }

    const rawBlob = await res.blob();
    const blobType = type === 'pdf' ? 'application/pdf' : 'text/csv;charset=utf-8;';
    const blob = new Blob([rawBlob], { type: blobType });

    // Trigger standard file download for both CSV and PDF
    const disposition = res.headers.get('Content-Disposition') ?? '';
    const match = /filename="([^"]+)"/.exec(disposition);
    const filename = match?.[1] ?? `laporan-${report}.${type === 'csv' ? 'csv' : 'pdf'}`;

    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    a.remove();
    URL.revokeObjectURL(url);
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
    return api.get<unknown>(`/stores/${storeId}/price-history?${q}`);
  },
  listByProduct: (
    storeId: string,
    productId: string,
    params?: { page?: number; per_page?: number }
  ) => {
    const q = new URLSearchParams();
    if (params?.page) q.set('page', String(params.page));
    if (params?.per_page) q.set('per_page', String(params.per_page));
    return api.get<unknown>(`/stores/${storeId}/products/${productId}/price-history?${q}`);
  },
};

export const customersApi = {
  list: (storeId: string, params?: { page?: number; per_page?: number; search?: string }) => {
    const q = new URLSearchParams();
    if (params?.page) q.set('page', String(params.page));
    if (params?.per_page) q.set('per_page', String(params.per_page));
    if (params?.search) q.set('search', params.search);
    return api.get<unknown>(`/stores/${storeId}/customers?${q}`);
  },
  search: (storeId: string, query: string) =>
    api.get<unknown>(`/stores/${storeId}/customers/search?q=${encodeURIComponent(query)}`),
  get: (storeId: string, id: string) => api.get<unknown>(`/stores/${storeId}/customers/${id}`),
  create: (storeId: string, body: object) =>
    api.post<unknown>(`/stores/${storeId}/customers`, body),
  update: (storeId: string, id: string, body: object) =>
    api.put<unknown>(`/stores/${storeId}/customers/${id}`, body),
  delete: (storeId: string, id: string) =>
    api.delete<unknown>(`/stores/${storeId}/customers/${id}`),
};

export const kdsApi = {
  getTickets: (storeId: string) => api.get<unknown>(`/stores/${storeId}/kds/tickets`),
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
    return api.get<unknown>(`/admin/users?${q}`);
  },
  get: (id: string) => api.get<unknown>(`/admin/users/${id}`),
  create: (body: object) => api.post<unknown>('/admin/users', body),
  update: (id: string, body: object) => api.put<unknown>(`/admin/users/${id}`, body),
  deactivate: (id: string) => api.post<unknown>(`/admin/users/${id}/deactivate`, {}),
  resetPassword: (id: string, body: { password: string }) =>
    api.post<unknown>(`/admin/users/${id}/reset-password`, body),
  setStores: (id: string, stores: { store_id: string; role_id: string }[]) =>
    api.put<unknown>(`/admin/users/${id}/stores`, { stores }),
};

export const rolesApi = {
  list: () => api.get<unknown>('/admin/roles'),
  get: (id: string) => api.get<unknown>(`/admin/roles/${id}`),
  create: (body: unknown) => api.post<unknown>('/admin/roles', body),
  update: (id: string, body: unknown) => api.put<unknown>(`/admin/roles/${id}`, body),
  delete: (id: string) => api.delete<unknown>(`/admin/roles/${id}`),
  listPermissions: () => api.get<unknown>('/admin/permissions'),
};

export const expensesApi = {
  listCategories: (params?: { include_deleted?: boolean }) => {
    const q = new URLSearchParams();
    if (params?.include_deleted) q.set('include_deleted', 'true');
    return api.get<unknown>(`/expense-categories?${q}`);
  },
  createCategory: (body: object) => api.post<unknown>('/expense-categories', body),
  updateCategory: (id: string, body: object) => api.put<unknown>(`/expense-categories/${id}`, body),
  deleteCategory: (id: string) => api.delete<unknown>(`/expense-categories/${id}`),
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
    return api.get<unknown>(`/stores/${storeId}/expenses?${q}`);
  },
  create: (storeId: string, body: object) => api.post<unknown>(`/stores/${storeId}/expenses`, body),
  update: (storeId: string, id: string, body: object) =>
    api.put<unknown>(`/stores/${storeId}/expenses/${id}`, body),
  updateStatus: (
    storeId: string,
    id: string,
    body: { payment_status: 'paid' | 'unpaid' | 'cancelled' }
  ) => api.patch<unknown>(`/stores/${storeId}/expenses/${id}/status`, body),
  delete: (storeId: string, id: string) => api.delete<unknown>(`/stores/${storeId}/expenses/${id}`),
};

export const recurringExpensesApi = {
  list: (storeId: string, params?: { category_id?: string; page?: number; per_page?: number }) => {
    const q = new URLSearchParams();
    if (params?.category_id) q.set('category_id', params.category_id);
    if (params?.page) q.set('page', String(params.page));
    if (params?.per_page) q.set('per_page', String(params.per_page));
    return api.get<unknown>(`/stores/${storeId}/recurring-expenses?${q}`);
  },
  create: (storeId: string, body: object) =>
    api.post<unknown>(`/stores/${storeId}/recurring-expenses`, body),
  update: (storeId: string, id: string, body: object) =>
    api.put<unknown>(`/stores/${storeId}/recurring-expenses/${id}`, body),
  delete: (storeId: string, id: string) =>
    api.delete<unknown>(`/stores/${storeId}/recurring-expenses/${id}`),
};

export const incomesApi = {
  listCategories: (params?: { include_deleted?: boolean }) => {
    const q = new URLSearchParams();
    if (params?.include_deleted) q.set('include_deleted', 'true');
    return api.get<unknown>(`/income-categories?${q}`);
  },
  createCategory: (body: object) => api.post<unknown>('/income-categories', body),
  updateCategory: (id: string, body: object) => api.put<unknown>(`/income-categories/${id}`, body),
  deleteCategory: (id: string) => api.delete<unknown>(`/income-categories/${id}`),
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
    return api.get<unknown>(`/stores/${storeId}/incomes?${q}`);
  },
  create: (storeId: string, body: object) => api.post<unknown>(`/stores/${storeId}/incomes`, body),
  update: (storeId: string, id: string, body: object) =>
    api.put<unknown>(`/stores/${storeId}/incomes/${id}`, body),
  delete: (storeId: string, id: string) => api.delete<unknown>(`/stores/${storeId}/incomes/${id}`),
};

export const loyaltyApi = {
  /** List all membership tiers for a store. */
  listTiers: async (storeId: string): Promise<MembershipTier[]> => {
    const res = await api.get<MembershipTier[]>(`/stores/${storeId}/loyalty/tiers`);
    return res.data;
  },

  /** Get a customer's current point balance and tier. */
  getBalance: async (storeId: string, customerId: string): Promise<LoyaltyBalance> => {
    const res = await api.get<LoyaltyBalance>(`/stores/${storeId}/customers/${customerId}/loyalty`);
    return res.data;
  },

  /** Credit points after a completed transaction. */
  earnPoints: async (
    storeId: string,
    customerId: string,
    body: { transaction_id: string; total: number }
  ): Promise<LoyaltyLedgerEntry> => {
    const res = await api.post<LoyaltyLedgerEntry>(
      `/stores/${storeId}/customers/${customerId}/loyalty/earn`,
      body
    );
    return res.data;
  },

  /** Redeem points during checkout. */
  redeemPoints: async (
    storeId: string,
    customerId: string,
    body: { points: number }
  ): Promise<LoyaltyLedgerEntry> => {
    const res = await api.post<LoyaltyLedgerEntry>(
      `/stores/${storeId}/customers/${customerId}/loyalty/redeem`,
      body
    );
    return res.data;
  },

  /** Get the full point ledger history for a customer. */
  getHistory: async (storeId: string, customerId: string): Promise<LoyaltyLedgerEntry[]> => {
    const res = await api.get<LoyaltyLedgerEntry[]>(
      `/stores/${storeId}/customers/${customerId}/loyalty/history`
    );
    return res.data;
  },

  /** Get paginated point ledger history for a customer. */
  getHistoryPaginated: async (
    storeId: string,
    customerId: string,
    page = 1,
    perPage = 20
  ): Promise<LoyaltyHistoryPage> => {
    const res = await api.get<LoyaltyHistoryPage>(
      `/stores/${storeId}/customers/${customerId}/loyalty/history/paged?page=${page}&per_page=${perPage}`
    );
    return res.data;
  },

  /** Get loyalty dashboard summary: top customers + points earned/used per period. */
  getSummary: async (storeId: string): Promise<LoyaltySummary> => {
    const res = await api.get<LoyaltySummary>(`/stores/${storeId}/loyalty/summary`);
    return res.data;
  },

  /** Assign a tier to a customer. */
  assignTier: async (
    storeId: string,
    customerId: string,
    body: { tier_id: string }
  ): Promise<{ status: string }> => {
    const res = await api.put<{ status: string }>(
      `/stores/${storeId}/customers/${customerId}/loyalty/tier`,
      body
    );
    return res.data;
  },
};
