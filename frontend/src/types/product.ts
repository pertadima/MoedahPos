export interface Category {
  id: string;
  store_id: string;
  name: string;
  parent_id?: string;
  parent_name?: string;
  created_at: string;
  updated_at: string;
}

export interface Product {
  id: string;
  store_id: string;
  category_id?: string;
  category_name?: string;
  sku: string;
  name: string;
  description: string;
  barcode?: string;
  unit: string;
  cost_price: number;
  sell_price: number;
  use_global_tax?: boolean;
  tax_percentage?: number | null;
  tax_rate: number;
  image_url?: string;
  is_active: boolean;
  stock_qty?: number;
  created_at: string;
  updated_at: string;
}
