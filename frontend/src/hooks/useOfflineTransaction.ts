import { useCallback } from 'react';
import { v4 as uuidv4 } from 'uuid';
import { db, type LocalTransaction } from '../lib/dexie';
import { transactionsApi } from '../lib/api/transactions';

export const useOfflineTransaction = (storeId: string, cashierId: string) => {
  const saveTransaction = useCallback(
    async (
      transactionData: Omit<
        LocalTransaction,
        'id' | 'store_id' | 'cashier_id' | 'is_dirty' | 'created_at'
      >
    ) => {
      const txId = uuidv4();
      const now = new Date().toISOString();
      const isOnline = navigator.onLine;

      // Check if there are any pending dirty transactions in the queue.
      // While Dexie's boolean indexing varies, we fall back to a count via filter.
      let pendingCount = 0;
      try {
        const dirty = await db.transactions.filter(t => t.is_dirty === true).count();
        pendingCount = dirty;
      } catch {
        pendingCount = 0;
      }

      // ── PATH 1: Online + no pending queue items ──────────────────────────────
      // Skip Dexie entirely. Fire exactly ONE direct request to the API.
      if (isOnline && pendingCount === 0) {
        try {
          const res = await transactionsApi.syncOffline(storeId, {
            ...transactionData,
            id: txId,
            store_id: storeId,
            cashier_id: cashierId,
            created_at: now,
          });
          // Use the server-generated ID so loyalty hooks won't hit FK constraints.
          return { success: true, transactionId: res.data.id };
        } catch (err) {
          // Network failed mid-flight — fall through to save locally so it can be retried.
          console.warn('Direct checkout failed, saving locally for retry.', err);
        }
      }

      // ── PATH 2: Offline, OR online but queue has pending items ───────────────
      // Save to Dexie first to preserve ordering guarantees.
      const localTx: LocalTransaction = {
        ...transactionData,
        id: txId,
        store_id: storeId,
        cashier_id: cashierId,
        is_dirty: true,
        created_at: now,
      };

      await db.transactions.add(localTx);

      // If we're online, immediately attempt to sync this item from the queue too.
      if (isOnline) {
        try {
          const res = await transactionsApi.syncOffline(storeId, localTx);
          await db.transactions.update(txId, { is_dirty: false });
          return { success: true, transactionId: res.data.id };
        } catch (err) {
          console.warn('Queued sync attempt failed, will retry in background.', err);
        }
      }

      // Offline — leave is_dirty: true for the background SyncStatusWidget to retry.
      return { success: false, offline: true, transactionId: txId };
    },
    [storeId, cashierId]
  );

  return { saveTransaction };
};
