import { api } from './client';
import type { LoginResponse, User } from '@/types';

export const authApi = {
  login: (email: string, password: string) =>
    api.post<LoginResponse>('/auth/login', { email, password }),

  register: (payload: { name: string; email: string; password: string }) =>
    api.post('/auth/register', payload),

  refresh: (refreshToken: string) => api.post('/auth/refresh', { refresh_token: refreshToken }),

  logout: (refreshToken: string) => api.post('/auth/logout', { refresh_token: refreshToken }),

  me: () => api.get<User>('/auth/me'),
};
