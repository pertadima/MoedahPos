import type { ApiResponse } from '@/types';

const BASE_URL = process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost:8080/api/v1';

// ── Token Management ─────────────────────────────────────────────────────────

export function getAccessToken(): string | null {
  if (typeof window === 'undefined') return null;
  return localStorage.getItem('access_token');
}
export function setAccessToken(token: string): void {
  localStorage.setItem('access_token', token);
}
export function getRefreshToken(): string | null {
  if (typeof window === 'undefined') return null;
  return localStorage.getItem('refresh_token');
}
export function setTokens(access: string, refresh: string): void {
  localStorage.setItem('access_token', access);
  localStorage.setItem('refresh_token', refresh);
}
export function clearTokens(): void {
  localStorage.removeItem('access_token');
  localStorage.removeItem('refresh_token');
}

// ── Refresh Logic ─────────────────────────────────────────────────────────────

let isRefreshing = false;
let refreshQueue: ((token: string | null) => void)[] = [];

async function attemptRefresh(): Promise<string | null> {
  const refreshToken = getRefreshToken();
  if (!refreshToken) return null;
  try {
    const res = await fetch(`${BASE_URL}/auth/refresh`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token: refreshToken }),
    });
    if (!res.ok) {
      clearTokens();
      return null;
    }
    const json = await res.json();
    const newAccess = json.data?.access_token;
    const newRefresh = json.data?.refresh_token;
    if (newAccess) setTokens(newAccess, newRefresh ?? refreshToken);
    return newAccess ?? null;
  } catch {
    return null;
  }
}

// ── Core Fetch ────────────────────────────────────────────────────────────────

export class ApiError extends Error {
  constructor(
    public status: number,
    message: string
  ) {
    super(message);
    this.name = 'ApiError';
  }
}

async function request<T>(path: string, options: RequestInit = {}, retry = true): Promise<T> {
  const token = getAccessToken();
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(options.headers as Record<string, string>),
  };
  if (token) headers['Authorization'] = `Bearer ${token}`;

  const res = await fetch(`${BASE_URL}${path}`, { ...options, headers });

  // 401 → try token refresh once
  if (res.status === 401 && retry) {
    if (!isRefreshing) {
      isRefreshing = true;
      const newToken = await attemptRefresh();
      isRefreshing = false;
      refreshQueue.forEach(cb => cb(newToken));
      refreshQueue = [];
      if (newToken) return request<T>(path, options, false);
    } else {
      await new Promise<string | null>(resolve => refreshQueue.push(resolve));
      return request<T>(path, options, false);
    }
    clearTokens();
    window.location.href = '/login';
    throw new ApiError(401, 'Session expired');
  }

  const json = await res.json().catch(() => ({}));
  if (!res.ok) {
    const msg = json?.message ?? json?.error ?? `HTTP ${res.status}`;
    throw new ApiError(res.status, msg);
  }
  return json as T;
}

// ── Exported API Methods ──────────────────────────────────────────────────────

export const api = {
  get: <T>(path: string) => request<ApiResponse<T>>(path, { method: 'GET' }),

  post: <T>(path: string, body: unknown) =>
    request<ApiResponse<T>>(path, { method: 'POST', body: JSON.stringify(body) }),

  put: <T>(path: string, body: unknown) =>
    request<ApiResponse<T>>(path, { method: 'PUT', body: JSON.stringify(body) }),

  delete: <T>(path: string) => request<ApiResponse<T>>(path, { method: 'DELETE' }),
};
