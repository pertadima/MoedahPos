import { api } from './client';
import type { ApiResponse } from '@/types';

// ─── Types ────────────────────────────────────────────────────────────────────

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
    due_date: string; // YYYY-MM-DD
    notes?: string;
  }[];
}

export interface RecordPaymentRequest {
  amount_paid: number;
  payment_date: string; // YYYY-MM-DD
  payment_method: 'cash' | 'transfer' | 'check' | 'other';
  notes?: string;
}

// ─── API Calls ────────────────────────────────────────────────────────────────

const base = (storeId: string, poId: string) => `/stores/${storeId}/purchase-orders/${poId}`;

/** GET /termins — list installment schedule with payment history */
export async function listTermins(storeId: string, poId: string): Promise<Termin[]> {
  const res = await api.get<Termin[]>(`${base(storeId, poId)}/termins`);
  return (res as ApiResponse<Termin[]>).data ?? [];
}

/** POST /termins — create/replace installment schedule */
export async function createTerminSchedule(
  storeId: string,
  poId: string,
  data: CreateTerminScheduleRequest
): Promise<Termin[]> {
  const res = await api.post<Termin[]>(`${base(storeId, poId)}/termins`, data);
  return (res as ApiResponse<Termin[]>).data ?? [];
}

/** POST /termins/:terminId/payments — record a payment against a termin */
export async function recordPayment(
  storeId: string,
  poId: string,
  terminId: string,
  data: RecordPaymentRequest
): Promise<PaymentRecord> {
  const res = await api.post<PaymentRecord>(
    `${base(storeId, poId)}/termins/${terminId}/payments`,
    data
  );
  return (res as ApiResponse<PaymentRecord>).data!;
}

/** GET /debt — aggregated debt summary for a PO */
export async function getPODebtSummary(storeId: string, poId: string): Promise<PODebtSummary> {
  const res = await api.get<PODebtSummary>(`${base(storeId, poId)}/debt`);
  return (res as ApiResponse<PODebtSummary>).data!;
}

/** GET /document?type=... — data for printable document */
export async function getPODocument(
  storeId: string,
  poId: string,
  type: 'invoice' | 'receipt' | 'termin_agreement'
): Promise<PODocumentData> {
  const res = await api.get<PODocumentData>(`${base(storeId, poId)}/document?type=${type}`);
  return (res as ApiResponse<PODocumentData>).data!;
}
