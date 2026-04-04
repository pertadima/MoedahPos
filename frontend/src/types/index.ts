// ── API envelope ─────────────────────────────────────────────────────────────
export interface ApiResponse<T> {
  success: boolean;
  data: T;
  message?: string;
}
export interface PaginatedData<T> {
  data: T[];
  meta: { page: number; per_page: number; total: number; total_pages: number };
}

// ── Auth ──────────────────────────────────────────────────────────────────────
export interface User {
  id: string;
  name: string;
  email: string;
  is_active: boolean;
  stores: UserStore[];
}
export interface UserStore {
  store_id: string;
  store_name: string;
  role: string;
  store_type?: 'retail' | 'restaurant';
}
export interface LoginResponse {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  user: User;
}

// ── Store ─────────────────────────────────────────────────────────────────────
export interface Store {
  id: string;
  name: string;
  address: string;
  phone: string;
  tax_number: string;
  currency: string;
  store_type: 'retail' | 'restaurant';
  is_active: boolean;
  created_at: string;
  updated_at: string;
  deleted_at?: string;
}
export interface StoreMember {
  user_id: string;
  user_name: string;
  user_email: string;
  role_id: string;
  role_name: string;
  is_active: boolean;
  joined_at: string;
}

// ── Category ──────────────────────────────────────────────────────────────────
export interface Category {
  id: string;
  store_id: string;
  name: string;
  parent_id?: string;
  parent_name?: string;
  created_at: string;
  updated_at: string;
}

// ── Restaurant ────────────────────────────────────────────────────────────────
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
}

export interface MenuItem {
  id: string;
  store_id: string;
  category_id?: string;
  category_name?: string;
  name: string;
  description: string;
  sell_price: number;
  tax_rate: number;
  image_url?: string;
  is_active: boolean;
  ingredients: MenuItemIngredient[];
  created_at: string;
  updated_at: string;
}

// ── Product ───────────────────────────────────────────────────────────────────
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
  tax_rate: number;
  image_url?: string;
  is_active: boolean;
  stock_qty?: number;
  created_at: string;
  updated_at: string;
}

// ── Stock ─────────────────────────────────────────────────────────────────────
export interface StockLevel {
  product_id: string;
  product_name: string;
  product_sku: string;
  unit: string;
  store_id: string;
  quantity: number;
  min_quantity: number;
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

// ── Transaction ───────────────────────────────────────────────────────────────
export interface TransactionItem {
  id: string;
  product_id?: string;
  product_name: string;
  sku: string;
  quantity: number;
  unit_price: number;
  discount_pct: number;
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
  customer_name: string;
  customer_phone: string;
  subtotal: number;
  discount_amt: number;
  tax_amt: number;
  total: number;
  payment_method: string;
  payment_amount: number;
  change_amount: number;
  status: string; // 'draft' | 'completed' | 'voided'
  notes: string;
  items: TransactionItem[];
  created_at: string;
  updated_at: string;
}

// ── Purchase Order ────────────────────────────────────────────────────────────
export interface POItem {
  id: string;
  product_id: string;
  product_name: string;
  product_sku: string;
  unit: string;
  quantity: number;
  unit_cost: number;
  received_qty: number;
  subtotal: number;
}
export interface PurchaseOrder {
  id: string;
  store_id: string;
  supplier_id?: string;
  supplier_name?: string;
  po_number: string;
  status: 'draft' | 'ordered' | 'received' | 'cancelled';
  total_amount: number;
  total_items?: number;
  ordered_by_name: string;
  received_by_name?: string;
  ordered_at?: string;
  received_at?: string;
  next_deadline?: string;
  amount_paid?: number;
  amount_due?: number;
  payment_status?: string;
  notes: string;
  items: POItem[];
  created_at: string;
  updated_at: string;
}

// ── Supplier ──────────────────────────────────────────────────────────────────
export interface Supplier {
  id: string;
  name: string;
  contact_name: string;
  phone: string;
  email: string;
  address: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

// ── Reports ───────────────────────────────────────────────────────────────────
export interface SalesSummaryRow {
  date: string;
  transaction_count: number;
  total_sales: number;
  total_tax: number;
  total_discount: number;
  total_net: number;
}
export interface SalesSummaryResponse {
  rows: SalesSummaryRow[];
  total_sales: number;
  total_transactions: number;
}
export interface SalesByProductRow {
  product_id: string;
  product_name: string;
  sku: string;
  total_quantity: number;
  total_revenue: number;
  total_tax: number;
}
export interface StockValuationResponse {
  rows: Array<{
    product_id: string;
    product_name: string;
    sku: string;
    unit: string;
    cost_price: number;
    quantity: number;
    total_value: number;
  }>;
  grand_total: number;
}

// ── Cart ──────────────────────────────────────────────────────────────────────
export interface CartItem {
  product: Product;
  quantity: number;
  discount_pct: number;
  unitPrice: number;
  subtotal: number;
  taxAmt: number;
}

// ── Customer ──────────────────────────────────────────────────────────────────
export interface Customer {
  id: string;
  store_id: string;
  name: string;
  phone?: string;
  email?: string;
  address?: string;
  notes?: string;
  created_at: string;
  updated_at: string;
}

// ── User Admin ────────────────────────────────────────────────────────────────
export interface UserStoreAssignment {
  store_id: string;
  store_name: string;
  store_type: string;
  role_id: string;
  role_name: string;
  is_active: boolean;
}

export interface UserAdmin {
  id: string;
  name: string;
  email: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
  deleted_at?: string;
  store_count: number;
  stores?: UserStoreAssignment[];
}

// ── Role ──────────────────────────────────────────────────────────────────────
export interface Role {
  id: string;
  name: string;
  description: string;
  permissions: string[];
}
