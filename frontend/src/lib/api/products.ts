import { api } from './client';
import type { Product, Category, PaginatedData } from '@/types';

export const productsApi = {
  list: (
    storeId: string,
    params?: { page?: number; per_page?: number; search?: string; category_id?: string }
  ) => {
    const q = new URLSearchParams();
    if (params?.page) q.set('page', String(params.page));
    if (params?.per_page) q.set('per_page', String(params.per_page));
    if (params?.search) q.set('search', params.search);
    if (params?.category_id) q.set('category_id', params.category_id);
    return api.get<PaginatedData<Product>>(`/stores/${storeId}/products?${q}`);
  },
  get: (storeId: string, productId: string) =>
    api.get<Product>(`/stores/${storeId}/products/${productId}`),
  create: (storeId: string, payload: Partial<Product> & { initial_qty?: number }) =>
    api.post<Product>(`/stores/${storeId}/products`, payload),
  update: (storeId: string, productId: string, payload: Partial<Product>) =>
    api.put<Product>(`/stores/${storeId}/products/${productId}`, payload),
  delete: (storeId: string, productId: string) =>
    api.delete(`/stores/${storeId}/products/${productId}`),
  byBarcode: (storeId: string, barcode: string) =>
    api.get<Product>(`/stores/${storeId}/products/barcode/${barcode}`),

  // Categories
  listCategories: (storeId: string) => api.get<Category[]>(`/stores/${storeId}/categories`),
  createCategory: (storeId: string, payload: { name: string; parent_id?: string }) =>
    api.post<Category>(`/stores/${storeId}/categories`, payload),
  updateCategory: (storeId: string, catId: string, payload: { name: string; parent_id?: string }) =>
    api.put<Category>(`/stores/${storeId}/categories/${catId}`, payload),
  deleteCategory: (storeId: string, catId: string) =>
    api.delete(`/stores/${storeId}/categories/${catId}`),
};
