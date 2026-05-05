import { api } from './client';
import type {
  ApiResponse,
  PaymentRecord,
  Termin,
  PODebtSummary,
  PODocumentData,
  CreateTerminScheduleRequest,
  RecordPaymentRequest,
} from '@/types';

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
  return (res as ApiResponse<PaymentRecord>).data ?? ({} as PaymentRecord);
}

/** GET /debt — aggregated debt summary for a PO */
export async function getPODebtSummary(storeId: string, poId: string): Promise<PODebtSummary> {
  const res = await api.get<PODebtSummary>(`${base(storeId, poId)}/debt`);
  return (res as ApiResponse<PODebtSummary>).data ?? ({} as PODebtSummary);
}

/** GET /document?type=... — data for printable document */
export async function getPODocument(
  storeId: string,
  poId: string,
  type: 'invoice' | 'receipt' | 'termin_agreement'
): Promise<PODocumentData> {
  const res = await api.get<PODocumentData>(`${base(storeId, poId)}/document?type=${type}`);
  return (res as ApiResponse<PODocumentData>).data ?? ({} as PODocumentData);
}
