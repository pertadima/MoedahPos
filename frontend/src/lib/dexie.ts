import Dexie, { type EntityTable } from 'dexie';

export interface LocalCategory {
  id: string;
  store_id: string;
  name: string;
  parent_id: string | null;
  server_updated_at: string;
  sync_version: number;
}

export interface LocalProduct {
  id: string;
  store_id: string;
  category_id: string | null;
  sku: string;
  name: string;
  description: string | null;
  barcode: string | null;
  unit: string;
  cost_price: number;
  sell_price: number;
  tax_percentage: number | null;
  use_global_tax: boolean;
  image_url: string | null;
  is_active: boolean;
  server_updated_at: string;
  sync_version: number;
}

export interface LocalTransactionItem {
  product_id: string | null;
  menu_item_id: string | null;
  product_name: string;
  sku: string;
  quantity: number;
  original_price: number;
  unit_price: number;
  cost_price: number;
  discount_pct: number;
  discount_type: string;
  discount_value: number;
  cart_discount_allocated: number;
  tax_rate: number;
  subtotal: number;
  status: string;
}

export interface LocalTransaction {
  id: string;
  store_id: string;
  cashier_id: string;
  table_id: string | null;
  customer_name: string;
  customer_phone: string;
  subtotal: number;
  discount_amt: number;
  tax_amt: number;
  total: number;
  payment_method: string;
  payment_amount: number;
  change_amount: number;
  status: string;
  notes: string;
  cart_discount_type: string;
  cart_discount_value: number;
  items: LocalTransactionItem[];
  is_dirty: boolean;
  created_at: string;
}

const db = new Dexie('MoedahPOSDatabase') as Dexie & {
  categories: EntityTable<LocalCategory, 'id'>;
  products: EntityTable<LocalProduct, 'id'>;
  transactions: EntityTable<LocalTransaction, 'id'>;
};

// Database schema configuration
db.version(1).stores({
  categories: 'id, store_id, name, parent_id, server_updated_at',
  products: 'id, store_id, category_id, sku, barcode, name, server_updated_at',
  transactions: 'id, store_id, cashier_id, status, is_dirty, created_at',
});

export { db };
