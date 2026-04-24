export interface MembershipTier {
  id: string;
  name: string;
  multiplier: number;
}

export interface LoyaltyBalance {
  customer_id: string;
  balance: number;
  tier?: MembershipTier;
}

export interface LoyaltyLedgerEntry {
  id: string;
  customer_id: string;
  points_delta: number;
  transaction_id?: string;
  type: 'EARN' | 'SPEND';
  created_at: string;
}
