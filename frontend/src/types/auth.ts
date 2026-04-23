export interface User {
  id: string;
  name: string;
  email: string;
  is_active: boolean;
  stores: UserStore[];
}
export interface UserStore {
  store_id: string;
  store_name: string;
  role: string;
  store_type?: 'retail' | 'restaurant';
  loyalty_points_per_rupiah?: number;
  loyalty_rupiah_per_point?: number;
}
export interface LoginResponse {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  user: User;
}
export interface UserStoreAssignment {
  store_id: string;
  store_name: string;
  store_type: string;
  role_id: string;
  role_name: string;
  is_active: boolean;
}
export interface UserAdmin {
  id: string;
  name: string;
  email: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
  deleted_at?: string;
  store_count: number;
  stores?: UserStoreAssignment[];
}
export interface Role {
  id: string;
  name: string;
  description: string;
  permissions: string[];
}
