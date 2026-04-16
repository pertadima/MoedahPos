export interface ExpenseCategory {
  id: string;
  name: string;
  description: string;
  created_at: string;
}

export interface Expense {
  id: string;
  store_id: string;
  category_id: string;
  category_name: string;
  amount: number;
  expense_date: string; // YYYY-MM-DD
  notes: string;
  payment_status: 'paid' | 'unpaid' | 'cancelled';
  created_by?: string;
  created_at: string;
  updated_at: string;
}

export interface RecurringExpense {
  id: string;
  store_id: string;
  category_id: string;
  category_name: string;
  name: string;
  amount: number;
  interval: 'daily' | 'weekly' | 'monthly' | 'yearly';
  interval_value: number;
  start_date: string; // YYYY-MM-DD
  end_date?: string; // YYYY-MM-DD
  next_run_date: string; // YYYY-MM-DD
  notes: string;
  is_active: boolean;
  created_by?: string;
  created_at: string;
  updated_at: string;
  last_generated_at?: string;
}

export interface IncomeCategory {
  id: string;
  name: string;
  description?: string;
  created_at: string;
}

export interface Income {
  id: string;
  store_id: string;
  category_id: string;
  category_name: string;
  amount: number;
  income_date: string; // YYYY-MM-DD
  payment_method: 'cash' | 'transfer' | 'qris' | 'other';
  reference?: string;
  notes?: string;
  created_by?: string;
  created_by_name?: string;
  created_at: string;
  updated_at: string;
}
