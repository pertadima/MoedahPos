export interface StockLevel {
  product_id: string;
  product_name: string;
  product_sku: string;
  unit: string;
  store_id: string;
  quantity: number;
  min_quantity: number;
  cost_price: number;
  is_low_stock: boolean;
  updated_at: string;
}
export interface StockMovement {
  id: string;
  product_id: string;
  product_name: string;
  store_id: string;
  ref_type: string;
  ref_id?: string;
  quantity_delta: number;
  notes: string;
  created_by: string;
  created_by_name: string;
  created_at: string;
}
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

export interface StockBatch {
  id: string;
  product_id: string;
  product_name: string;
  product_sku: string;
  unit: string;
  store_id: string;
  po_id?: string;
  quantity_remaining: number;
  purchase_price: number;
  received_at: string;
  created_at: string;
}

export interface BatchStockSummary {
  product_id: string;
  product_name: string;
  product_sku: string;
  unit: string;
  total_qty: number;
  batch_count: number;
  avg_cost_price: number;
}
