export interface Store {
  id: string;
  name: string;
  address: string;
  phone: string;
  tax_number: string;
  currency: string;
  store_type: 'retail' | 'restaurant';
  default_tax_percentage: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
  deleted_at?: string;
}
export interface StoreMember {
  user_id: string;
  user_name: string;
  user_email: string;
  role_id: string;
  role_name: string;
  is_active: boolean;
  joined_at: string;
}
