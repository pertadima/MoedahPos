import type { Product } from './product';

export type DiscountType = 'PERCENTAGE' | 'FIXED' | 'OVERRIDE';

export interface TransactionItem {
  id: string;
  product_id?: string;
  menu_item_id?: string;
  product_name: string;
  sku: string;
  quantity: number;
  original_price: number;
  unit_price: number;
  discount_pct: number;
  discount_type: DiscountType;
  discount_value: number;
  cart_discount_allocated: number;
  tax_rate: number;
  subtotal: number;
  status: 'pending' | 'completed';
  completed_at?: string;
}

export interface Transaction {
  id: string;
  store_id: string;
  cashier_id: string;
  cashier_name: string;
  table_id?: string;
  table_number?: string;
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
  points_redeemed?: number;
  points_discount?: number;
  items: TransactionItem[];
  created_at: string;
  updated_at: string;
}

export interface CartItem {
  product: Product;
  quantity: number;
  discount_pct: number;
  discountType: DiscountType;
  discountValue: number;
  originalPrice: number;
  unitPrice: number;
  subtotal: number;
  taxAmt: number;
}

export type TxItemInput = {
  product_id?: string;
  menu_item_id?: string;
  quantity: number;
  discount_pct: number;
  discount_type?: DiscountType;
  discount_value?: number;
};
