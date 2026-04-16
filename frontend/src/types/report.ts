export interface SalesSummaryRow {
  date: string;
  transaction_count: number;
  total_sales: number;
  total_tax: number;
  total_discount: number;
  total_net: number;
  total_cost: number;
  gross_profit: number;
  total_expense: number;
  net_profit: number;
  profit_margin?: number;
}
export interface SalesSummaryResponse {
  rows: SalesSummaryRow[];
  total_sales: number;
  total_transactions: number;
  total_cost: number;
  gross_profit: number;
  total_expense: number;
  net_profit: number;
  profit_margin: number;
}
export interface SalesByProductRow {
  product_id: string;
  product_name: string;
  sku: string;
  total_quantity: number;
  total_revenue: number;
  total_cost: number;
  gross_profit: number;
  profit_margin: number;
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

export interface ProfitPeriodRow {
  period: string;
  total_sales: number;
  total_cost: number;
  gross_profit: number;
  total_expense: number;
  net_profit: number;
  profit_margin: number;
}

export interface ProfitSummaryResponse {
  rows: ProfitPeriodRow[];
  total_sales: number;
  total_cost: number;
  gross_profit: number;
  total_expense: number;
  net_profit: number;
  profit_margin: number;
}

export interface CashFlowDetailEntry {
  type: 'SALE' | 'INCOME' | 'EXPENSE' | 'PO_PAYMENT';
  label: string;
  amount: number;
  payment_method: string;
  category?: string;
  notes?: string;
  timestamp: string;
}
