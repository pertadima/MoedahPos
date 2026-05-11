import { api } from './client';
import type { LoginResponse, User, ApiResponse } from '@/types';

interface RefreshResponse {
  access_token: string;
  expires_in: number;
}

export const authApi = {
  login: (email: string, password: string) =>
    api.post<LoginResponse>('/auth/login', { email, password }),

  register: (payload: { name: string; email: string; password: string }) =>
    api.post('/auth/register', payload),

  refresh: (refreshToken: string) =>
    api.post<ApiResponse<RefreshResponse>>('/auth/refresh', { refresh_token: refreshToken }),

  logout: (refreshToken: string) => api.post('/auth/logout', { refresh_token: refreshToken }),

  me: () => api.get<User>('/auth/me'),
};
