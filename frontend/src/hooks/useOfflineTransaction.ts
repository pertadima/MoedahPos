import { useCallback } from 'react';
import { v4 as uuidv4 } from 'uuid';
import { db, LocalTransaction } from '../lib/dexie';

export const useOfflineTransaction = (storeId: string, cashierId: string) => {
  const saveTransaction = useCallback(
    async (
      transactionData: Omit<
        LocalTransaction,
        'id' | 'store_id' | 'cashier_id' | 'is_dirty' | 'created_at'
      >
    ) => {
      // 1. Generate UUID for ID client-side
      const txId = uuidv4();
      const now = new Date().toISOString();

      const localTx: LocalTransaction = {
        ...transactionData,
        id: txId,
        store_id: storeId,
        cashier_id: cashierId,
        is_dirty: true, // Explicitly mark dirty to ensure background-sync if fails
        created_at: now,
      };

      // 2. Persist locally first before network
      await db.transactions.add(localTx);

      // 3. Attempt network POST
      try {
        const response = await fetch(`/api/v1/stores/${storeId}/transactions`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(localTx),
        });

        if (!response.ok) {
          throw new Error(`Failed to push transaction: ${response.statusText}`);
        }

        // 4. On success, mark clean. 
        // Our backend sync endpoint will issue the server_updated_at later.
        await db.transactions.update(txId, { is_dirty: false });

        return { success: true, transactionId: txId };
      } catch (err) {
        // Leave is_dirty: true in our local Dexie table to be retried
        console.warn('Network error during checkout, transaction saved for offline sync.', err);
        return { success: false, offline: true, transactionId: txId };
      }
    },
    [storeId, cashierId]
  );

  return { saveTransaction };
};
