import { api } from './client';
import type { StockBatch, BatchStockSummary } from '@/types';

// ─── API Calls ────────────────────────────────────────────────────────────────

/**
 * GET /stores/:storeId/stock/batches
 * Returns all non-empty FIFO batches for a store, ordered product name then received_at ASC.
 * @param productId  Optional — filter to a single product.
 */
export async function listBatches(storeId: string, productId?: string): Promise<StockBatch[]> {
  const params = new URLSearchParams();
  if (productId) params.set('product_id', productId);
  const qs = params.toString() ? `?${params.toString()}` : '';
  const res = await api.get<StockBatch[]>(`/stores/${storeId}/stock/batches${qs}`);
  return res.data ?? [];
}

/**
 * GET /stores/:storeId/stock/batch-summary
 * Returns per-product totals (total_qty, batch_count, avg_cost_price).
 */
export async function getBatchSummary(storeId: string): Promise<BatchStockSummary[]> {
  const res = await api.get<BatchStockSummary[]>(`/stores/${storeId}/stock/batch-summary`);
  return res.data ?? [];
}
