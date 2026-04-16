export type TableStatus = 'available' | 'occupied' | 'reserved' | 'unavailable';
export interface RestaurantTable {
  id: string;
  store_id: string;
  table_number: string;
  capacity: number;
  status: TableStatus;
  notes?: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface MenuItemIngredient {
  id: string;
  product_id: string;
  product_name: string;
  product_sku: string;
  unit: string;
  quantity: number;
  cost_price: number;
}

export interface MenuItem {
  id: string;
  store_id: string;
  category_id?: string;
  category_name?: string;
  name: string;
  description: string;
  sell_price: number;
  cost_price: number;
  ingredient_cost?: number;
  packaging_cost?: number;
  overhead_cost?: number;
  labor_cost?: number;
  use_global_tax?: boolean;
  tax_percentage?: number | null;
  tax_rate: number;
  image_url?: string;
  is_active: boolean;
  ingredients: MenuItemIngredient[];
  created_at: string;
  updated_at: string;
}
