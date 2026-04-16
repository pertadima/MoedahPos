import { api } from './client';
import type { StockAdjustment, CreateAdjustmentInput } from '@/types';

export const stockAdjustmentApi = {
  getHistory: (storeId: string, productId?: string) => {
    const params = new URLSearchParams();
    if (productId) params.append('product_id', productId);
    return api.get<StockAdjustment[]>(`/stores/${storeId}/adjustments?${params.toString()}`);
  },

  create: (storeId: string, data: CreateAdjustmentInput) => {
    return api.post<{ status: string }>(`/stores/${storeId}/adjustments`, data);
  },
};
