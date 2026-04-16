import { api } from './client';
import type { PaginatedData, ActivityLog, ActivityLogFilter } from '@/types';

export const activityLogsApi = {
  list: (storeId: string, params: ActivityLogFilter) => {
    const searchParams = new URLSearchParams();
    Object.entries(params).forEach(([key, value]) => {
      if (value !== undefined && value !== null) {
        searchParams.append(key, value.toString());
      }
    });
    const queryString = searchParams.toString();
    const url = `/stores/${storeId}/activity-logs${queryString ? `?${queryString}` : ''}`;
    return api.get<PaginatedData<ActivityLog>>(url);
  },
};
