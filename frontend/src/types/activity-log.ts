export interface ActivityLog {
  id: string;
  user_id: string;
  user_name: string;
  store_id: string | null;
  action_type: string;
  module: string;
  reference_id: string | null;
  metadata: Record<string, unknown> | null;
  created_at: string;
}

export interface ActivityLogFilter {
  page?: number;
  per_page?: number;
  user_id?: string;
  module?: string;
  action_type?: string;
  start_date?: string;
  end_date?: string;
}
