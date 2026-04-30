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
  type: 'EARN' | 'SPEND' | 'VOID' | 'ADJUST';
  balance_snapshot?: number;
  created_at: string;
}

export interface LoyaltyHistoryPage {
  data: LoyaltyLedgerEntry[];
  meta: {
    total: number;
    page: number;
    per_page: number;
    total_pages: number;
  };
}

export interface TopCustomerLoyalty {
  customer_id: string;
  customer_name: string;
  balance: number;
  tier_name?: string;
  tier_multiplier?: number;
}

export interface LoyaltyPointsSummary {
  period: 'today' | 'week' | 'month';
  earned: number;
  used: number;
  net_change: number;
}

export interface LoyaltySummary {
  top_customers: TopCustomerLoyalty[];
  periods: LoyaltyPointsSummary[];
}
