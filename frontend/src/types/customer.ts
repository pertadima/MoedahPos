import type { MembershipTier } from './loyalty';

export interface Customer {
  id: string;
  store_id: string;
  name: string;
  phone?: string;
  email?: string;
  address?: string;
  notes?: string;
  loyalty_tier_id?: string;
  loyalty_tier?: MembershipTier;
  loyalty_balance?: number;
  created_at: string;
  updated_at: string;
  server_updated_at?: string;
  sync_version?: number;
}
