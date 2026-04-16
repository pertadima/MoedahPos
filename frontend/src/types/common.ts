export interface ApiResponse<T> {
  success: boolean;
  data: T;
  message?: string;
}
export interface PaginatedData<T> {
  data: T[];
  meta: { page: number; per_page: number; total: number; total_pages: number };
}
