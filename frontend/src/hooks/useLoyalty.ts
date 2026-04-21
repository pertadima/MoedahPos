'use client';

import { useCallback, useState } from 'react';

import type { LoyaltyBalance, LoyaltyLedgerEntry, MembershipTier } from '@/types';
import { loyaltyApi } from '@/lib/api/store-apis';

interface LoyaltyState {
  balance: LoyaltyBalance | null;
  tiers: MembershipTier[];
  loading: boolean;
  redeemLoading: boolean;
  earnLoading: boolean;
  error: string | null;
}

interface UseLoyaltyReturn extends LoyaltyState {
  /** Load the loyalty balance and tier for a customer. */
  fetchBalance: (storeId: string, customerId: string) => Promise<void>;
  /** Load all available membership tiers. */
  fetchTiers: (storeId: string) => Promise<void>;
  /** Credit loyalty points after a successful checkout. Returns the ledger entry or null on error. */
  earnPoints: (
    storeId: string,
    customerId: string,
    transactionId: string,
    total: number
  ) => Promise<LoyaltyLedgerEntry | null>;
  /** Redeem loyalty points during checkout. Returns the updated ledger entry or null on error. */
  redeemPoints: (
    storeId: string,
    customerId: string,
    points: number
  ) => Promise<LoyaltyLedgerEntry | null>;
  /** How many points the current cart total will earn (preview, not yet credited). */
  previewEarnings: (total: number, multiplier?: number) => number;
  /** Reset loyalty state (e.g. when a new customer is selected). */
  reset: () => void;
}

const POINTS_PER_UNIT = 1000;

/** Calculate points earned for a given total and multiplier (mirrors backend formula). */
function calculatePoints(total: number, multiplier = 1): number {
  if (total <= 0 || multiplier <= 0) return 0;
  return Math.floor(total / POINTS_PER_UNIT) * multiplier;
}

const initialState: LoyaltyState = {
  balance: null,
  tiers: [],
  loading: false,
  redeemLoading: false,
  earnLoading: false,
  error: null,
};

export function useLoyalty(): UseLoyaltyReturn {
  const [state, setState] = useState<LoyaltyState>(initialState);

  const fetchBalance = useCallback(async (storeId: string, customerId: string) => {
    setState(s => ({ ...s, loading: true, error: null }));
    try {
      const data = await loyaltyApi.getBalance(storeId, customerId);
      setState(s => ({ ...s, balance: data, loading: false }));
    } catch {
      setState(s => ({ ...s, loading: false, error: 'Failed to load loyalty balance' }));
    }
  }, []);

  const fetchTiers = useCallback(async (storeId: string) => {
    try {
      const data = await loyaltyApi.listTiers(storeId);
      setState(s => ({ ...s, tiers: data }));
    } catch {
      // Non-critical; tiers are informational only
    }
  }, []);

  const earnPoints = useCallback(
    async (
      storeId: string,
      customerId: string,
      transactionId: string,
      total: number
    ): Promise<LoyaltyLedgerEntry | null> => {
      setState(s => ({ ...s, earnLoading: true }));
      try {
        const entry = await loyaltyApi.earnPoints(storeId, customerId, {
          transaction_id: transactionId,
          total,
        });
        // Refresh balance after earning
        const newBalance = await loyaltyApi.getBalance(storeId, customerId);
        setState(s => ({ ...s, balance: newBalance, earnLoading: false }));
        return entry;
      } catch {
        setState(s => ({ ...s, earnLoading: false, error: 'Failed to credit loyalty points' }));
        return null;
      }
    },
    []
  );

  const redeemPoints = useCallback(
    async (
      storeId: string,
      customerId: string,
      points: number
    ): Promise<LoyaltyLedgerEntry | null> => {
      setState(s => ({ ...s, redeemLoading: true, error: null }));
      try {
        const entry = await loyaltyApi.redeemPoints(storeId, customerId, { points });
        // Optimistically update balance
        setState(s => ({
          ...s,
          redeemLoading: false,
          balance: s.balance ? { ...s.balance, balance: s.balance.balance - points } : null,
        }));
        return entry;
      } catch (err: unknown) {
        const message = err instanceof Error ? err.message : 'Failed to redeem loyalty points';
        setState(s => ({ ...s, redeemLoading: false, error: message }));
        return null;
      }
    },
    []
  );

  const previewEarnings = useCallback(
    (total: number, multiplier?: number): number => {
      const tier = state.balance?.tier;
      const effectiveMultiplier = multiplier ?? tier?.multiplier ?? 1;
      return calculatePoints(total, effectiveMultiplier);
    },
    [state.balance]
  );

  const reset = useCallback(() => {
    setState(initialState);
  }, []);

  return {
    ...state,
    fetchBalance,
    fetchTiers,
    earnPoints,
    redeemPoints,
    previewEarnings,
    reset,
  };
}
