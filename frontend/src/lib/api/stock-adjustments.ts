import { api } from './client';

export interface StockAdjustment {
  id: string;
  product_id: string;
  store_id: string;
  type: 'IN' | 'OUT';
  reason: 'DAMAGED' | 'LOST' | 'MANUAL_CORRECTION';
  quantity: number;
  notes: string;
  created_by: string;
  created_at: string;
  updated_at: string;

  product_name: string;
  product_sku: string;
  unit: string;
  created_by_name: string;
}

export interface CreateAdjustmentInput {
  product_id: string;
  type: 'IN' | 'OUT';
  reason: 'DAMAGED' | 'LOST' | 'MANUAL_CORRECTION';
  quantity: number;
  notes: string;
}

export const stockAdjustmentApi = {
  getHistory: (storeId: string, productId?: string) => {
    const params = new URLSearchParams();
    if (productId) params.append('product_id', productId);
    return api.get<StockAdjustment[]>(
      `/stores/${storeId}/adjustments?${params.toString()}`
    );
  },

  create: (storeId: string, data: CreateAdjustmentInput) => {
    return api.post<{ status: string }>(`/stores/${storeId}/adjustments`, data);
  },
};
