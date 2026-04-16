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

export interface PaymentRecord {
  id: string;
  termin_id: string;
  amount_paid: number;
  payment_date: string;
  payment_method: 'cash' | 'transfer' | 'check' | 'other';
  notes: string;
  recorded_by_name: string;
  created_at: string;
}

export interface Termin {
  id: string;
  po_id: string;
  termin_number: number;
  amount: number;
  due_date: string;
  status: 'unpaid' | 'partial' | 'paid' | 'overdue';
  notes: string;
  amount_paid: number;
  amount_due: number;
  is_overdue: boolean;
  payments: PaymentRecord[];
  created_at: string;
}

export interface PODebtSummary {
  po_id: string;
  po_number: string;
  total_amount: number;
  total_termin: number;
  total_paid: number;
  remaining_debt: number;
  status: 'unpaid' | 'partial' | 'paid';
  termin_count: number;
  overdue_count: number;
}

export interface PODocumentData {
  doc_type: 'invoice' | 'receipt' | 'termin_agreement';
  generated_at: string;
  po: {
    id: string;
    po_number: string;
    supplier_name?: string;
    total_amount: number;
    status: string;
    notes: string;
    created_at: string;
  };
  debt_summary: PODebtSummary;
  termins: Termin[];
  supplier_name: string;
  store_name?: string;
}

export interface CreateTerminScheduleRequest {
  termins: {
    termin_number: number;
    amount: number;
    due_date: string;
    notes?: string;
  }[];
}

export interface RecordPaymentRequest {
  amount_paid: number;
  payment_date: string;
  payment_method: 'cash' | 'transfer' | 'check' | 'other';
  notes?: string;
}
